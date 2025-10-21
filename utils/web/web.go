package web

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	pb "gopkg.in/cheggaaa/pb.v1"
)

var client = &http.Client{
	Timeout: 180 * time.Second, // 从60s延长到180s，避免下载超时
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		}, // 忽略证书验证
	},
}

// SetProxy 设置 HTTP 代理
// 参数:
//
//	p - 代理服务器地址，如果为空或 "none" 则不使用代理
func SetProxy(p string) {
	if p != "" && p != "none" {
		fmt.Println("设置代理服务器")
		proxyUrl, _ := url.Parse(p)
		client = &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxyUrl)}}
	} else {
		fmt.Println("没有代理服务器")
		client = &http.Client{}
	}
}

// Download 下载文件并显示进度条
// 参数:
//
//	url - 下载文件的 URL
//	target - 保存文件的目标路径
//
// 返回值:
//
//	bool - 下载成功返回 true，失败返回 false
func Download(url string, target string) bool {
	// 创建请求并设置 User-Agent 头，避免 418 错误
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("创建请求时出错", url, "-", err)
		return false
	}

	// 如果下载错误，返回418，华为云更新反爬虫策略，可能需要更新 User-Agent 字符串
	// HTTP 418 "I'm a teapot" 错误是华为云镜像的反爬虫机制：
	// - 服务器检测到请求缺少 User-Agent 头或使用默认的 User-Agent（如 Go 的 Go-http-client/1.1）
	// - 为了防止爬虫和自动化工具滥用，返回 418 状态码拒绝请求
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

	response, err := client.Do(req)
	if err != nil {
		fmt.Println("下载时出错", url, "-", err)
		return false
	}
	if response.StatusCode != 200 {
		fmt.Println("下载时状态错误", url, "-", response.StatusCode)
		return false
	}
	defer response.Body.Close()

	output, err := os.Create(target)
	if err != nil {
		fmt.Println("创建文件时出错", target, "-", err)
		return false
	}
	defer output.Close()

	// 创建一个进度条
	bar := pb.New(int(response.ContentLength)).SetUnits(pb.U_BYTES_DEC).SetRefreshRate(time.Millisecond * 10)
	// 显示下载速度
	bar.ShowSpeed = true

	// 显示剩余时间
	bar.ShowTimeLeft = true

	// 显示完成时间
	bar.ShowFinalTime = true

	bar.SetWidth(80)

	bar.Start()
	writer := io.MultiWriter(output, bar)
	_, err = io.Copy(writer, response.Body)
	if err != nil {
		fmt.Println("下载时出错", url, "-", err)
		return false
	}
	bar.Finish()

	return true
}

// GetJDK 下载指定版本的 JDK
// 参数:
//
//	download - 下载文件保存的目录路径
//	v - JDK 版本号
//	url - JDK 下载 URL
//
// 返回值:
//
//	string - 下载的文件路径，失败则返回空字符串
//	bool - 下载成功返回 true，失败返回 false
func GetJDK(download string, v string, url string) (string, bool) {
	if url == "" {
		fmt.Printf("JDK版本 %s 的下载无效\n", v)
		return "", false
	}

	// 从 URL 中提取完整的文件名
	urlParts := strings.Split(url, "/")
	originalFileName := urlParts[len(urlParts)-1]

	// 提取扩展名（处理 .tar.gz 等多重扩展名）
	ext := extractExtension(originalFileName)

	fileName := filepath.Join(download, fmt.Sprintf("%s%s", v, ext))
	os.Remove(fileName)

	if Download(url, fileName) {
		fmt.Println("下载完成")
		return fileName, true
	}
	return "", false

}

// 提取文件扩展名（如 .tar.gz  .zip）
func extractExtension(filename string) string {
	// 处理双扩展名
	if strings.HasSuffix(filename, ".tar.gz") {
		return ".tar.gz"
	}
	if strings.HasSuffix(filename, ".tar.xz") {
		return ".tar.xz"
	}
	if strings.HasSuffix(filename, ".tar.bz2") {
		return ".tar.bz2"
	}

	// 单扩展名
	return filepath.Ext(filename)
}
