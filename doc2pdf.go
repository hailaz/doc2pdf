package doc2pdf

import (
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"github.com/go-rod/rod/lib/utils"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
)

const (
	// DocDownloadModePDF pdf模式
	DocDownloadModePDF = "pdf"
	// DocDownloadModeMD markdown模式
	DocDownloadModeMD = "md"
)

// DocDownload description
type DocDownload struct {
	MainURL        string // 文档入口地址
	outputDir      string // 输出目录
	MergePDFNums   int    // 每次合并的pdf数量，多文档时能减轻内存压力
	TempSuffix     string // 临时文件后缀
	IsDownloadMain bool

	pageFrom int
	baseURL  string
	browser  *rod.Browser
	OpDelay  time.Duration

	Mode string // 下载模式: pdf,md

	// for pdf
	fileList      []string
	bookmark      []pdfcpu.Bookmark
	SavePDFBefore func(page *rod.Page)
	PageToPDF     func(page *rod.Page, filePath string) error
	// for markdown
	PageToMD func(doc *DocDownload, filePath string, pageUrl string) error

	// menu
	MenuRootSelector string
	ParseMenu        func(doc *DocDownload, root *rod.Element, level int, dirPath string, bms *[]pdfcpu.Bookmark) // 解析菜单
}

// NewDocDownload description
//
// createTime: 2023-07-26 11:42:19
//
// author: hailaz
func NewDocDownload(mainURL, outputDir string) *DocDownload {
	log.SetFlags(log.Llongfile | log.Ldate | log.Ltime)
	var browser *rod.Browser
	var launcherSet = launcher.New().Leakless(false)
	// 是否调试
	var isDebug = false
	if isDebug {
		launcherSet.Headless(false)
	}
	if binPath, exists := launcher.LookPath(); exists {
		log.Println("找到浏览器", binPath)
		launcherSet.Bin(binPath)
	} else {
		// 如果没有找到浏览器，就使用默认的浏览器
	}
	browser = rod.New().ControlURL(launcherSet.MustLaunch()).MustConnect()
	if isDebug {
		browser.SlowMotion(time.Second).Trace(true)
	}
	// 从mainURL获取baseURL
	parsedURL, err := url.Parse(mainURL)
	if err != nil {
		log.Println("url.Parse Error:", err)
		return nil
	}
	baseURL := fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Host)
	return &DocDownload{
		MainURL:        mainURL,
		outputDir:      path.Join(outputDir),
		MergePDFNums:   20,
		TempSuffix:     ".temp.pdf",
		IsDownloadMain: false,
		fileList:       make([]string, 0),
		bookmark:       make([]pdfcpu.Bookmark, 0),
		pageFrom:       1,
		browser:        browser,
		baseURL:        baseURL,
		OpDelay:        200 * time.Millisecond,
		PageToPDF:      PageToPDF,
		Mode:           DocDownloadModePDF,
	}
}

// Start 开始任务
//
// createTime: 2023-07-28 15:03:11
//
// author: hailaz
func (doc *DocDownload) Start() {
	doc.Show()
	log.Println("判断是否保存入口页")

	if doc.Mode == DocDownloadModeMD {
		gfile.Remove(doc.OutputDir())
		if doc.IsDownloadMain {
			doc.Index(&doc.bookmark)
		}
		if doc.ParseMenu != nil {
			log.Println("菜单解析")
			root := doc.GetMenuRoot(doc.MenuRootSelector)
			doc.ParseMenu(doc, root, 0, doc.OutputDir(), nil)
		}
	} else if doc.Mode == DocDownloadModePDF {
		if doc.IsDownloadMain {
			doc.Index(&doc.bookmark)
		}
		if doc.ParseMenu != nil {
			log.Println("菜单解析")
			root := doc.GetMenuRoot(doc.MenuRootSelector)
			doc.ParseMenu(doc, root, 0, doc.OutputDir(), &doc.bookmark)
		}

		log.Println("判断是否合并文件")
		if len(doc.fileList) > 0 {
			doc.MrPDF()
		}

		log.Println("判断是否有书签数据")
		if len(doc.bookmark) > 0 {
			doc.AddBookmarks()
		}
	}
	// 关闭浏览器
	doc.Close()
}

// GetBrowser 返回浏览器对象
//
// createTime: 2023-07-28 14:23:07
//
// author: hailaz
func (doc *DocDownload) GetBrowser() *rod.Browser {
	return doc.browser
}

// SetDebug description
//
// createTime: 2025-02-13 14:39:22
func (doc *DocDownload) SetDebug() {
	doc.browser.SlowMotion(time.Second).Trace(true)
}

// OutputDir description
//
// createTime: 2024-02-05 15:57:59
func (doc *DocDownload) OutputDir() string {
	if doc.Mode == DocDownloadModeMD {
		return doc.outputDir + "-md"
	}
	return doc.outputDir
}

// StaticDir description
//
// createTime: 2024-02-05 15:57:59
func (doc *DocDownload) StaticDir() string {
	return doc.OutputDir() + "-static"
}

// HTMLDir description
//
// createTime: 2024-02-05 15:57:59
func (doc *DocDownload) HTMLDir() string {
	return doc.OutputDir() + "-html"
}

// Close 关闭
//
// createTime: 2023-07-28 14:23:07
//
// author: hailaz
func (doc *DocDownload) Close() error {
	doc.browser.Close()
	return nil
}

// Show description
//
// createTime: 2023-07-26 14:15:15
//
// author: hailaz
func (doc *DocDownload) Show() {
	fmt.Println("MainURL:", doc.MainURL)
	fmt.Println("OutputDir:", doc.OutputDir())
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
		tempName := fmt.Sprintf("%s.%d%s", doc.OutputDir(), index, fileName)
		for {
			if index+preNum >= fLen {
				log.Printf("最后合并%d-%d(%d)", index, fLen, fLen)
				if index == 0 {
					api.MergeCreateFile(doc.fileList[index:fLen], doc.OutputDir()+fileName, false, nil)
				} else {
					api.MergeCreateFile(append([]string{tempOldName}, doc.fileList[index:fLen]...), doc.OutputDir()+fileName, false, nil)
					os.Remove(tempOldName)
				}
				break
			}
			log.Printf("临时合并%d-%d(%d)", index, index+preNum, fLen)
			if index == 0 {
				api.MergeCreateFile(doc.fileList[index:index+preNum], tempName, false, nil)
			} else {
				api.MergeCreateFile(append([]string{tempOldName}, doc.fileList[index:index+preNum]...), tempName, false, nil)
				os.Remove(tempOldName)
			}

			index += preNum
			tempOldName = tempName
			tempName = fmt.Sprintf("%s.%d%s", doc.OutputDir(), index, fileName)
		}
	}
}

// AddBookmarks 添加书签
//
// createTime: 2023-07-26 16:22:46
//
// author: hailaz
func (doc *DocDownload) AddBookmarks() error {
	log.Println("添加书签", doc.OutputDir()+".pdf")
	// gutil.Dump(doc.bookmark)
	return api.AddBookmarksFile(doc.OutputDir()+doc.TempSuffix, doc.OutputDir()+".pdf", doc.bookmark, true, nil)
}

// GetMenuRoot description
//
// createTime: 2023-07-26 16:31:12
//
// author: hailaz
func (doc *DocDownload) GetMenuRoot(selector string) *rod.Element {
	return doc.browser.MustPage(doc.MainURL).MustWaitStable().MustElement(selector)
}

// Index description
//
// createTime: 2023-07-28 14:42:42
//
// author: hailaz
func (doc *DocDownload) Index(bms *[]pdfcpu.Bookmark) {
	dirPath := doc.OutputDir()
	text := "首页"

	if doc.Mode == DocDownloadModePDF {
		*bms = append(*bms, pdfcpu.Bookmark{
			Title:    text,
			PageFrom: doc.pageFrom,
		})

		fileName := fmt.Sprintf("%s.pdf", text)
		filePath := path.Join(dirPath, fileName)
		doc.fileList = append(doc.fileList, filePath)

		err := doc.SavePDF(filePath, doc.MainURL)
		if err != nil {
			log.Println("SavePDF Error:", err)
		}
		page, err := api.PageCountFile(filePath)
		if err != nil {
			log.Println("SavePDF Error:", err)
		}
		doc.pageFrom = doc.pageFrom + page
	} else {
		fileNameMD := fmt.Sprintf("%s.md", text)
		filePath := path.Join(dirPath, fileNameMD)
		filePath = strings.ReplaceAll(filePath, "(🔥重点🔥)", "")
		filePath = strings.ReplaceAll(filePath, "🔥", "")
		filePath = strings.ReplaceAll(filePath, "(", "-")
		filePath = strings.ReplaceAll(filePath, ")", "")
		doc.SaveMD(filePath, doc.MainURL)
	}
}

// SavePDF description
//
// createTime: 2023-07-11 16:51:31
//
// author: hailaz
func (doc *DocDownload) SavePDF(filePath string, pageUrl string) error {
	// log.Println("SavePDF", filePath)
	if doc.PageToPDF == nil {
		return nil
	}
	dir := path.Dir(filePath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		log.Println("创建目录", dir)
		os.MkdirAll(dir, os.ModePerm)
	}
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		page := doc.browser.MustPage(pageUrl).MustWaitStable()
		defer page.Close()
		if doc.SavePDFBefore != nil {
			doc.SavePDFBefore(page)
		}
		if err := doc.PageToPDF(page, filePath); err != nil {
			return err
		}
	}
	return nil
}

// SaveMD description
//
// createTime: 2023-07-11 16:51:31
//
// author: hailaz
func (doc *DocDownload) SaveMD(filePath string, pageUrl string) error {
	if doc.PageToMD == nil {
		return nil
	}
	dir := path.Dir(filePath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		fmt.Println("创建目录", dir)
		os.MkdirAll(dir, os.ModePerm)
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		dir = path.Dir(filePath)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			fmt.Println("创建目录", dir)
			os.MkdirAll(dir, os.ModePerm)
		}

		if err := doc.PageToMD(doc, filePath, pageUrl); err != nil {
			return err
		}
	}
	return nil
}

// PageToPDF description
//
// createTime: 2023-07-28 16:45:39
//
// author: hailaz
func PageToPDF(page *rod.Page, filePath string) error {
	page.MustPDF(filePath)
	return nil
}

// PageToPDFWithCfg description
//
// createTime: 2023-08-03 11:02:25
//
// author: hailaz
func PageToPDFWithCfg(page *rod.Page, filePath string, req *proto.PagePrintToPDF) error {
	r, err := page.PDF(req)
	if err != nil {
		log.Printf("PDF[err]: %s", err)
		return err
	}
	bin, err := io.ReadAll(r)
	if err != nil {
		log.Printf("ReadAll[err]: %s", err)
		return err
	}
	return utils.OutputFile(filePath, bin)
}

// Move 移动文件
//
// createTime: 2023-07-27 15:07:55
//
// author: hailaz
func (doc *DocDownload) Move(targetDir string) error {
	src := doc.OutputDir() + ".pdf"
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
