# doc2pdf

将网页文档转换为单个 pdf 文件（带书签）。

## 支持文档类型

- [x] confluence
- [x] docusaurus

## 安装

```shell
go install github.com/hailaz/doc2pdf/cmd/doc2pdf@latest
```

## 使用

```shell
doc2pdf -h
# 示例
doc2pdf confluence --index="https://goframe.org/display/gf" --output="./output/temp"
```

### 环境准备

### ubuntu 无界面环境

#### 字体

##### 中文字体

```shell
sudo apt update
sudo apt install ttf-wqy-zenhei
fc-cache -f -v
```

#### 谷歌浏览器（选装）

```shell
sudo apt update
wget https://dl.google.com/linux/direct/google-chrome-stable_current_amd64.deb
apt install ./google-chrome-stable_current_amd64.deb
```

## 原理

先使用[rod](https://go-rod.github.io/i18n/zh-CN/#/)控制浏览器，将网页转换为 pdf 文件。

然后使用[unipdf](https://github.com/pdfcpu/pdfcpu)将 pdf 文件合并，最后再将目录插入到合并后的 pdf 文件中。
