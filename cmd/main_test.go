package main_test

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/devices"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"github.com/go-rod/rod/lib/utils"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
)

// TestDown description
//
// createTime: 2022-08-11 11:13:59
//
// author: hailaz
func TestDown(t *testing.T) {
	download("https://goframe.org/display/gf", "goframe.html")
}

func download(url, filename string) (err error) {
	fmt.Println("Downloading ", url, " to ", filename)

	resp, err := http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	f, err := os.Create(filename)
	if err != nil {
		return
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	return
}

// ErrTest description
//
// createTime: 2022-08-16 15:31:47
//
// author: hailaz
func ErrTest() (err error) {
	var a int
	for i := 0; i < 10; i++ {
		a, err = ReturnErr()
		if err != nil {
			return
		}
		fmt.Println(a)
	}
	return nil
}

// ReturnErr description
//
// createTime: 2022-08-16 15:32:28
//
// author: hailaz
func ReturnErr() (int, error) {
	return 0, nil
}

// TestMR description
//
// createTime: 2023-07-26 09:54:18
//
// author: hailaz
func TestMR(t *testing.T) {
	fileList := make([]string, 0)
	dirPath := "./output/gfdoc"
	fileList = append(fileList, path.Join(dirPath, "/0-快速开始.pdf"))
	fileList = append(fileList, path.Join(dirPath, "/0-快速开始.pdf"))
	fileList = append(fileList, path.Join(dirPath, "/0-快速开始.pdf"))
	fileList = append(fileList, path.Join(dirPath, "/0-快速开始.pdf"))
	fileList = append(fileList, path.Join(dirPath, "/0-快速开始.pdf"))
	fileList = append(fileList, path.Join(dirPath, "/0-快速开始.pdf"))
	MRList(fileList, dirPath, ".temp.pdf", 2)
	// err := api.MergeCreateFile(fileList, dirPath+".mrtest.pdf", nil)
	// if err != nil {
	// 	log.Printf("[err]: %s", err)
	// }
}

// TestGfDoc description
//
// createTime: 2023-07-21 16:20:23
//
// author: hailaz
func TestGfDoc(t *testing.T) {
	if binPath, exists := launcher.LookPath(); exists {
		t.Log(binPath)
		u := launcher.New().Bin(binPath).MustLaunch()
		b := rod.New().ControlURL(u).MustConnect()
		b = b.DefaultDevice(devices.Device{
			AcceptLanguage: "zh-CN",
		})
		// b.MustPage("https://goframe.org/display/gf").MustWaitLoad().MustPDF("gf_index.pdf")

		menuEl := b.MustPage("https://goframe.org/display/gf").MustWaitLoad().MustElement("ul.plugin_pagetree_children_list.plugin_pagetree_children_list_noleftspace ul")
		t.Log(menuEl.MustText())
		bms := make([]pdfcpu.Bookmark, 0)
		fileList := make([]string, 0)
		pageFrom := 1

		baseUrl := "https://goframe.org"
		dirPath := "./output/gfdoc"
		FindMenuGf(b, menuEl, baseUrl, 0, dirPath, &pageFrom, &fileList, &bms)
		// g.Dump(bms)
		MRList(fileList, dirPath, ".temp.pdf", 50)
		// err := api.MergeCreateFile(fileList, dirPath+".temp.pdf", nil)
		// if err != nil {
		// 	log.Printf("[err]: %s", err)
		// }
		err := api.AddBookmarksFile(dirPath+".temp.pdf", dirPath+".pdf", bms, true, nil)
		if err != nil {
			log.Printf("[err]: %s", err)
		}
	}
}

// MRList 分批合并pdf
//
// createTime: 2023-07-26 10:02:57
//
// author: hailaz
func MRList(fileList []string, dirPath string, fileName string, preNum int) {
	fLen := len(fileList)

	if fLen > 0 {
		index := 0
		tempOldName := ""
		tempName := fmt.Sprintf("%s.%d%s", dirPath, index, fileName)
		for {
			if index+preNum >= fLen {
				log.Printf("合并%d-%d\n%+v", index, fLen, fileList[index:fLen])
				if index == 0 {
					api.MergeCreateFile(fileList[index:fLen], tempName, nil)
				} else {
					api.MergeCreateFile(append([]string{tempOldName}, fileList[index:fLen]...), dirPath+fileName, nil)
					os.Remove(tempOldName)
				}
				break
			}
			log.Printf("合并%d-%d\n%+v", index, index+preNum, fileList[index:index+preNum])
			if index == 0 {
				api.MergeCreateFile(fileList[index:index+preNum], tempName, nil)
			} else {
				api.MergeCreateFile(append([]string{tempOldName}, fileList[index:index+preNum]...), tempName, nil)
				os.Remove(tempOldName)
			}

			index += preNum
			tempOldName = tempName
			tempName = fmt.Sprintf("%s.%d%s", dirPath, index, fileName)
		}
	}
}

// FindMenu description
//
// createTime: 2023-07-11 16:13:27
//
// author: hailaz
func FindMenuGf(browser *rod.Browser, root *rod.Element, baseUrl string, level int, dirPath string, pageFrom *int, fileList *[]string, bms *[]pdfcpu.Bookmark) {
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
			PageFrom: *pageFrom,
		})

		{
			// 拼接完整的url
			url := baseUrl + *href
			// 打印当前节点的层级和url

			// 打印当前节点的文本
			// fmt.Println(text)

			fmt.Printf("%s[%s](%s)\n", strings.Repeat("--", level), text, url)
			// 保存pdf
			fileName := fmt.Sprintf("%d-%s.pdf", index, text)
			*fileList = append(*fileList, path.Join(dirPath, fileName))

			SavePdf(browser, path.Join(dirPath, fileName), url)

			page, _ := api.PageCountFile(path.Join(dirPath, fileName))
			*pageFrom = *pageFrom + page

			fmt.Printf("文件页码%d/%d： %s\n", page, *pageFrom, path.Join(dirPath, fileName))

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
					FindMenuGf(browser, ul, baseUrl, level+1, path.Join(dirPath, dirName), pageFrom, fileList, &((*bms)[index].Children))
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

// TestDownPdf description
//
// createTime: 2023-07-11 15:48:28
//
// author: hailaz
func TestDownPdf(t *testing.T) {
	if binPath, exists := launcher.LookPath(); exists {
		t.Log(binPath)
		u := launcher.New().Bin(binPath).MustLaunch()
		b := rod.New().ControlURL(u).MustConnect()
		b.MustPage("https://goframe.org/pages/viewpage.action?pageId=41900259").MustWaitLoad().MustPDF("sample.pdf")
		t.Log(b.MustPage("http://www.hailaz.cn/docs/learn/index").MustElement("ul.theme-doc-sidebar-menu.menu__list").MustText())
		bms := make([]pdfcpu.Bookmark, 0)
		fileList := make([]string, 0)
		pageFrom := 1
		FindMenu(
			b,
			b.MustPage("http://www.hailaz.cn/docs/learn/index").MustElement("ul.theme-doc-sidebar-menu.menu__list"),
			"http://www.hailaz.cn", 0, "./output", &pageFrom, &fileList, &bms)
		fmt.Println("wrote sample.pdf")
		// g.Dump(bms)
		api.MergeCreateFile(fileList, "./mr.pdf", nil)
		api.AddBookmarksFile("./mr.pdf", "./mrbm.pdf", bms, true, nil)
	}
	// rod.New().MustConnect().MustPage("https://www.google.com/").MustWaitLoad().MustPDF("sample.pdf")
	// fmt.Println("wrote sample.pdf")
}

// FindMenu description
//
// createTime: 2023-07-11 16:13:27
//
// author: hailaz
func FindMenu(browser *rod.Browser, root *rod.Element, baseUrl string, level int, dirPath string, pageFrom *int, fileList *[]string, bms *[]pdfcpu.Bookmark) {
	index := 0
	// 循环当前节点的li
	for li, err := root.Element("li"); err == nil; li, err = li.Next() {
		// 获取当前节点的a标签
		a, err := li.Element("a")
		if err != nil {
			continue
		}
		// 获取a标签的href属性
		href, err := a.Attribute("href")
		if err != nil {
			continue
		}
		// 获取a标签的文本
		text, err := a.Text()
		if err != nil {
			continue
		}
		fmt.Printf("title: %s\n", text)

		*bms = append(*bms, pdfcpu.Bookmark{
			Title:    text,
			PageFrom: *pageFrom,
		})

		if class, err := li.Attribute("class"); err == nil && strings.Contains(*class, "menu__list-item--collapsed") {
			if err := li.Click(proto.InputMouseButtonLeft, 1); err == nil {
				// 如果当前节点有子节点
				if ul, err := li.Element("ul"); err == nil {
					// 递归子节点
					dirName := fmt.Sprintf("%d-%s", index, text)
					(*bms)[index].Children = make([]pdfcpu.Bookmark, 0)
					FindMenu(browser, ul, baseUrl, level+1, path.Join(dirPath, dirName), pageFrom, fileList, &((*bms)[index].Children))
					index++
				} else {
					fmt.Println("没有子节点", err)
				}
			}

		} else {
			// 拼接完整的url
			url := baseUrl + *href
			// 打印当前节点的层级和url

			// 打印当前节点的文本
			// fmt.Println(text)

			fmt.Printf("%s[%s](%s)\n", strings.Repeat("--", level), text, url)
			// 保存pdf
			fileName := fmt.Sprintf("%d-%s.pdf", index, text)
			*fileList = append(*fileList, path.Join(dirPath, fileName))

			SavePdf(browser, path.Join(dirPath, fileName), url)

			page, _ := api.PageCountFile(path.Join(dirPath, fileName))
			*pageFrom = *pageFrom + page

			fmt.Printf("文件%d/%d： %s\n", page, *pageFrom, path.Join(dirPath, fileName))

			index++
		}

		// fmt.Println(li.Text())
	}

	// fmt.Println(root.MustElements("li").First().Text())
	// fmt.Println(root.MustElements("li").Last().MustNext().MustText())
}

// TestSavePdf description
//
// createTime: 2023-07-23 23:43:49
//
// author: hailaz
func TestSavePdf(t *testing.T) {
	os.Remove("./output/test/test.pdf")
	if binPath, exists := launcher.LookPath(); exists {
		t.Log(binPath)
		u := launcher.New().Bin(binPath).MustLaunch()
		b := rod.New().ControlURL(u).MustConnect()
		b = b.DefaultDevice(devices.Device{
			AcceptLanguage: "zh-CN",
		})
		SavePdf(b, "./output/test/test.pdf", "https://goframe.org/pages/viewpage.action?pageId=57183756")
	} else {
		b := rod.New().MustConnect()
		b = b.DefaultDevice(devices.Device{
			AcceptLanguage: "zh-CN",
		})
		SavePdf(b, "./output/test/test.pdf", "https://goframe.org/pages/viewpage.action?pageId=57183756")
	}
	// time.Sleep(time.Second * 10)
}

// SavePdf description
//
// createTime: 2023-07-11 16:51:31
//
// author: hailaz
func SavePdf(browser *rod.Browser, filePath string, pageUrl string) {
	fmt.Println(filePath)
	dir := path.Dir(filePath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		fmt.Println("创建目录", dir)
		os.MkdirAll(dir, os.ModePerm)
	}
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		page := browser.MustPage(pageUrl).MustWaitLoad()
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
		var width float64 = 10
		r, err := page.PDF(&proto.PagePrintToPDF{
			// Landscape: true,
			PaperWidth: &width,
		})
		if err != nil {
			log.Printf("PDF[err]: %s", err)
		}
		bin, err := ioutil.ReadAll(r)
		if err != nil {
			log.Printf("ReadAll[err]: %s", err)
		}
		utils.OutputFile(filePath, bin)
		page.Close()
	}
}

// TestMerge description
//
// createTime: 2023-07-14 11:45:06
//
// author: hailaz
func TestMerge(t *testing.T) {

	//api.MergeCreateFile([]string{"./output/0-目录.pdf", "./output/3-文档.pdf"}, "./mr.pdf", nil)

	api.AddBookmarksFile("./mr.pdf", "./mrbm.pdf", []pdfcpu.Bookmark{
		{PageFrom: 1, Title: "Page 1: Applicant’s Form"},
		{PageFrom: 2, Title: "Page 2: Bold 这是一个测试", Bold: true},
	}, true, nil)
}
