package doc2pdf

import (
	"fmt"
	"io"
	"log"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/PuerkitoBio/goquery"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/devices"
	"github.com/go-rod/rod/lib/proto"
	"github.com/go-rod/rod/lib/utils"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
)

var (
	versionList map[string]string = map[string]string{
		"v1.14":  "https://goframe.org/display/gf114/GoFrame+%28ZH%29-v1.14",
		"v1.15":  "https://goframe.org/pages/viewpage.action?pageId=7297616",
		"v1.16":  "https://goframe.org/display/gf116/GoFrame+%28ZH%29-v1.16",
		"v2.0":   "https://goframe.org/pages/viewpage.action?pageId=61149363",
		"v2.1":   "https://goframe.org/pages/viewpage.action?pageId=59864464",
		"v2.2":   "https://goframe.org/pages/viewpage.action?pageId=73224713",
		"v2.3":   "https://goframe.org/pages/viewpage.action?pageId=92131939",
		"v2.4":   "https://goframe.org/pages/viewpage.action?pageId=96885694",
		"latest": "https://goframe.org/display/gf",
	}
)

// DownloadGoFrameAll description
//
// createTime: 2023-07-28 15:27:17
//
// author: hailaz
func DownloadGoFrameAll() {
	wg := sync.WaitGroup{}
	for ver, main := range versionList {
		ver, main := ver, main
		wg.Add(1)
		func() {
			DownloadConfluence(main, "./output/goframe-"+ver)
			wg.Done()
		}()
	}
	wg.Wait()
}

// DownloadGoFrameWithVersion description
//
// createTime: 2023-08-08 18:34:08
//
// author: hailaz
func DownloadGoFrameWithVersion(version string) {
	if main, ok := versionList[version]; ok {
		DownloadConfluence(main, "./output/goframe-"+version)
	} else {
		log.Printf("版本号不存在")
	}
}

// DownloadGoFrameLatest description
//
// createTime: 2023-07-28 15:21:19
//
// author: hailaz
func DownloadGoFrameLatest() {
	DownloadConfluence("https://goframe.org/display/gf", "./output/goframe-latest")
}

// DownloadConfluence 下载confluence文档
//
// createTime: 2023-07-27 15:26:56
//
// author: hailaz
func DownloadConfluence(mainURL string, outputDir string) {
	doc := NewDocDownload(mainURL, outputDir)
	doc.PageToMD = PageToMD

	doc.GetBrowser().DefaultDevice(devices.Device{
		AcceptLanguage: "zh-CN",
	})

	doc.OpDelay = 100 * time.Millisecond

	doc.SavePDFBefore = func(page *rod.Page) {
		// 保存pdf前可自定义操作
		page.MustEval(`() => {
			// 右侧菜单加长显示
			var tocMacroDiv = document.querySelector("div.toc-macro");
			if(tocMacroDiv&&tocMacroDiv.style){
				tocMacroDiv.style.maxHeight = "5000px";
			} 


			// 代码块自动换行
			
			// 获取所有的 <pre> 元素
			const preElements = document.querySelectorAll('pre');

			// 循环遍历每个元素并设置样式
			preElements.forEach((preElement) => {
			preElement.style.whiteSpace = 'pre-wrap';
			preElement.style.wordWrap = 'break-word';
			});

			// 移除页脚
			var element = document.getElementById("footer");
			if (element) {
				element.parentNode.removeChild(element);
			}
			
		}`)
		// // 移除评论
		// var elementToRemove = document.getElementById('comments-section');
		// // 确认元素存在后再删除
		// if (elementToRemove) {
		// 	// 获取父级元素，并从父级中移除要删除的元素
		// 	var parentElement = elementToRemove.parentNode;
		// 	parentElement.removeChild(elementToRemove);
		// }

	}
	doc.MergePDFNums = 100
	doc.PageToPDF = func(page *rod.Page, filePath string) error {
		var width float64 = 15
		req := &proto.PagePrintToPDF{
			PrintBackground: true,
			PaperWidth:      &width,
		}

		err := PageToPDFWithCfg(page, filePath, req)
		if err != nil {
			return err
		}
		// 获取页数，合并成单页
		pageCount, err := api.PageCountFile(filePath)
		if err == nil {
			height := 11 * float64(pageCount)
			req.PaperHeight = &height
			return PageToPDFWithCfg(page, filePath, req)
		}

		return nil
	}
	doc.MenuRootSelector = "ul.plugin_pagetree_children_list.plugin_pagetree_children_list_noleftspace ul"
	doc.ParseMenu = ParseConfluenceMenu
	doc.IsDownloadMain = true
	doc.Start()
	// 复制文件到其它目录
	log.Println(doc.Move("./dist"))
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

// ParseConfluenceMenu 解析菜单
//
// createTime: 2023-07-11 16:13:27
//
// author: hailaz
func ParseConfluenceMenu(doc *DocDownload, root *rod.Element, level int, dirPath string, bms *[]pdfcpu.Bookmark) {
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
		log.Printf("title: %s\n", text)
		// if index >= 1 {
		// 	break
		// }
		// if len(*bms) >= 5 {
		// 	return
		// }
		// 拼接完整的url
		pageURL := doc.baseURL + *href
		if doc.Mode == DocDownloadModePDF {

			// 保存书签
			*bms = append(*bms, pdfcpu.Bookmark{
				Title:    text,
				PageFrom: doc.pageFrom,
			})
			// 保存pdf
			fileName := fmt.Sprintf("%d-%s.pdf", index, text)
			doc.fileList = append(doc.fileList, path.Join(dirPath, fileName))
			doc.SavePDF(path.Join(dirPath, fileName), pageURL)

			page, _ := api.PageCountFile(path.Join(dirPath, fileName))
			doc.pageFrom = doc.pageFrom + page

			log.Printf("文档累计页数%d，当前文件页数%d： %s\n", doc.pageFrom, page, path.Join(dirPath, fileName))
		}

		if a, err := li.Element("div.plugin_pagetree_childtoggle_container a"); err == nil {
			time.Sleep(500 * time.Millisecond)
			if err := a.Click(proto.InputMouseButtonLeft, 1); err == nil {
				time.Sleep(200 * time.Millisecond)
				// 如果当前节点有子节点
				count := 1
				for {
					time.Sleep(100 * time.Millisecond)
					if ul, err := li.Element("div.plugin_pagetree_children_container ul"); err == nil {
						// log.Printf("[子菜单]: %s", ul.MustText())
						// 递归子节点
						dirName := fmt.Sprintf("%d-%s", index, text)
						(*bms)[index].Children = make([]pdfcpu.Bookmark, 0)
						doc.ParseMenu(doc, ul, level+1, path.Join(dirPath, dirName), &((*bms)[index].Children))
						// index++
						log.Printf("在第%d次找到子节点\n", count)
						break
					} else {
						log.Printf("尝试第%d次，没有子节点，待重试: %s\n", count, err)
					}

					if count >= 50 {
						log.Printf("经过%d次，真的没有子节点\n", count)
						break
					}
					count++
				}

			}
			fileNameMD := fmt.Sprintf("%d-%s/%d-%s.md", index, text, index, text)

			doc.SaveMD(ReplacePath(path.Join(dirPath, fileNameMD), doc.OutputDir), pageURL)
		} else {
			fileNameMD := fmt.Sprintf("%d-%s.md", index, text)
			doc.SaveMD(ReplacePath(path.Join(dirPath, fileNameMD), doc.OutputDir), pageURL)
		}
		index++
	}

}

// ReplacePath description
//
// createTime: 2024-01-31 18:48:10
func ReplacePath(filePath string, outPath string) string {
	filePath = strings.ReplaceAll(filePath, "(🔥重点🔥)", "")
	filePath = strings.ReplaceAll(filePath, "🔥", "")
	filePath = strings.ReplaceAll(filePath, "(", "-")
	filePath = strings.ReplaceAll(filePath, ")", "")
	filePath = strings.Replace(filePath, outPath, outPath+"-md", 1)
	return filePath
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
