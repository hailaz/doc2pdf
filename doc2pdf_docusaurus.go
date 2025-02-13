package doc2pdf

import (
	"fmt"
	"log"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
)

// DownloadHailaz description
//
// createTime: 2023-07-28 15:21:19
//
// author: hailaz
func DownloadHailaz() {
	// http://www.hailaz.cn/docs/learn/index
	// DownloadDocusaurus("http://www.hailaz.cn/docs/learn/index", "./output/hailaz-learn")
	DownloadDocusaurus("https://www.hailaz.cn/docs/live/", "./output/hailaz-live")
}

// DownloadGoFrame 下载GoFrame文档
func DownloadGoFrame() {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		DownloadDocusaurus("https://goframe.org/quick/install", "./output/goframe/quick")
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		DownloadDocusaurus("https://goframe.org/docs/cli", "./output/goframe/docs")
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		DownloadDocusaurus("https://goframe.org/examples/grpc", "./output/goframe/examples")
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		DownloadDocusaurus("https://goframe.org/release/note", "./output/goframe/release")
	}()
	wg.Wait()
}

// DownloadDocusaurus 下载confluence文档
//
// createTime: 2023-07-27 15:26:56
//
// author: hailaz
func DownloadDocusaurus(mainURL string, outputDir string) {
	doc := NewDocDownload(mainURL, outputDir)
	doc.SavePDFBefore = func(page *rod.Page) {
		// 删除class 元素 petercat-lui-assistant
		page.MustEval(`() => {
			var elementToRemove = document.querySelector('.petercat-lui-assistant');
			// 确认元素存在后再删除
			if (elementToRemove) {
				// 获取父级元素，并从父级中移除要删除的元素
				var parentElement = elementToRemove.parentNode;
				parentElement.removeChild(elementToRemove);
			}

			// 移除评论 id comments
			var element = document.getElementById("comments");
			if (element) {
				element.parentNode.removeChild(element);
			}
		}`)
		time.Sleep(time.Millisecond * 500)
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
	log.Printf("开始解析菜单级别 %d, 路径: %s", level, dirPath)
	index := 0
	// 循环当前节点的li
	for li, err := root.Element("li"); err == nil; li, err = li.Next() {
		log.Printf("处理第 %d 个菜单项", index+1)

		// 获取当前节点的a标签
		a, err := li.Element("a")
		if err != nil {
			log.Printf("[错误] 获取a标签失败: %s", err)
			continue
		}

		// 获取a标签的href属性
		href, err := a.Attribute("href")
		if err != nil {
			log.Printf("[错误] 获取href属性失败: %s", err)
			continue
		}

		// 获取a标签的文本
		text, err := a.Text()
		if err != nil {
			log.Printf("[错误] 获取文本失败: %s", err)
			continue
		}
		log.Printf("正在处理菜单项: [%s] href=%s", text, *href)

		*bms = append(*bms, pdfcpu.Bookmark{
			Title:    text,
			PageFrom: doc.pageFrom,
		})

		{
			url := doc.baseURL + *href
			log.Printf("准备下载页面: %s", url)

			fileName := fmt.Sprintf("%d-%s.pdf", index, text)
			fullPath := path.Join(dirPath, fileName)
			log.Printf("保存PDF文件: %s", fullPath)

			doc.fileList = append(doc.fileList, fullPath)
			doc.SavePDF(fullPath, url)

			page, err := api.PageCountFile(fullPath)
			if err != nil {
				log.Printf("[错误] 获取PDF页数失败: %s", err)
			}
			doc.pageFrom = doc.pageFrom + page
			log.Printf("文档累计页数%d，当前文件页数%d： %s", doc.pageFrom, page, fullPath)
		}

		// 一级菜单 theme-doc-sidebar-item-link theme-doc-sidebar-item-link-level-1 menu__list-item
		// 一级菜单（目录） theme-doc-sidebar-item-category theme-doc-sidebar-item-category-level-1 menu__list-item
		// 二级菜单 theme-doc-sidebar-item-link theme-doc-sidebar-item-link-level-2 menu__list-item
		// 二级菜单（目录） theme-doc-sidebar-item-category theme-doc-sidebar-item-category-level-2 menu__list-item

		if class, err := li.Attribute("class"); err == nil {
			// 判断是否目录
			if strings.Contains(*class, "theme-doc-sidebar-item-category") {
				log.Printf("发现菜单: %s", text)
				// 判断是否展开
				if strings.Contains(*class, "menu__list-item--collapsed") {
					// 点击展开
					log.Printf("点击展开菜单: %s", text)
					root.Page().MustEval(`() => {
						// 滚动到页面底部
						window.scrollTo(0, document.documentElement.scrollHeight);
					}`)
					time.Sleep(doc.OpDelay)
					if err := a.Click(proto.InputMouseButtonLeft, 1); err == nil {
						log.Printf("点击展开菜单成功: %s", text)
						time.Sleep(doc.OpDelay)
					} else {
						log.Printf("[错误] 点击展开菜单失败: %s", err)
					}
				}

				if ul, err := li.Element("ul"); err == nil {
					log.Printf("开始处理子菜单: %s", text)
					dirName := fmt.Sprintf("%d-%s", index, text)
					(*bms)[index].Kids = make([]pdfcpu.Bookmark, 0)
					doc.ParseMenu(doc, ul, level+1, path.Join(dirPath, dirName), &((*bms)[index].Kids))
				} else {
					log.Printf("[错误] 获取子菜单ul元素失败: %s", err)
				}
			}
		}

		index++
		log.Printf("完成处理第 %d 个菜单项", index)
	}
	log.Printf("完成菜单级别 %d 的解析，共处理 %d 个项目", level, index)
}
