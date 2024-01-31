package doc2pdf

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/PuerkitoBio/goquery"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/gogf/gf/v2/os/gfile"
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
	PageToPDF        func(page *rod.Page, filePath string) error
	PageToMD         func(page *rod.Page, filePath string) error
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
		log.Println("找到浏览器", binPath)
		u := launcher.New().Leakless(false).Bin(binPath).MustLaunch()
		browser = rod.New().ControlURL(u).MustConnect()
	} else {
		// 如果没有找到浏览器，就使用默认的浏览器
		log.Println("没有找到浏览器，使用默认的浏览器，下载中...")
		browser = rod.New().MustConnect()
		log.Println("下载完成")
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
		OutputDir:      path.Join(outputDir),
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
	if doc.IsDownloadMain {
		doc.Index(&doc.bookmark)
	}

	if doc.ParseMenu != nil {
		log.Println("菜单解析")
		root := doc.GetMenuRoot(doc.MenuRootSelector)
		doc.ParseMenu(doc, root, 0, doc.OutputDir, &doc.bookmark)
	}

	log.Println("判断是否合并文件")
	if len(doc.fileList) > 0 {
		doc.MrPDF()
	}

	log.Println("判断是否有书签数据")
	if len(doc.bookmark) > 0 {
		doc.AddBookmarks()
	}
	doc.browser.Close()
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

	fileName := fmt.Sprintf("%s.md", text)
	doc.fileList = append(doc.fileList, path.Join(dirPath, fileName))

	doc.SavePDF(path.Join(dirPath, fileName), doc.MainURL)
	fileNameMD := fmt.Sprintf("%s.md", text)
	doc.SaveMD(path.Join(dirPath, fileNameMD), doc.MainURL, 0)

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
		if doc.PageToPDF != nil {
			if err := doc.PageToPDF(page, filePath); err != nil {
				return err
			}
		}
	}
	return nil
}

// SaveMD description
//
// createTime: 2023-07-11 16:51:31
//
// author: hailaz
func (doc *DocDownload) SaveMD(filePath string, pageUrl string, plen int) error {
	filePath = strings.ReplaceAll(filePath, "(🔥重点🔥)", "")
	filePath = strings.ReplaceAll(filePath, "🔥", "")
	filePath = strings.ReplaceAll(filePath, "(", "-")
	filePath = strings.ReplaceAll(filePath, ")", "")
	filePath = strings.Replace(filePath, doc.OutputDir, doc.OutputDir+"-md", 1)
	// markdown
	fmt.Println("filePath", filePath, doc.OutputDir, plen)
	// if plen==1{
	// 	filePath=filePath
	// }
	dir := path.Dir(filePath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		fmt.Println("创建目录", dir)
		os.MkdirAll(dir, os.ModePerm)
	}
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		page := doc.browser.MustPage(pageUrl).MustWaitLoad()
		defer page.Close()

		dir = path.Dir(filePath)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			fmt.Println("创建目录", dir)
			os.MkdirAll(dir, os.ModePerm)
		}
		if doc.PageToMD != nil {
			if err := doc.PageToMD(page, filePath); err != nil {
				return err
			}
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

// PageToMD description
//
// createTime: 2023-07-28 16:45:39
//
// author: hailaz
func PageToMD(page *rod.Page, filePath string) error {
	// page.GetResource()

	// page.MustEval(`() => {
	// 	// 移除菜单
	// 	var element = document.querySelector('.ia-splitter-left');
	// 	if(element) {
	// 		element.parentNode.removeChild(element);
	// 	}

	// }`)
	html, _ := page.HTML()

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		log.Fatal(err)
	}
	doc.Find("div.page-metadata").Remove()
	doc.Find("div.cell.aside").Remove()
	doc.Find("#likes-and-labels-container").Remove()
	doc.Find("#comments-section").Remove()
	// page.MustElement("img").MustResource()
	host := "https://goframe.org"
	// pageDir := path.Dir(filePath)

	doc.Find("#content").Find("img").Each(func(i int, s *goquery.Selection) {
		src, _ := s.Attr("src")
		resBaseName := strings.Split(filepath.Base(src), "?")[0]
		// // 保存资源文件
		// res, err := page.GetResource(host + src)
		// if err != nil {
		// 	fmt.Println("GetResource", err)
		// }
		// resPath := path.Join(pageDir, resBaseName)
		// // fmt.Println("resPath", resPath)
		// gfile.PutBytes(resPath, res)

		// // 替换src
		// s.SetAttr("src", resBaseName)
		s.SetAttr("src", host+src)
		fmt.Println("src change", resBaseName)
	})
	html, _ = doc.Find("#content").Html()
	// fmt.Println(content)
	// fmt.Println(html)
	converter := md.NewConverter("", true, nil)
	// converter.Use(func(c *md.Converter) []md.Rule {
	// 	// character := "```"
	// 	return []md.Rule{
	// 		{
	// 			Filter: []string{"disssv"},
	// 			Replacement: func(content string, selec *goquery.Selection, opt *md.Options) *string {
	// 				fmt.Println("ac:structured-macro", content)

	// 				// for _, node := range selec.Nodes {
	// 				// // 	if node.Data == "ac:structured-macro" {
	// 				// // 		// node's last child -> <ac:plain-text-body>. We don't want to filter on that
	// 				// // 		// because we would end up with structured-macro around us.
	// 				// // 		// ac:plain-text-body's last child is [CDATA which has the actual content we are looking for.
	// 				// // 		data := strings.TrimPrefix(node.LastChild.LastChild.Data, "[CDATA[")
	// 				// // 		data = strings.TrimSuffix(data, "]]")
	// 				// // 		// content, if set, will contain the language that has been set in the field.
	// 				// // 		var language string
	// 				// // 		if content != "" {
	// 				// // 			language = content
	// 				// // 		}
	// 				// // 		formatted := fmt.Sprintf("%s%s\n%s\n%s", character, language, data, character)
	// 				// // 		return md.String(formatted)
	// 				// // 	}
	// 				// }
	// 				return md.String(content)
	// 			},
	// 		},
	// 	}
	// })
	// converter.Use(plugin.ConfluenceCodeBlock())
	// converter.Use(plugin.ConfluenceAttachments())
	markdown, err := converter.ConvertString(html)
	if err != nil {
		log.Fatal(err)
	}
	// fmt.Println("md ->", markdown)

	gfile.PutContents(filePath, markdown)
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
