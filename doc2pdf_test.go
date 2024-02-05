package doc2pdf_test

import (
	"log"
	"testing"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/JohannesKaufmann/html-to-markdown/plugin"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/hailaz/doc2pdf"
)

// TestDownloadHailaz description
//
// createTime: 2023-07-28 14:46:43
//
// author: hailaz
func TestDownloadHailaz(t *testing.T) {
	doc2pdf.DownloadHailaz()
}

// TestDownloadGoFrameLatest description
//
// createTime: 2023-07-28 14:46:43
//
// author: hailaz
func TestDownloadGoFrameLatest(t *testing.T) {
	doc2pdf.DownloadGoFrameLatest(doc2pdf.DocDownloadModePDF)
}

// TestDownloadGoFrameLatestMD description
//
// createTime: 2023-07-28 14:46:43
//
// author: hailaz
func TestDownloadGoFrameLatestMD(t *testing.T) {
	doc2pdf.DownloadGoFrameLatest(doc2pdf.DocDownloadModeMD)
}

// TestDownloadGoFrameAll description
//
// createTime: 2023-07-28 15:29:07
//
// author: hailaz
func TestDownloadGoFrameAll(t *testing.T) {
	doc2pdf.DownloadGoFrameAll()
}

// TestDownloadRuanyifeng description
//
// createTime: 2023-12-07 16:03:21
//
// author: hailaz
func TestDownloadRuanyifeng(t *testing.T) {
	doc2pdf.DownloadRuanyifengWeekly()
}

var htmlpath = `output\goframe-latest-md-html\6-微服务开发\6-服务负载均衡1.html`

// TestH description
//
// createTime: 2024-02-05 18:17:01
func TestH(t *testing.T) {
	html := gfile.GetContents(htmlpath)
	converter := md.NewConverter("", true, nil)
	converter.Use(plugin.Strikethrough(""))
	converter.Use(doc2pdf.ConverterTable())
	markdown, err := converter.ConvertString(html)
	if err != nil {
		log.Fatal(err)
	}
	t.Log(markdown)
}
