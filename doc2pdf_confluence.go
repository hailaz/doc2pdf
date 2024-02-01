package doc2pdf

import (
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"regexp"
	"strings"
	"sync"
	"time"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/JohannesKaufmann/html-to-markdown/plugin"
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
	mapData = make(map[string]string)
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
			DownloadConfluence(main, "./output/goframe-"+ver, DocDownloadModePDF)
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
		DownloadConfluence(main, "./output/goframe-"+version, DocDownloadModePDF)
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
	DownloadConfluence("https://goframe.org/display/gf", "./output/goframe-latest", DocDownloadModePDF)
}

// DownloadConfluence 下载confluence文档
//
// createTime: 2023-07-27 15:26:56
//
// author: hailaz
func DownloadConfluence(mainURL string, outputDir string, mode string) {
	doc := NewDocDownload(mainURL, outputDir)
	doc.PageToMD = PageToMD
	doc.Mode = mode

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
	if doc.Mode == DocDownloadModePDF {
		// 复制文件到其它目录
		log.Println(doc.Move("./dist"))
	}
	// 遍历文件，替换链接
	if doc.Mode == DocDownloadModeMD {
		// 地址转换
		files, err := gfile.ScanDir(doc.OutputDir+"-md", "*.md", true)
		if err != nil {
			log.Fatal(err)
		}
		regx := regexp.MustCompile(`/pages/viewpage.action\?pageId=\d+|/display/gf/.*\)`)
		for _, file := range files {
			contents := gfile.GetContents(file)
			urlList := regx.FindAllString(contents, -1)
			for _, u := range urlList {
				u = strings.TrimSuffix(u, ")")
				// t.Log(u)
				// t.Log(mapData[u])
				if newURL, ok := mapData[u]; ok {
					contents = strings.ReplaceAll(contents, u, newURL)
				} else {
					log.Println("not found")
					contents = strings.ReplaceAll(contents, u, doc.baseURL+u)
				}
			}
			contents = strings.ReplaceAll(contents, "；]", "]")
			contents = strings.ReplaceAll(contents, "；)", ")")
			contents = strings.ReplaceAll(contents, "- ```", "```")
			contents = strings.ReplaceAll(contents, "git.w", "woahailaz")
			contents = strings.ReplaceAll(contents, "woahailazoa.com", "github.com")
			contents = strings.ReplaceAll(contents, "git.code.o", "gitcodeohailaz")
			contents = strings.ReplaceAll(contents, "gitcodeohailaza.com", "github.com")

			err := gfile.PutContents(file, contents)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
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
		docTitle, err := a.Text()
		if err != nil {
			continue
		}
		log.Printf("title: %s\n", docTitle)
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
				Title:    docTitle,
				PageFrom: doc.pageFrom,
			})
			// 保存pdf
			fileName := fmt.Sprintf("%d-%s.pdf", index, docTitle)
			doc.fileList = append(doc.fileList, path.Join(dirPath, fileName))
			doc.SavePDF(path.Join(dirPath, fileName), pageURL)

			page, _ := api.PageCountFile(path.Join(dirPath, fileName))
			doc.pageFrom = doc.pageFrom + page

			log.Printf("文档累计页数%d，当前文件页数%d： %s\n", doc.pageFrom, page, path.Join(dirPath, fileName))
		}
		fileNameMD := ""
		if a, err := li.Element("div.plugin_pagetree_childtoggle_container a"); err == nil {
			time.Sleep(300 * time.Millisecond)
			if err := a.Click(proto.InputMouseButtonLeft, 1); err == nil {
				time.Sleep(200 * time.Millisecond)
				// 如果当前节点有子节点
				count := 1
				for {
					time.Sleep(100 * time.Millisecond)
					if ul, err := li.Element("div.plugin_pagetree_children_container ul"); err == nil {
						// log.Printf("[子菜单]: %s", ul.MustText())
						// 递归子节点
						dirName := fmt.Sprintf("%d-%s", index, docTitle)
						if bms != nil {
							(*bms)[index].Children = make([]pdfcpu.Bookmark, 0)
							doc.ParseMenu(doc, ul, level+1, path.Join(dirPath, dirName), &((*bms)[index].Children))
						} else {
							doc.ParseMenu(doc, ul, level+1, path.Join(dirPath, dirName), nil)
						}
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
			fileNameMD = fmt.Sprintf("%d-%s/%d-%s.md", index, docTitle, index, docTitle)
		} else {
			fileNameMD = fmt.Sprintf("%d-%s.md", index, docTitle)
		}
		if doc.Mode == DocDownloadModeMD {
			filePath := ReplacePath(path.Join(dirPath, fileNameMD), doc.OutputDir)
			doc.SaveMD(filePath, pageURL)
			SaveMap(filePath, pageURL)
			// 加标题
			contents := gfile.GetContents(filePath)
			contents = fmt.Sprintf("---\ntitle: %s\n---\n\n", docTitle) + contents
			gfile.PutContents(filePath, contents)
		}
		// mapData := fmt.Sprintf("%s=>%s\n", pageURL, ReplacePath(path.Join(dirPath, fileNameMD), doc.OutputDir))
		// gfile.PutContentsAppend(path.Join(doc.OutputDir+"-md-map", "map.txt"), mapData)
		index++
	}

}

// SaveMap description
//
// createTime: 2024-02-01 10:07:46
func SaveMap(filePath string, pageURL string) {
	pageURL = strings.ReplaceAll(pageURL, "https://goframe.org", "")
	pageURL = strings.ReplaceAll(pageURL, "&src=contextnavpagetreemode", "")
	pageURL = strings.ReplaceAll(pageURL, "?src=contextnavpagetreemode", "")
	filePath = strings.ReplaceAll(filePath, "output/goframe-latest-md", "/docs")
	// 编写正则表达式
	regex := regexp.MustCompile(`/\d+-`)
	filePath = regex.ReplaceAllString(filePath, "/")
	filePath = strings.TrimSuffix(filePath, ".md")
	mapData[pageURL] = filePath
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
func PageToMD(doc *DocDownload, filePath string, pageUrl string) error {

	// page.GetResource()

	// page.MustEval(`() => {
	// 	// 移除菜单
	// 	var element = document.querySelector('.ia-splitter-left');
	// 	if(element) {
	// 		element.parentNode.removeChild(element);
	// 	}

	// }`)

	cacheHtml := strings.ReplaceAll(filePath, doc.OutputDir+"-md", doc.OutputDir+"-html")
	cacheHtml = strings.TrimSuffix(cacheHtml, ".md") + ".html"
	html := ""

	if _, err := os.Stat(cacheHtml); os.IsNotExist(err) {
		// 加个缓存，免得每次都下载
		page := doc.browser.MustPage(pageUrl).MustWaitLoad()
		defer page.Close()
		html, _ := page.HTML()

		queryDoc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
		if err != nil {
			log.Fatal(err)
		}
		queryDoc.Find("div.page-metadata").Remove()
		queryDoc.Find("div.cell.aside").Remove()
		queryDoc.Find("#likes-and-labels-container").Remove()
		queryDoc.Find("#comments-section").Remove()
		// page.MustElement("img").MustResource()
		host := doc.baseURL
		// pageDir := path.Dir(filePath)
		pageDir := path.Join(doc.OutputDir + "-md-static")

		queryDoc.Find("#main-content").Find("img").Each(func(i int, s *goquery.Selection) {
			src, _ := s.Attr("src")
			// log.Println("img src:", src)
			if strings.HasPrefix(src, "https://") || strings.HasPrefix(src, "http://") {
				return
			}
			// resBaseName := strings.Split(filepath.Base(src), "?")[0]
			// 保存资源文件
			res, err := page.GetResource(host + src)
			if err != nil {
				log.Fatal(err)
				s.SetAttr("src", host+src)
			}
			resPath := path.Join(pageDir, strings.Split(src, "?")[0])
			// fmt.Println("resPath", resPath)
			// log.Println("save file:", resPath)
			err = gfile.PutBytes(resPath, res)
			if err != nil {
				log.Fatal(err)
			}
			// // 替换src
			// s.SetAttr("src", resBaseName)
			// s.SetAttr("src", host+src)
			// log.Println("src change", resBaseName)
		})
		html, _ = queryDoc.Find("#main-content").Html()

		gfile.PutContents(cacheHtml, html)
	} else {
		html = gfile.GetContents(cacheHtml)
	}

	converter := md.NewConverter("", true, nil)
	// converter.Use(plugin.ConfluenceCodeBlock())
	// converter.Use(plugin.ConfluenceAttachments())
	converter.Use(plugin.Strikethrough(""))
	markdown, err := converter.ConvertString(html)
	if err != nil {
		log.Fatal(err)
	}
	gfile.PutContents(filePath, markdown)
	return nil
}
