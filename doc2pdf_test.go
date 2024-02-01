package doc2pdf_test

import (
	"testing"

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
	doc2pdf.DownloadGoFrameLatest()
}

// TestDownloadGoFrameLatestMD description
//
// createTime: 2023-07-28 14:46:43
//
// author: hailaz
func TestDownloadGoFrameLatestMD(t *testing.T) {
	doc2pdf.DownloadConfluence("https://goframe.org/display/gf", "./output/goframe-latest", doc2pdf.DocDownloadModeMD)
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
