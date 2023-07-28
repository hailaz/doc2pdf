package doc2pdf

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"path"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/devices"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"github.com/go-rod/rod/lib/utils"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
)

// DocDownload description
type DocDownload struct {
	MainURL          string // 文档入口地址
	OutputDir        string // 输出目录
	MergePDFNums     int    // 每次合并的pdf数量，多文档时能减轻内存压力
	TempSuffix       string // 临时文件后缀
	IsDownloadMain   bool
	fileList         []string
	bookmark         []pdfcpu.Bookmark
	pageFrom         int
	baseURL          string
	browser          *rod.Browser
	OpDelay          time.Duration
	SavePDFBefore    func(page *rod.Page)
	MenuRootSelector string
	ParseMenu        func(doc *DocDownload, root *rod.Element, level int, dirPath string, bms *[]pdfcpu.Bookmark)
}

// NewDocDownload description
//
// createTime: 2023-07-26 11:42:19
//
// author: hailaz
func NewDocDownload(mainURL, outputDir string) *DocDownload {
	var browser *rod.Browser
	if binPath, exists := launcher.LookPath(); exists {
		u := launcher.New().Bin(binPath).MustLaunch()
		browser = rod.New().ControlURL(u).MustConnect()
	} else {
		// 如果没有找到浏览器，就使用默认的浏览器
		fmt.Println("没有找到浏览器，使用默认的浏览器，下载中...")
		browser = rod.New().MustConnect()
		fmt.Println("下载完成")
	}
	browser = browser.DefaultDevice(devices.Device{
		AcceptLanguage: "zh-CN",
	})
	// 从mainURL获取baseURL
	parsedURL, err := url.Parse(mainURL)
	if err != nil {
		fmt.Println("url.Parse Error:", err)
		return nil
	}
	baseURL := fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Host)
	return &DocDownload{
		MainURL:        mainURL,
		OutputDir:      outputDir,
		MergePDFNums:   20,
		TempSuffix:     ".temp.pdf",
		IsDownloadMain: false,
		fileList:       make([]string, 0),
		bookmark:       make([]pdfcpu.Bookmark, 0),
		pageFrom:       1,
		browser:        browser,
		baseURL:        baseURL,
		OpDelay:        200 * time.Millisecond,
	}
}

// Start 开始任务
//
// createTime: 2023-07-28 15:03:11
//
// author: hailaz
func (doc *DocDownload) Start() {
	doc.Show()
	if doc.IsDownloadMain {
		doc.Index(&doc.bookmark)
	}

	if doc.ParseMenu != nil {
		doc.ParseMenu(doc, doc.GetMenuRoot(doc.MenuRootSelector), 0, doc.OutputDir, &doc.bookmark)
	}

	if len(doc.fileList) > 0 {
		doc.MrPDF()
	}

	if len(doc.bookmark) > 0 {
		doc.AddBookmarks()
	}

}

// GetBrowser 返回浏览器对象
//
// createTime: 2023-07-28 14:23:07
//
// author: hailaz
func (doc *DocDownload) GetBrowser() *rod.Browser {
	return doc.browser
}

// Show description
//
// createTime: 2023-07-26 14:15:15
//
// author: hailaz
func (doc *DocDownload) Show() {
	fmt.Println("MainURL:", doc.MainURL)
	fmt.Println("OutputDir:", doc.OutputDir)
	fmt.Println("baseURL:", doc.baseURL)
}

// MrPDF description
//
// createTime: 2023-07-26 14:18:53
//
// author: hailaz
func (doc *DocDownload) MrPDF() {
	fLen := len(doc.fileList)
	preNum := doc.MergePDFNums
	fileName := doc.TempSuffix
	if preNum < 2 {
		preNum = 2
	}

	if fLen > 0 {
		index := 0
		tempOldName := ""
		tempName := fmt.Sprintf("%s.%d%s", doc.OutputDir, index, fileName)
		for {
			if index+preNum >= fLen {
				log.Printf("最后合并%d-%d(%d)", index, fLen, fLen)
				if index == 0 {
					api.MergeCreateFile(doc.fileList[index:fLen], doc.OutputDir+fileName, nil)
				} else {
					api.MergeCreateFile(append([]string{tempOldName}, doc.fileList[index:fLen]...), doc.OutputDir+fileName, nil)
					os.Remove(tempOldName)
				}
				break
			}
			log.Printf("临时合并%d-%d(%d)", index, index+preNum, fLen)
			if index == 0 {
				api.MergeCreateFile(doc.fileList[index:index+preNum], tempName, nil)
			} else {
				api.MergeCreateFile(append([]string{tempOldName}, doc.fileList[index:index+preNum]...), tempName, nil)
				os.Remove(tempOldName)
			}

			index += preNum
			tempOldName = tempName
			tempName = fmt.Sprintf("%s.%d%s", doc.OutputDir, index, fileName)
		}
	}
}

// AddBookmarks 添加书签
//
// createTime: 2023-07-26 16:22:46
//
// author: hailaz
func (doc *DocDownload) AddBookmarks() error {
	log.Println("添加书签", doc.OutputDir+".pdf")
	return api.AddBookmarksFile(doc.OutputDir+doc.TempSuffix, doc.OutputDir+".pdf", doc.bookmark, true, nil)
}

// GetMenuRoot description
//
// createTime: 2023-07-26 16:31:12
//
// author: hailaz
func (doc *DocDownload) GetMenuRoot(selector string) *rod.Element {
	return doc.browser.MustPage(doc.MainURL).MustWaitLoad().MustElement(selector)
}

// Index description
//
// createTime: 2023-07-28 14:42:42
//
// author: hailaz
func (doc *DocDownload) Index(bms *[]pdfcpu.Bookmark) {
	dirPath := doc.OutputDir
	text := "首页"
	*bms = append(*bms, pdfcpu.Bookmark{
		Title:    text,
		PageFrom: doc.pageFrom,
	})

	fileName := fmt.Sprintf("%s.pdf", text)
	doc.fileList = append(doc.fileList, path.Join(dirPath, fileName))

	doc.SavePDF(path.Join(dirPath, fileName), doc.MainURL)

	page, _ := api.PageCountFile(path.Join(dirPath, fileName))
	doc.pageFrom = doc.pageFrom + page
}

// SavePDF description
//
// createTime: 2023-07-11 16:51:31
//
// author: hailaz
func (doc *DocDownload) SavePDF(filePath string, pageUrl string) error {
	dir := path.Dir(filePath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		fmt.Println("创建目录", dir)
		os.MkdirAll(dir, os.ModePerm)
	}
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		page := doc.browser.MustPage(pageUrl).MustWaitLoad()
		defer page.Close()
		if doc.SavePDFBefore != nil {
			doc.SavePDFBefore(page)
		}
		var width float64 = 15
		r, err := page.PDF(&proto.PagePrintToPDF{
			// Landscape: true,
			PaperWidth: &width,
		})
		if err != nil {
			log.Printf("PDF[err]: %s", err)
			return err
		}
		bin, err := ioutil.ReadAll(r)
		if err != nil {
			log.Printf("ReadAll[err]: %s", err)
			return err
		}
		return utils.OutputFile(filePath, bin)
	}
	return nil
}

// Move 移动文件
//
// createTime: 2023-07-27 15:07:55
//
// author: hailaz
func (doc *DocDownload) Move(targetDir string) error {
	src := doc.OutputDir + ".pdf"
	dst := path.Join(targetDir, path.Base(src))

	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		fmt.Println("创建目录", targetDir)
		os.MkdirAll(targetDir, os.ModePerm)
	}

	if _, err := os.Stat(src); os.IsNotExist(err) {
		return err
	}
	// 复制文件
	return os.Rename(src, dst)

}
