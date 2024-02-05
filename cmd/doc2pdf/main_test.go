package main_test

import (
	"fmt"
	"log"
	"os"
	"path"
	"regexp"
	"strings"
	"testing"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/devices"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/hailaz/doc2pdf"
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
	outputFile := "./output/test/test.pdf"
	mainURL := "https://goframe.org/pages/viewpage.action?pageId=1115782"
	var b *rod.Browser
	if binPath, exists := launcher.LookPath(); exists {
		t.Log(binPath)
		u := launcher.New().Leakless(false).Bin(binPath).MustLaunch()
		b = rod.New().ControlURL(u).MustConnect()

	} else {
		b = rod.New().MustConnect()

	}
	b.DefaultDevice(devices.Device{
		AcceptLanguage: "zh-CN",
	})
	err := SavePdf(b, outputFile, mainURL)
	if err != nil {
		t.Error(err)
	}
	// time.Sleep(time.Second * 10)
}

// SavePdf description
//
// createTime: 2023-07-11 16:51:31
//
// author: hailaz
func SavePdf(browser *rod.Browser, filePath string, pageUrl string) error {
	fmt.Println(filePath)
	dir := path.Dir(filePath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		fmt.Println("创建目录", dir)
		os.MkdirAll(dir, os.ModePerm)
	}
	if _, err := os.Stat(filePath); os.IsNotExist(err) || !os.IsNotExist(err) {
		page := browser.MustPage(pageUrl).MustWaitLoad()
		defer page.Close()
		// _ = proto.EmulationSetLocaleOverride{Locale: "zh-CN"}.Call(page)
		// page.MustEmulate(devices.)
		page.MustEval(`() => {
	var tocMacroDiv = document.querySelector("div.toc-macro");
	if(tocMacroDiv&&tocMacroDiv.style){
		tocMacroDiv.style.maxHeight = "5000px";
	} 

	// 获取所有的 <pre> 元素
	const preElements = document.querySelectorAll('pre');

	// 循环遍历每个元素并设置样式
	preElements.forEach((preElement) => {
	preElement.style.whiteSpace = 'pre-wrap';
	preElement.style.wordWrap = 'break-word';
	});
	var element = document.getElementById("footer");
	if (element) {
		element.parentNode.removeChild(element);
	}
	

	

}`)

		//  wget -O gf https://github.com/gogf/gf/releases/latest/download/gf_$(go env GOOS)_$(go env GOARCH) && chmod +x gf && ./gf install -y && rm ./gf
		// 		val := page.MustEval(`() => {
		// 	return document.body.offsetHeight;
		// }`)
		// 		log.Printf("offsetHeight: %v", val)

		// time.Sleep(time.Second * 10)
		// menu,err:=page.Element("div.toc-macro")
		// if err==nil{
		// 	menu.
		// }
		var width float64 = 15
		req := &proto.PagePrintToPDF{
			PrintBackground: true,
			PaperWidth:      &width,
		}

		err := doc2pdf.PageToPDFWithCfg(page, filePath, req)
		if err != nil {
			return err
		}
		// 获取页数，合并成单页
		pageCount, err := api.PageCountFile(filePath)
		if err == nil {
			height := 11 * float64(pageCount)
			req.PaperHeight = &height
			return doc2pdf.PageToPDFWithCfg(page, filePath+".pdf", req)
		}

	}
	return nil
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

// TestMDPath description
//
// createTime: 2024-01-31 22:29:45
func TestMDPath(t *testing.T) {
	mdpath := "output/goframe-latest-md/15-其他资料/15-其他资料.md"

	// 编写正则表达式
	regex := regexp.MustCompile(`/\d+-`)

	mdpath = regex.ReplaceAllString(mdpath, "/")
	mdpath = strings.TrimSuffix(mdpath, ".md")

	t.Log(mdpath)

}

// TestFuncName description
//
// createTime: 2024-01-31 22:38:03
func TestFuncName(t *testing.T) {
	files, err := gfile.ScanDir(`output\goframe-latest-md1`, "*.md", true)
	if err != nil {
		t.Error(err)
	}
	mapData := LoadMapData()
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
				t.Log("not found")
				contents = strings.ReplaceAll(contents, u, "https://goframe.org"+u)
			}
		}
		contents = strings.ReplaceAll(contents, "；]", "]")
		contents = strings.ReplaceAll(contents, "；)", ")")
		contents = strings.ReplaceAll(contents, "- ```", "```")
		err := gfile.PutContents(file, contents)
		if err != nil {
			t.Error(err)
		}
	}
}

// LoadMapData description
//
// createTime: 2024-01-31 22:45:27
func LoadMapData() map[string]string {
	mapData := make(map[string]string)
	contents := gfile.GetContents(`D:\RAndD\go\project\doc2pdf\cmd\doc2pdf\output\goframe-latest-md-map\map.txt`)
	contents = strings.ReplaceAll(contents, "https://goframe.org", "")
	contents = strings.ReplaceAll(contents, "&src=contextnavpagetreemode", "")
	contents = strings.ReplaceAll(contents, "?src=contextnavpagetreemode", "")
	contents = strings.ReplaceAll(contents, "output/goframe-latest-md", "/docs")
	// 编写正则表达式
	regex := regexp.MustCompile(`/\d+-`)
	lines := strings.Split(contents, "\n")
	for _, line := range lines {
		line = regex.ReplaceAllString(line, "/")
		line = strings.TrimSuffix(line, ".md")
		if strings.Contains(line, "=>") {
			kv := strings.Split(line, "=>")
			fmt.Println(line)
			mapData[kv[0]] = kv[1]
		}
	}
	return mapData
}
