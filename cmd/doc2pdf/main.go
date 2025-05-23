package main

import (
	"context"
	"log"

	"github.com/gogf/gf/v2/os/gcmd"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/hailaz/doc2pdf"
)

var (
	Main = &gcmd.Command{
		Name:        "doc2pdf",
		Brief:       "文档转换为pdf",
		Description: "",
	}

	pubArgs = []gcmd.Argument{
		{
			Name:  "index",
			Brief: "confluence文档地址, https://goframe.org/display/gf",
		},
		{
			Name:  "output",
			Brief: "输出目录, ./output/temp",
		},
		{
			Name:  "mode",
			Brief: "下载模式，pdf或md，默认pdf",
			Short: "m",
		},
	}

	confluence = &gcmd.Command{
		Name:        "confluence",
		Brief:       "confluence文档转换为pdf，doc2pdf confluence -h",
		Description: "doc2pdf confluence --index=\"https://goframe.org/display/gf\" --output=\"./output/temp\"",
		Arguments:   pubArgs,
		Func:        confluenceFunc,
	}

	goframeOld = &gcmd.Command{
		Name:        "gf",
		Brief:       "GoFrame文档转换为pdf，doc2pdf gf -h",
		Description: "doc2pdf gf",
		Arguments: append(pubArgs, gcmd.Argument{
			Name:  "version",
			Brief: "下载版本",
			Short: "v",
		}),
		Func: goframeFuncOld,
	}

	goframe = &gcmd.Command{
		Name:        "gf",
		Brief:       "GoFrame文档转换为pdf，doc2pdf gf -h",
		Description: "doc2pdf gf",
		Arguments:   pubArgs,
		Func:        goframeFunc,
	}

	docusaurus = &gcmd.Command{
		Name:        "docusaurus",
		Brief:       "docusaurus文档转换为pdf，doc2pdf docusaurus -h",
		Description: "doc2pdf docusaurus --index=\"http://www.hailaz.cn/docs/learn/index\" --output=\"./output/hailaz-learn\"",
		Arguments:   pubArgs,
		Func:        docusaurusFunc,
	}
)

// main description
//
// createTime: 2023-07-26 11:42:19
//
// author: hailaz
func main() {
	// doc := NewDocDownload("https://goframe.org/pages/viewpage.action?pageId=7296490", "./output/hailaz")

	// doc2pdf.DownloadGoFrameAll()
	// doc2pdf.DownloadGoFrameLatest()
	err := Main.AddCommand(goframe, confluence, docusaurus)
	if err != nil {
		panic(err)
	}
	Main.Run(gctx.New())
}

// confluenceFunc description
func confluenceFunc(ctx context.Context, parser *gcmd.Parser) (err error) {
	// go run main.go confluence --index="https://goframe.org/display/gf" --output="./output/temp"
	index := parser.GetOpt("index")
	output := parser.GetOpt("output")
	inMode := parser.GetOpt("mode", doc2pdf.DocDownloadModePDF)
	log.Printf("index: %v, output: %v, mode: %v", index, output, inMode)
	if index == nil || output == nil {
		log.Printf("index or output is nil")
		return
	}

	doc2pdf.DownloadConfluence(index.String(), output.String(), inMode.String(), false)
	return
}

// goframeFunc description
func goframeFuncOld(ctx context.Context, parser *gcmd.Parser) (err error) {
	version := parser.GetOpt("version")
	inMode := parser.GetOpt("mode", doc2pdf.DocDownloadModePDF)
	log.Printf("version: %v, mode: %v", version, inMode)
	if version != nil {
		switch version.String() {
		case "all":
			doc2pdf.DownloadGoFrameAll(inMode.String())
		default:
			doc2pdf.DownloadGoFrameWithVersion(version.String(), inMode.String())
		}
	} else {
		doc2pdf.DownloadGoFrameLatest(inMode.String())
	}
	return
}

// goframeFunc description
func goframeFunc(ctx context.Context, parser *gcmd.Parser) (err error) {
	index := parser.GetOpt("index")
	log.Printf("index: %v", index)
	doc2pdf.DownloadGoFrame(index.String())
	return
}

// docusaurusFunc description
func docusaurusFunc(ctx context.Context, parser *gcmd.Parser) (err error) {
	// go run main.go docusaurus --index="http://www.hailaz.cn/docs/learn/index" --output="./output/hailaz-learn"
	index := parser.GetOpt("index")
	output := parser.GetOpt("output")
	log.Printf("index: %v, output: %v", index, output)
	if index == nil || output == nil {
		log.Printf("index or output is nil")
		return
	}

	doc2pdf.DownloadDocusaurus(index.String(), output.String())
	return
}
