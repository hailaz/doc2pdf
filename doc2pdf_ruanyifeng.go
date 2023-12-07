package doc2pdf

import (
	"fmt"
	"log"
	"path"
	"time"

	"github.com/go-rod/rod"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
)

// DownloadRuanyifengWeekly description
//
// createTime: 2023-12-07 16:26:50
//
// author: hailaz
func DownloadRuanyifengWeekly() {
	DownloadRuanyifeng("http://www.ruanyifeng.com/blog/weekly/", "./output/ruanyifeng")
}

// DownloadRuanyifeng 下载文档
//
// createTime: 2023-07-27 15:26:56
//
// author: hailaz
func DownloadRuanyifeng(mainURL string, outputDir string) {
	doc := NewDocDownload(mainURL, outputDir)
	doc.SavePDFBefore = func(page *rod.Page) {
		time.Sleep(time.Second * 1)
	}
	doc.MenuRootSelector = "div#alpha-inner"
	doc.ParseMenu = ParseRuanyifengMenu
	// doc.MergePDFNums = 10
	doc.Start()
	// 复制文件到其它目录
	// log.Println(doc.Move("./dist"))
}

// ParseRuanyifengMenu 解析菜单
//
// createTime: 2023-07-11 16:13:27
//
// author: hailaz
func ParseRuanyifengMenu(doc *DocDownload, root *rod.Element, level int, dirPath string, bms *[]pdfcpu.Bookmark) {
	index := 0
	// log.Println("ParseRuanyifengMenu", root.MustText())
	// 循环当前节点的li
	ulList, err := root.Elements("ul")
	if err != nil {
		log.Printf("[err]: %s", err)
		return
	}
	for _, ul := range ulList {
		liList, err := ul.Elements("li")
		if err != nil {
			log.Printf("[err]: %s", err)
			continue
		}
		// log.Printf("len %d", len(liList))

		for _, li := range liList {
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
				url := *href
				// 打印当前节点的层级和url

				// 打印当前节点的文本
				// fmt.Println(text)

				// fmt.Printf("%s[%s](%s)\n", strings.Repeat("--", level), text, url)
				// 保存pdf
				fileName := fmt.Sprintf("%s.pdf", text)
				doc.fileList = append(doc.fileList, path.Join(dirPath, fileName))
				doc.SavePDF(path.Join(dirPath, fileName), url)

				page, _ := api.PageCountFile(path.Join(dirPath, fileName))
				doc.pageFrom = doc.pageFrom + page

				log.Printf("文档累计页数%d，当前文件页数%d： %s\n", doc.pageFrom, page, path.Join(dirPath, fileName))

			}
			index++
		}
	}
}
