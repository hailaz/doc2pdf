package doc2pdf

import (
	"fmt"
	"log"
	"path"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
)

// DownloadGoFrameLatest description
//
// createTime: 2023-07-28 15:21:19
//
// author: hailaz
func DownloadHailaz() {
	// http://www.hailaz.cn/docs/learn/index
	// DownloadDocusaurus("http://www.hailaz.cn/docs/learn/index", "./output/hailaz-learn")
	DownloadDocusaurus("https://www.hailaz.cn/docs/live/", "./output/hailaz-live")
}

// DownloadConfluence 下载confluence文档
//
// createTime: 2023-07-27 15:26:56
//
// author: hailaz
func DownloadDocusaurus(mainURL string, outputDir string) {
	doc := NewDocDownload(mainURL, outputDir)
	doc.SavePDFBefore = func(page *rod.Page) {
		time.Sleep(time.Second * 1)
	}
	doc.MenuRootSelector = "ul.theme-doc-sidebar-menu.menu__list"
	doc.ParseMenu = ParseDocusaurusMenu
	doc.Start()
	// 复制文件到其它目录
	// log.Println(doc.Move("./dist"))
}

// ParseDocusaurusMenu 解析菜单
//
// createTime: 2023-07-11 16:13:27
//
// author: hailaz
func ParseDocusaurusMenu(doc *DocDownload, root *rod.Element, level int, dirPath string, bms *[]pdfcpu.Bookmark) {
	index := 0
	// log.Println("ParseDocusaurusMenu", root.MustText())
	// 循环当前节点的li
	for li, err := root.Element("li"); err == nil; li, err = li.Next() {
		// 获取当前节点的a标签
		a, err := li.Element("a")
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
		// log.Printf("title: %s\n", text)

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

			log.Printf("文档累计页数%d，当前文件页数%d： %s\n", doc.pageFrom, page, path.Join(dirPath, fileName))

		}
		if class, err := li.Attribute("class"); err == nil && strings.Contains(*class, "menu__list-item--collapsed") {
			if err := a.Click(proto.InputMouseButtonLeft, 1); err == nil {
				time.Sleep(doc.OpDelay)
				// 如果当前节点有子节点
				if ul, err := li.Element("ul"); err == nil {
					// log.Printf("[子菜单]: %s", ul.MustText())
					// 递归子节点
					dirName := fmt.Sprintf("%d-%s", index, text)
					(*bms)[index].Children = make([]pdfcpu.Bookmark, 0)
					doc.ParseMenu(doc, ul, level+1, path.Join(dirPath, dirName), &((*bms)[index].Children))
					// index++
				} else {
					fmt.Println("没有子节点", err)
				}
			}

		}
		index++
	}

}
