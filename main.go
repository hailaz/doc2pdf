package main

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"path"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
)

// main description
//
// createTime: 2023-07-26 11:42:19
//
// author: hailaz
func main() {
	// doc := NewDocDownload("https://goframe.org/pages/viewpage.action?pageId=7296490", "./output/hailaz")
	doc := NewDocDownload("https://goframe.org/display/gf", "./output/gfdoc")
	doc.Show()
	doc.ParseMenu(doc.GetMenuRoot("ul.plugin_pagetree_children_list.plugin_pagetree_children_list_noleftspace ul"), 0, doc.OutputDir, &doc.bookmark)
	doc.MrPDF(50)
	doc.AddBookmarks()
}

// DocDownload description
type DocDownload struct {
	MainURL   string
	OutputDir string
	fileList  []string
	bookmark  []pdfcpu.Bookmark
	pageFrom  int
	baseURL   string
	browser   *rod.Browser
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
		browser = rod.New().MustConnect()
	}
	// 从mainURL获取baseURL
	parsedURL, err := url.Parse(mainURL)
	if err != nil {
		fmt.Println("url.Parse Error:", err)
		return nil
	}
	baseURL := fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Host)
	return &DocDownload{
		MainURL:   mainURL,
		OutputDir: outputDir,
		fileList:  make([]string, 0),
		bookmark:  make([]pdfcpu.Bookmark, 0),
		pageFrom:  1,
		browser:   browser,
		baseURL:   baseURL,
	}
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
func (doc *DocDownload) MrPDF(preNum int) {
	fLen := len(doc.fileList)
	fileName := ".temp.pdf"
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
	return api.AddBookmarksFile(doc.OutputDir+".temp.pdf", doc.OutputDir+".pdf", doc.bookmark, true, nil)
}

// GetMenuRoot description
//
// createTime: 2023-07-26 16:31:12
//
// author: hailaz
func (doc *DocDownload) GetMenuRoot(selector string) *rod.Element {
	return doc.browser.MustPage(doc.MainURL).MustWaitLoad().MustElement(selector)
}

// ParseMenu 解析菜单
//
// createTime: 2023-07-11 16:13:27
//
// author: hailaz
func (doc *DocDownload) ParseMenu(root *rod.Element, level int, dirPath string, bms *[]pdfcpu.Bookmark) {
	index := 0
	// 循环当前节点的li
	for li, err := root.Element("li"); err == nil; li, err = li.Next() {
		// 获取当前节点的a标签
		a, err := li.Element("div.plugin_pagetree_children_content a")
		if err != nil {
			log.Printf("[err]: %s", err)
			continue
		}
		// 获取a标签的href属性
		href, err := a.Attribute("href")
		if err != nil {
			log.Printf("[err]: %s", err)
			continue
		}
		// 获取a标签的文本
		text, err := a.Text()
		if err != nil {
			continue
		}
		// fmt.Printf("title: %s\n", text)

		*bms = append(*bms, pdfcpu.Bookmark{
			Title:    text,
			PageFrom: doc.pageFrom,
		})

		{
			// 拼接完整的url
			url := doc.baseURL + *href
			// 打印当前节点的层级和url

			// 打印当前节点的文本
			// fmt.Println(text)

			// fmt.Printf("%s[%s](%s)\n", strings.Repeat("--", level), text, url)
			// 保存pdf
			fileName := fmt.Sprintf("%d-%s.pdf", index, text)
			doc.fileList = append(doc.fileList, path.Join(dirPath, fileName))

			doc.SavePDF(path.Join(dirPath, fileName), url)

			page, _ := api.PageCountFile(path.Join(dirPath, fileName))
			doc.pageFrom = doc.pageFrom + page

			fmt.Printf("文件页码%d/%d： %s\n", page, doc.pageFrom, path.Join(dirPath, fileName))

		}
		if a, err := li.Element("div.plugin_pagetree_childtoggle_container a"); err == nil {
			if err := a.Click(proto.InputMouseButtonLeft, 1); err == nil {
				time.Sleep(200 * time.Millisecond)
				// 如果当前节点有子节点
				if ul, err := li.Element("div.plugin_pagetree_children_container ul"); err == nil {
					// log.Printf("[子菜单]: %s", ul.MustText())
					// 递归子节点
					dirName := fmt.Sprintf("%d-%s", index, text)
					(*bms)[index].Children = make([]pdfcpu.Bookmark, 0)
					doc.ParseMenu(ul, level+1, path.Join(dirPath, dirName), &((*bms)[index].Children))
					// index++
				} else {
					fmt.Println("没有子节点", err)
				}
			}

		}
		index++
		// fmt.Println(li.Text())
	}

	// fmt.Println(root.MustElements("li").First().Text())
	// fmt.Println(root.MustElements("li").Last().MustNext().MustText())
}

// SavePDF description
//
// createTime: 2023-07-11 16:51:31
//
// author: hailaz
func (doc *DocDownload) SavePDF(filePath string, pageUrl string) {
	fmt.Println(filePath)
	dir := path.Dir(filePath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		fmt.Println("创建目录", dir)
		os.MkdirAll(dir, os.ModePerm)
	}
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		page := doc.browser.MustPage(pageUrl).MustWaitLoad()
		// _ = proto.EmulationSetLocaleOverride{Locale: "zh-CN"}.Call(page)
		// page.MustEmulate(devices.)
		page.MustEval(`() => {
	var tocMacroDiv = document.querySelector("div.toc-macro");
	if(tocMacroDiv&&tocMacroDiv.style){
		tocMacroDiv.style.maxHeight = "5000px";
	} 
}`)
		// time.Sleep(time.Second * 10)
		// menu,err:=page.Element("div.toc-macro")
		// if err==nil{
		// 	menu.
		// }
		page.MustPDF(filePath)
		page.Close()
	}
}
