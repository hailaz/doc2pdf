package main_test

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"testing"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/devices"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"github.com/go-rod/rod/lib/utils"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
)

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
