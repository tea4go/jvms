package jdk

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	logs "github.com/tea4go/gh/log4go"
	"golang.org/x/net/html"
)

// 清华大学开源软件镜像站
type TWebTuna struct {
	BaseURL string // 入口地址
}

// 兰州大学开源软件镜像站
type TWebLzu struct {
	BaseURL string // 入口地址
}

// 华为软件镜像站
type TWebHuawei struct {
	BaseURL string // 入口地址
}

// injdk 软件镜像站
type TWebInjdk struct {
	BaseURL string // 入口地址
}

// Azul 软件镜像站
type TWebAzul struct {
	BaseURL string // 入口地址
}

// Adoptium 软件镜像站
type TWebAdoptium struct {
	BaseURL string // 入口地址
}

// TWebFileInfo 表示从 HTML 表格中解析出的文件信息
type TWebFileInfo struct {
	Name         string // 文件名
	LastModified string // 最后修改时间
	Size         string // 文件大小
}

// generateUserAgent 生成随机User-Agent
func generateUserAgent() string {
	browsers := []string{
		"Mozilla/5.0 (Windows NT 11.0; Win64; x64)",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 26_15_7)",
		"Mozilla/5.0 (X11; Linux x86_64)",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 15_0 like Mac OS X)",
		"Mozilla/5.0 (Android 24; Mobile; rv:68.0)",
	}

	engines := []string{
		"AppleWebKit/537.36 (KHTML, like Gecko)",
		"AppleWebKit/605.1.15 (KHTML, like Gecko)",
	}

	chromeVersions := []string{
		"Chrome/98.0.4758.102",
		"Chrome/99.0.4844.51",
		"Chrome/100.0.4896.60",
		"Chrome/101.0.4951.41",
		"Chrome/102.0.5005.61",
	}

	safariVersions := []string{
		"Safari/537.36",
		"Safari/605.1.15",
		"Safari/137.36",
		"Safari/905.1.15",
	}

	firefoxVersions := []string{
		"Gecko/20100101 Firefox/103.0",
		"Gecko/20100101 Firefox/102.0",
		"Gecko/20100101 Firefox/101.0",
		"Gecko/20100101 Firefox/98.0",
		"Gecko/20100101 Firefox/99.0",
		"Gecko/20100101 Firefox/100.0",
	}

	mobileSuffixes := []string{
		" Mobile/12E168",
		" Mobile/13E178",
		" Mobile/14E158",
		" Mobile/15E148",
		" Mobile/16E145",
		" Mobile/17E142",
		" Mobile/18E138",
	}

	// 随机选择浏览器类型
	browser := browsers[rand.Intn(len(browsers))]
	engine := engines[rand.Intn(len(engines))]

	// 根据浏览器类型选择不同的版本
	var version string
	switch rand.Intn(3) {
	case 0:
		version = chromeVersions[rand.Intn(len(chromeVersions))]
	case 1:
		version = safariVersions[rand.Intn(len(safariVersions))]
	case 2:
		version = firefoxVersions[rand.Intn(len(firefoxVersions))]
	}

	mobileSuffix := mobileSuffixes[rand.Intn(len(mobileSuffixes))]

	return fmt.Sprintf("%s %s %s%s%d", browser, engine, version, mobileSuffix, rand.Intn(1000))
}

// mapOSToGOOS 将操作系统目录名映射为 GOOS 标准名称
// 参数:
//   - osDir: 操作系统目录名（如 "windows/", "linux/"）
//
// 返回:
//   - string: GOOS 标准名称
func mapOSToGOOS(osDir string) string {
	osDir = strings.TrimSuffix(osDir, "/")
	osMap := map[string]string{
		"windows":      "windows",
		"linux":        "linux",
		"mac":          "darwin",
		"macos":        "darwin",
		"osx":          "darwin",
		"alpine-linux": "linux",
		"aix":          "aix",
		"solaris":      "solaris",
	}
	if goos, ok := osMap[osDir]; ok {
		return goos
	}
	return osDir
}

// mapArchToGOARCH 将架构目录名映射为 GOARCH 标准名称
// 参数:
//   - archDir: 架构目录名（如 "x64/", "aarch64/"）
//
// 返回:
//   - string: GOARCH 标准名称
func mapArchToGOARCH(archDir string) string {
	archDir = strings.TrimSuffix(archDir, "/")
	archMap := map[string]string{
		"x64":     "amd64",
		"x32":     "386",
		"aarch64": "arm64",
		"arm":     "arm",
		"ppc64":   "ppc64",
		"ppc64le": "ppc64le",
		"s390x":   "s390x",
		"riscv64": "riscv64",
		"sparcv9": "sparc64",
	}
	if goarch, ok := archMap[archDir]; ok {
		return goarch
	}
	return archDir
}

// extractVersion 从文件名中提取版本号
// 参数:
//   - filename: 文件名
//
// 返回:
//   - string: 版本号
func extractVersion(filename string) string {
	// 匹配版本号模式，例如 "8u462b08", "25", "21.0.1", "17.0.2+8" 等
	patterns := []string{
		`(\d+u\d+b\d+)`,        // 8u462b08
		`(\d+\.\d+\.\d+\+\d+)`, // 17.0.2+8
		`(\d+\.\d+\.\d+)`,      // 21.0.1
		`(\d+)`,                // 25
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		if matches := re.FindStringSubmatch(filename); len(matches) > 1 {
			return matches[1]
		}
	}

	return "unknown"
}

// parseTime 解析时间字符串并转换为标准格式
// 参数:
//   - timeStr: 时间字符串（如 "22 Jul 2025 15:07:40 +0000"）
//
// 返回:
//   - string: 格式化后的时间字符串（"2006-01-02 15:04:05"）
func parseTime(timeStr string) string {
	// 尝试解析多种时间格式
	formats := []string{
		"02 Jan 2006 15:04:05 -0700", // 22 Jul 2025 15:07:40 +0000
		"2006-Jan-02 15:04",
		"2006-01-02 15:04:05", // 已经是目标格式
		"2006-01-02 15:04",    // 已经是目标格式
		time.RFC3339,          // 2006-01-02T15:04:05Z07:00
	}
	for _, format := range formats {
		if t, err := time.Parse(format, timeStr); err == nil {
			// 转换为目标格式
			return t.Format("2006-01-02 15:04")
		}
	}

	// 如果无法解析，返回原始字符串
	return timeStr
}

// fetchHTML 从指定 URL 获取 HTML 内容
// 参数:
//   - url: 要获取的网页地址
//
// 返回:
//   - string: HTML 内容
//   - error: 错误信息
func fetchHTML(url string) (string, error) {
	// 创建请求并设置 User-Agent（某些网站可能检查 User-Agent）
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("创建请求失败，%v", err)
	}
	//req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	UserAgent := generateUserAgent()
	req.Header.Set("User-Agent", UserAgent)

	// 忽略证书验证
	customTransport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: customTransport,
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("获取网页失败，%v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		logs.Warning("%s <== %s", resp.Status, UserAgent)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取网页内容失败，%v", err)
	}

	return string(body), nil
}

// parseLinks 解析 HTML 中的所有 <a> 标签链接
// 参数:
//   - htmlContent: HTML 内容字符串
//   - pattern: 正则表达式模式，用于过滤链接
//
// 返回:
//   - []string: 匹配的链接列表
//   - error: 错误信息
func parseLinks(htmlContent string, pattern string) ([]string, error) {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return nil, fmt.Errorf("解析 HTML 失败，%v", err)
	}

	var links []string
	var re *regexp.Regexp
	if pattern != "" {
		re = regexp.MustCompile(pattern)
	}

	// 递归遍历 HTML 节点
	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, attr := range n.Attr {
				if attr.Key == "href" {
					href := attr.Val
					// 如果有正则模式，进行匹配
					if re == nil || re.MatchString(href) {
						links = append(links, href)
					}
					break
				}
			}
		}
		// 遍历子节点
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}

	traverse(doc)
	return links, nil
}

// parseFileTable 解析 HTML 表格中的文件信息（包括文件名、大小、修改时间）
// 参数:
//   - htmlContent: HTML 内容字符串
//   - pattern: 正则表达式模式，用于过滤文件名
//
// 返回:
//   - []TWebFileInfo: 文件信息列表
//   - error: 错误信息
func parseFileTable(htmlContent string, pattern string) ([]TWebFileInfo, error) {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return nil, fmt.Errorf("解析 HTML 失败，%v", err)
	}

	var files []TWebFileInfo
	var re *regexp.Regexp
	if pattern != "" {
		re = regexp.MustCompile(pattern)
	}

	// 递归遍历找到表格行
	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		// 找到 <tr> 标签
		if n.Type == html.ElementNode && n.Data == "tr" {
			var fileName, fileSize, fileDate string
			colIndex := 0

			// 遍历表格列 <td>
			for td := n.FirstChild; td != nil; td = td.NextSibling {
				if td.Type == html.ElementNode && td.Data == "td" {
					// 获取列的文本内容
					text := getTextContent(td)

					// 第一列通常是文件名（在 <a> 标签中）
					if colIndex == 0 {
						for a := td.FirstChild; a != nil; a = a.NextSibling {
							if a.Type == html.ElementNode && a.Data == "a" {
								for _, attr := range a.Attr {
									if attr.Key == "href" {
										fileName = attr.Val
										break
									}
								}
							}
						}
					} else if colIndex == 1 {
						// 第二列是文件大小
						fileSize = strings.TrimSpace(text)
					} else if colIndex == 2 {
						// 第三列是修改时间
						fileDate = strings.TrimSpace(text)
					}
					colIndex++
				}
			}

			// 如果找到文件名且匹配模式，添加到列表
			if fileName != "" && (re == nil || re.MatchString(fileName)) {
				files = append(files, TWebFileInfo{
					Name:         fileName,
					LastModified: fileDate,
					Size:         fileSize,
				})
			}
		}

		// 遍历子节点
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}

	traverse(doc)
	return files, nil
}

// parseFileListHuawei 解析华为云的 HTML 文件列表
// 华为云的 HTML 结构:
//   - 使用简单的 <pre> 标签格式
//   - 格式: <a href="filename"...>filename</a> date time size
//   - 注意：<a> 标签可能跨越多行
//
// 参数:
//   - htmlContent: HTML 内容字符串
//   - pattern: 正则表达式模式，用于过滤文件名
//
// 返回:
//   - []TWebFileInfo: 文件信息列表
//   - error: 错误信息
func parseFileListHuawei(htmlContent string, pattern string) ([]TWebFileInfo, error) {
	var files []TWebFileInfo
	var re *regexp.Regexp
	if pattern != "" {
		re = regexp.MustCompile(pattern)
	}

	// 移除换行符，将多行的 <a> 标签合并
	htmlContent = strings.ReplaceAll(htmlContent, "\n", " ")

	// 正则表达式匹配文件行
	// 格式: <a href="filename"...>filename</a> date time size
	lineRe := regexp.MustCompile(`<a\s+href="([^"]+)"[^>]*>([^<]+)</a>\s+(\d{2}-\w{3}-\d{4})\s+(\d{2}:\d{2})\s+([^\s<]+(?:\s+[^\s<]+)?)`)

	matches := lineRe.FindAllStringSubmatch(htmlContent, -1)
	for _, match := range matches {
		if len(match) >= 6 {
			fileName := match[1]
			date := match[3]
			time := match[4]
			size := strings.TrimSpace(match[5])

			// 如果匹配模式，添加到列表
			if re == nil || re.MatchString(fileName) {
				files = append(files, TWebFileInfo{
					Name:         fileName,
					LastModified: date + " " + time,
					Size:         size,
				})
			}
		}
	}

	return files, nil
}

// parseFileTableInjdk 解析 InJDK 网站的 HTML 表格中的文件信息
// InJDK 的 HTML 结构:
//   - 文件名在 <span class="name"> 中
//   - 文件大小在 <td class="size" data-size="字节数"> 中
//   - 时间在 <time datetime="ISO格式"> 中
//
// 参数:
//   - htmlContent: HTML 内容字符串
//   - pattern: 正则表达式模式，用于过滤文件名
//
// 返回:
//   - []TWebFileInfo: 文件信息列表
//   - error: 错误信息
func parseFileTableInjdk(htmlContent string, pattern string) ([]TWebFileInfo, error) {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return nil, fmt.Errorf("解析 HTML 失败，%v", err)
	}

	var files []TWebFileInfo
	var re *regexp.Regexp
	if pattern != "" {
		re = regexp.MustCompile(pattern)
	}

	// 递归遍历找到表格行
	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		// 找到 <tr> 标签
		if n.Type == html.ElementNode && n.Data == "tr" {
			var fileName, fileSize, fileDate string

			// 遍历表格列 <td>
			for td := n.FirstChild; td != nil; td = td.NextSibling {
				if td.Type == html.ElementNode && td.Data == "td" {
					// 查找文件名（在 <span class="name"> 中）
					var findName func(*html.Node)
					findName = func(node *html.Node) {
						if node.Type == html.ElementNode && node.Data == "span" {
							for _, attr := range node.Attr {
								if attr.Key == "class" && attr.Val == "name" {
									fileName = getTextContent(node)
									return
								}
							}
						}
						for c := node.FirstChild; c != nil; c = c.NextSibling {
							findName(c)
						}
					}
					findName(td)

					// 查找文件大小（在 data-size 属性中）
					for _, attr := range td.Attr {
						if attr.Key == "class" && strings.Contains(attr.Val, "size") {
							// 查找 data-size 属性
							for _, sizeAttr := range td.Attr {
								if sizeAttr.Key == "data-size" {
									fileSize = sizeAttr.Val
									break
								}
							}
						}
					}

					// 查找时间（在 <time datetime=""> 中）
					var findTime func(*html.Node)
					findTime = func(node *html.Node) {
						if node.Type == html.ElementNode && node.Data == "time" {
							for _, attr := range node.Attr {
								if attr.Key == "datetime" {
									fileDate = attr.Val
									return
								}
							}
						}
						for c := node.FirstChild; c != nil; c = c.NextSibling {
							findTime(c)
						}
					}
					findTime(td)
				}
			}

			// 如果找到文件名且匹配模式，添加到列表
			if fileName != "" && (re == nil || re.MatchString(fileName)) {
				files = append(files, TWebFileInfo{
					Name:         fileName,
					LastModified: fileDate,
					Size:         fileSize,
				})
			}
		}

		// 遍历子节点
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}

	traverse(doc)
	return files, nil
}

// getTextContent 获取节点的文本内容
// 参数:
//   - n: HTML 节点
//
// 返回:
//   - string: 文本内容
func getTextContent(n *html.Node) string {
	if n.Type == html.TextNode {
		return n.Data
	}
	var text string
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		text += getTextContent(c)
	}
	return text
}

// getVerDirs 获取所有版本目录（如 25.0.3/, 24/, 等）
// 返回:
//   - []string: 版本目录列表
//   - error: 错误信息
func getVerDirs(url string) ([]string, error) {
	logs.Debug("正在获取版本目录列表...")
	htmlContent, err := fetchHTML(url)
	if err != nil {
		return nil, err
	}

	// 尝试使用 HTML 解析器
	links, err := parseLinks(htmlContent, `^(?:\./)?(\d+(?:\.\d+)*)/$`)
	if err != nil {
		return nil, err
	}

	// 如果 HTML 解析器没有找到链接，使用正则表达式直接解析
	if len(links) == 0 {
		// 使用正则表达式直接从 HTML 中提取版本目录
		re := regexp.MustCompile(`<a\s+href="((?:\./)?(\d+(?:\.\d+)*)/)"`)
		matches := re.FindAllStringSubmatch(htmlContent, -1)
		for _, match := range matches {
			if len(match) > 1 {
				links = append(links, match[1])
			}
		}
	}

	// 清理链接，移除 "./" 前缀
	var cleanLinks []string
	for _, link := range links {
		cleanLink := strings.TrimPrefix(link, "./")
		cleanLinks = append(cleanLinks, cleanLink)
	}

	logs.Info("找到 %d 个JDK下载地址", len(cleanLinks))
	return cleanLinks, nil
}

// getJDKDirectory 获取 jdk/ 目录
// 参数:
//   - versionURL: 版本目录的 URL
//
// 返回:
//   - string: jdk 目录路径
//   - error: 错误信息
func getJDKDirectory(versionURL string) (string, error) {
	htmlContent, err := fetchHTML(versionURL)
	if err != nil {
		return "", err
	}

	links, err := parseLinks(htmlContent, `^jdk/$`)
	if err != nil {
		return "", err
	}

	if len(links) == 0 {
		return "", fmt.Errorf("未找到 jdk 目录")
	}

	return links[0], nil
}

// getArchDirs 获取所有架构目录（如 x64/, aarch64/, 等）
// 参数:
//   - jdkURL: jdk 目录的 URL
//
// 返回:
//   - []string: 架构目录列表
//   - error: 错误信息
func getArchDirs(jdkURL string) ([]string, error) {
	htmlContent, err := fetchHTML(jdkURL)
	if err != nil {
		return nil, err
	}

	// 匹配以 / 结尾的目录
	links, err := parseLinks(htmlContent, `^[^/]+/$`)
	if err != nil {
		return nil, err
	}

	// 过滤掉父目录链接
	var archs []string
	for _, link := range links {
		if link != "../" {
			archs = append(archs, link)
		}
	}

	return archs, nil
}

func formatFileSize(filesize string) string {
	// 将字符串转为 float64
	bytes, err := strconv.ParseFloat(filesize, 64)
	if err != nil {
		return "0"
	}
	units := []string{"B", "KiB", "MiB", "GiB", "TiB", "PiB", "EiB"}

	if bytes < 1 {
		return "0 B"
	}

	// 计算适合的单位
	base := 1024.0
	exp := int(math.Log(bytes) / math.Log(base))
	if exp >= len(units) {
		exp = len(units) - 1
	}

	// 计算值并格式化
	value := bytes / math.Pow(base, float64(exp))

	// 根据值的大小决定小数位数
	var format string
	if value < 10 {
		format = "%.2f %s"
	} else if value < 100 {
		format = "%.1f %s"
	} else {
		format = "%.0f %s"
	}

	return fmt.Sprintf(format, value, units[exp])
}

// getOSDirs 获取所有操作系统目录（如 windows/, linux/, mac/, 等）
// 参数:
//   - archURL: 架构目录的 URL
//
// 返回:
//   - []string: 操作系统目录列表
//   - error: 错误信息
func getOSDirs(archURL string) ([]string, error) {
	htmlContent, err := fetchHTML(archURL)
	if err != nil {
		return nil, err
	}

	links, err := parseLinks(htmlContent, `^[^/]+/$`)
	if err != nil {
		return nil, err
	}

	// 过滤掉父目录链接
	var osDirs []string
	for _, link := range links {
		if link != "../" {
			osDirs = append(osDirs, link)
		}
	}

	return osDirs, nil
}

// saveToJSON 将下载列表保存为 JSON 文件
// 参数:
//   - downloads: JDK 下载条目列表
//   - filename: 输出文件名
//
// 返回:
//   - error: 错误信息
func saveToJSON(downloads []TOpenJDK, filename string) error {
	// 按版本和文件名排序
	sort.Slice(downloads, func(i, j int) bool {
		if downloads[i].Version != downloads[j].Version {
			return downloads[i].Version > downloads[j].Version
		}
		return downloads[i].Filename < downloads[j].Filename
	})

	// 格式化 JSON，使用缩进
	data, err := json.MarshalIndent(downloads, "", "  ")
	if err != nil {
		return fmt.Errorf("JSON 序列化失败，%v", err)
	}

	// 写入文件
	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		return fmt.Errorf("写入文件失败，%v", err)
	}

	logs.Debug("成功保存 %d 个下载地址到 %s", len(downloads), filename)
	return nil
}

// ParseURL 爬取所有 清华大学 JDK 下载地址
// 目录层次:
//   - Adoptium/25/jdk/x64/windows/OpenJDK25U-jdk_x64_windows_hotspot_25_36.zip
//
// 返回:
//   - []TOpenJDK: 所有 JDK 下载条目
//   - error: 错误信息
func (s *TWebTuna) ParseURL() ([]TOpenJDK, error) {
	var allDownloads []TOpenJDK

	// 1. 获取版本目录
	versions, err := getVerDirs(s.BaseURL)
	if err != nil {
		return nil, err
	}

	// 2. 遍历每个版本
	for _, version := range versions {
		versionURL := s.BaseURL + version
		// 去除版本号中的 '/' 字符用于显示
		versionDisplay := strings.TrimSuffix(version, "/")
		logs.Debug("==================================================")
		logs.Debug("-= 处理 JDK %s 版本 =-", versionDisplay)
		logs.Debug("==================================================")

		// 3. 获取 jdk 目录
		jdkDir, err := getJDKDirectory(versionURL)
		if err != nil {
			logs.Warning("  获取jdk目录失败，%v", err)
			continue
		}
		jdkURL := versionURL + jdkDir

		// 4. 获取架构目录
		archs, err := getArchDirs(jdkURL)
		if err != nil {
			logs.Warning("  获取架构目录失败，%v", err)
			continue
		}

		// 5. 遍历每个架构（只处理 amd64 和 arm64）
		for _, arch := range archs {
			// 过滤架构：只处理 x64(amd64) 和 aarch64(arm64)
			archName := strings.TrimSuffix(arch, "/")
			if archName != "x64" && archName != "aarch64" {
				continue
			}

			archURL := jdkURL + arch

			// 6. 获取操作系统目录
			osDirs, err := getOSDirs(archURL)
			if err != nil {
				logs.Warning("获取操作系统目录失败，%v", err)
				continue
			}

			// 7. 遍历每个操作系统（只处理 windows, darwin, linux）
			for _, osDir := range osDirs {
				// 过滤操作系统：只处理 windows, mac(darwin), linux
				goos := mapOSToGOOS(osDir)
				if goos != "windows" && goos != "darwin" && goos != "linux" {
					continue
				}

				osURL := archURL + osDir

				// 8. 获取 JDK 文件
				downloads, err := s.GetJDKFiles(osURL, osDir, arch)
				if err != nil {
					logs.Warning("获取文件失败，%v", err)
					continue
				}

				// 打印找到的 OpenJDK 信息
				if len(downloads) > 0 {
					for _, jdk := range downloads {
						logs.Debug("%s", jdk.String())
					}
				}
				allDownloads = append(allDownloads, downloads...)
			}
		}
	}

	return allDownloads, nil
}

// getJDKFiles 获取指定 URL 中的所有 JDK 文件下载地址及详细信息
// 参数:
//   - fileURL: 文件列表页面的 URL
//   - osDir: 操作系统目录名
//   - archDir: 架构目录名
//
// 返回:
//   - []TOpenJDK: JDK 下载条目列表
//   - error: 错误信息
func (s *TWebTuna) GetJDKFiles(fileURL string, osDir string, archDir string) ([]TOpenJDK, error) {
	htmlContent, err := fetchHTML(fileURL)
	if err != nil {
		return nil, err
	}

	// 匹配 .zip 和 .tar.gz 文件
	files, err := parseFileTable(htmlContent, `\.(zip|tar\.gz)$`)
	if err != nil {
		return nil, err
	}

	var downloads []TOpenJDK
	for _, file := range files {
		// 从文件名中提取版本号
		version := extractVersion(file.Name)

		// 格式化时间
		formattedTime := parseTime(file.LastModified)

		downloads = append(downloads, TOpenJDK{
			Version:      version,
			Filename:     file.Name,
			URL:          fileURL + file.Name,
			Size:         file.Size,
			LastModified: formattedTime,
			GOOS:         mapOSToGOOS(osDir),
			GOARCH:       mapArchToGOARCH(archDir),
		})
	}

	return downloads, nil
}

// ParseURL 爬取所有 兰州大学 JDK 下载地址
// 目录层次:
//   - /openjdk/11.0.2/openjdk-11.0.2_windows-x64_bin.zip
//
// 返回:
//   - []TOpenJDK: 所有 JDK 下载条目
//   - error: 错误信息
func (s *TWebLzu) ParseURL() ([]TOpenJDK, error) {
	var allDownloads []TOpenJDK

	// 1. 获取版本目录
	versions, err := getVerDirs(s.BaseURL)
	if err != nil {
		return nil, err
	}

	// 2. 遍历每个版本
	for _, version := range versions {
		if !strings.Contains(version, "20.") {
			continue
		}
		versionURL := s.BaseURL + version
		// 去除版本号中的 '/' 字符用于显示
		versionDisplay := strings.TrimSuffix(version, "/")
		logs.Debug("==================================================")
		logs.Debug("-= 处理 JDK %s 版本 =-", versionDisplay)
		logs.Debug("==================================================")

		// 8. 获取 JDK 文件
		downloads, err := s.GetJDKFiles(versionURL)
		if err != nil {
			logs.Warning("获取文件失败，%v", err)
			continue
		}

		// 打印找到的 OpenJDK 信息
		if len(downloads) > 0 {
			for _, jdk := range downloads {
				logs.Debug("%s", jdk.String())
			}
		}
		allDownloads = append(allDownloads, downloads...)

	}

	return allDownloads, nil
}

func (s *TWebLzu) GetJDKFiles(fileURL string) ([]TOpenJDK, error) {
	htmlContent, err := fetchHTML(fileURL)
	if err != nil {
		return nil, err
	}

	// 匹配 .zip 和 .tar.gz 文件
	files, err := parseFileTable(htmlContent, `\.(zip|tar\.gz)$`)
	if err != nil {
		return nil, err
	}

	var downloads []TOpenJDK
	for _, file := range files {
		// 从文件名中提取版本号
		version := extractVersion(file.Name)

		// 格式化时间
		formattedTime := parseTime(file.LastModified)

		goos, goarch, _, err := s.ParseWebFileName(file.Name)
		if err != nil {
			logs.Warning(err)
			continue
		}

		downloads = append(downloads, TOpenJDK{
			Version:      version,
			Filename:     file.Name,
			URL:          fileURL + file.Name,
			Size:         formatFileSize(file.Size),
			LastModified: formattedTime,
			GOOS:         goos,
			GOARCH:       goarch,
		})
	}

	return downloads, nil
}

// 输入:openjdk-10.0.1_windows-x64_bin.tar.gz
// 返回:Goos,Arch,Version
func (s *TWebLzu) ParseWebFileName(filename string) (string, string, string, error) {
	filename = strings.TrimSpace(filename)
	if filename == "" {
		return "", "", "", fmt.Errorf("文件名不能为空")
	}

	// 解析文件名获取版本、GOOS和GOARCH
	filenameParts := strings.Split(filename, "_")
	if len(filenameParts) < 3 {
		return "", "", "", fmt.Errorf("文件名格式错误(%s)", filename)
	}

	version := strings.TrimPrefix(filenameParts[0], "openjdk-")
	filenameParts1 := strings.Split(filenameParts[1], "-")
	var goos, goarch string
	if len(filenameParts1) > 1 {
		goos = filenameParts1[0]
		goarch = filenameParts1[1]
	}

	// 标准化GOOS值
	goos = strings.ReplaceAll(goos, "osx", "darwin")
	goos = strings.ReplaceAll(goos, "macos", "darwin")
	switch goos {
	case "linux", "darwin", "windows":
		// 已经是标准值
	default:
		return "", "", "", fmt.Errorf("未知操作系统(%s)", goos)
	}

	// 标准化GOARCH值
	goarch = strings.ReplaceAll(goarch, "aarch64", "arm64")
	switch goarch {
	case "x64", "amd64", "arm64", "aarch64":
		if goarch == "x64" {
			goarch = "amd64" // Go标准中使用amd64而不是x64
		}
	default:
		return "", "", "", fmt.Errorf("未知系统架构(%s)", goarch)
	}

	return goos, goarch, version, nil
}

// ParseURL 爬取所有 InJDK JDK 下载地址
// 目录层次:
//   - /openjdk/11.0.2/openjdk-11.0.2_windows-x64_bin.zip
//
// 返回:
//   - []TOpenJDK: 所有 JDK 下载条目
//   - error: 错误信息
func (s *TWebInjdk) ParseURL() ([]TOpenJDK, error) {
	var allDownloads []TOpenJDK

	// 1. 获取版本目录
	versions, err := getVerDirs(s.BaseURL)
	if err != nil {
		return nil, err
	}

	// 2. 遍历每个版本
	for _, version := range versions {
		versionURL := s.BaseURL + version
		// 去除版本号中的 '/' 字符用于显示
		versionDisplay := strings.TrimSuffix(version, "/")
		logs.Debug("==================================================")
		logs.Debug("-= 处理 JDK %s 版本 =-", versionDisplay)
		logs.Debug("==================================================")

		// 3. 获取 JDK 文件
		downloads, err := s.GetJDKFiles(versionURL)
		if err != nil {
			logs.Warning("获取文件失败，%v", err)
			continue
		}

		// 打印找到的 OpenJDK 信息
		if len(downloads) > 0 {
			for _, jdk := range downloads {
				logs.Debug("%s", jdk.String())
			}
		}
		allDownloads = append(allDownloads, downloads...)
	}

	return allDownloads, nil
}

// GetJDKFiles 获取指定 URL 中的所有 JDK 文件下载地址及详细信息
// 参数:
//   - fileURL: 文件列表页面的 URL
//
// 返回:
//   - []TOpenJDK: JDK 下载条目列表
//   - error: 错误信息
func (s *TWebInjdk) GetJDKFiles(fileURL string) ([]TOpenJDK, error) {
	htmlContent, err := fetchHTML(fileURL)
	if err != nil {
		return nil, err
	}

	// 使用 InJDK 专用的解析函数，匹配 .zip 和 .tar.gz 文件
	files, err := parseFileTableInjdk(htmlContent, `\.(zip|tar\.gz)$`)
	if err != nil {
		return nil, err
	}

	var downloads []TOpenJDK
	for _, file := range files {
		// 从文件名中提取版本号
		version := extractVersion(file.Name)

		// 格式化时间（InJDK 使用 ISO 8601 格式）
		formattedTime := parseTime(file.LastModified)

		goos, goarch, _, err := s.ParseWebFileName(file.Name)
		if err != nil {
			logs.Warning(err)
			continue
		}

		downloads = append(downloads, TOpenJDK{
			Version:      version,
			Filename:     file.Name,
			URL:          fileURL + file.Name,
			Size:         formatFileSize(file.Size),
			LastModified: formattedTime,
			GOOS:         goos,
			GOARCH:       goarch,
		})
	}

	return downloads, nil
}

// ParseWebFileName 解析文件名获取GOOS、GOARCH和版本信息
// 支持两种格式:
//   - 标准格式: openjdk-10.0.1_windows-x64_bin.tar.gz
//   - JDK8格式: openjdk-8u43-linux-x64.tar.gz
//
// 返回:Goos,Arch,Version,error
func (s *TWebInjdk) ParseWebFileName(filename string) (string, string, string, error) {
	filename = strings.TrimSpace(filename)
	if filename == "" {
		return "", "", "", fmt.Errorf("文件名不能为空")
	}

	// 跳过源代码文件（如 openjdk-11+28_src.zip, openjdk-8u41-src-b04-14_jan_2020.zip）
	if strings.Contains(filename, "_src.") || strings.Contains(filename, "-src-") {
		return "", "", "", fmt.Errorf("跳过源代码文件(%s)", filename)
	}

	// 移除文件扩展名
	nameWithoutExt := strings.TrimSuffix(filename, ".tar.gz")
	nameWithoutExt = strings.TrimSuffix(nameWithoutExt, ".zip")

	// 检查是否是 JDK 8 格式（使用 - 分隔）
	// 格式1: openjdk-8u43-linux-x64
	// 格式2: openjdk-8u41-b04-windows-i586-14_jan_2020
	if strings.Contains(nameWithoutExt, "openjdk-8") {
		parts := strings.Split(nameWithoutExt, "-")
		if len(parts) >= 4 {
			var version, goos, goarch string

			// 格式2: openjdk-8u41-b04-windows-i586-14_jan_2020
			if len(parts) >= 5 && strings.HasPrefix(parts[2], "b") {
				// parts: ["openjdk", "8u41", "b04", "windows", "i586", "14_jan_2020"]
				version = parts[1]
				goos = parts[3]
				goarch = parts[4]
			} else {
				// 格式1: openjdk-8u43-linux-x64
				// parts: ["openjdk", "8u43", "linux", "x64"]
				version = parts[1]
				goos = parts[2]
				goarch = parts[3]
			}

			// 标准化GOOS值
			goos = strings.ReplaceAll(goos, "osx", "darwin")
			goos = strings.ReplaceAll(goos, "macos", "darwin")
			switch goos {
			case "linux", "darwin", "windows":
				// 已经是标准值
			default:
				return "", "", "", fmt.Errorf("未知操作系统(%s)", goos)
			}

			// 标准化GOARCH值
			goarch = strings.ReplaceAll(goarch, "aarch64", "arm64")
			goarch = strings.ReplaceAll(goarch, "i586", "386")
			switch goarch {
			case "x64", "amd64", "arm64", "aarch64", "386":
				if goarch == "x64" {
					goarch = "amd64"
				}
			default:
				return "", "", "", fmt.Errorf("未知系统架构(%s)", goarch)
			}

			return goos, goarch, version, nil
		}
	}

	// 标准格式（使用 _ 分隔）: openjdk-10.0.1_windows-x64_bin
	filenameParts := strings.Split(filename, "_")
	if len(filenameParts) < 3 {
		return "", "", "", fmt.Errorf("文件名格式错误(%s)", filename)
	}

	version := strings.TrimPrefix(filenameParts[0], "openjdk-")
	filenameParts1 := strings.Split(filenameParts[1], "-")
	var goos, goarch string
	if len(filenameParts1) > 1 {
		goos = filenameParts1[0]
		goarch = filenameParts1[1]
	}

	// 标准化GOOS值
	goos = strings.ReplaceAll(goos, "osx", "darwin")
	goos = strings.ReplaceAll(goos, "macos", "darwin")
	switch goos {
	case "linux", "darwin", "windows":
		// 已经是标准值
	default:
		return "", "", "", fmt.Errorf("未知操作系统(%s)", goos)
	}

	// 标准化GOARCH值
	goarch = strings.ReplaceAll(goarch, "aarch64", "arm64")
	switch goarch {
	case "x64", "amd64", "arm64", "aarch64":
		if goarch == "x64" {
			goarch = "amd64" // Go标准中使用amd64而不是x64
		}
	default:
		return "", "", "", fmt.Errorf("未知系统架构(%s)", goarch)
	}

	return goos, goarch, version, nil
}

// ParseURL 爬取所有 华为云 JDK 下载地址
// 目录层次:
//   - /openjdk/11.0.1/openjdk-11.0.1_windows-x64_bin.zip
//
// 返回:
//   - []TOpenJDK: 所有 JDK 下载条目
//   - error: 错误信息
func (s *TWebHuawei) ParseURL() ([]TOpenJDK, error) {
	var allDownloads []TOpenJDK

	// 1. 获取版本目录
	versions, err := getVerDirs(s.BaseURL)
	if err != nil {
		return nil, err
	}

	// 2. 遍历每个版本
	for _, version := range versions {
		versionURL := s.BaseURL + version
		// 去除版本号中的 '/' 字符用于显示
		versionDisplay := strings.TrimSuffix(version, "/")
		logs.Debug("==================================================")
		logs.Debug("-= 处理 JDK %s 版本 =-", versionDisplay)
		logs.Debug("==================================================")

		// 3. 获取 JDK 文件
		downloads, err := s.GetJDKFiles(versionURL)
		if err != nil {
			logs.Warning("获取文件失败，%v", err)
			continue
		}

		// 打印找到的 OpenJDK 信息
		if len(downloads) > 0 {
			for _, jdk := range downloads {
				logs.Debug("%s", jdk.String())
			}
		}
		allDownloads = append(allDownloads, downloads...)
	}

	return allDownloads, nil
}

// GetJDKFiles 获取指定 URL 中的所有 JDK 文件下载地址及详细信息
// 参数:
//   - fileURL: 文件列表页面的 URL
//
// 返回:
//   - []TOpenJDK: JDK 下载条目列表
//   - error: 错误信息
func (s *TWebHuawei) GetJDKFiles(fileURL string) ([]TOpenJDK, error) {
	htmlContent, err := fetchHTML(fileURL)
	if err != nil {
		return nil, err
	}

	// 使用华为云专用的解析函数，匹配 .zip 和 .tar.gz 文件
	files, err := parseFileListHuawei(htmlContent, `\.(zip|tar\.gz)$`)
	if err != nil {
		return nil, err
	}

	var downloads []TOpenJDK
	for _, file := range files {
		// 从文件名中提取版本号
		version := extractVersion(file.Name)

		// 格式化时间
		formattedTime := parseTime(file.LastModified)

		goos, goarch, _, err := s.ParseWebFileName(file.Name)
		if err != nil {
			logs.Warning(err)
			continue
		}

		downloads = append(downloads, TOpenJDK{
			Version:      version,
			Filename:     file.Name,
			URL:          fileURL + file.Name,
			Size:         file.Size,
			LastModified: formattedTime,
			GOOS:         goos,
			GOARCH:       goarch,
		})
	}

	return downloads, nil
}

// ParseWebFileName 解析文件名获取GOOS、GOARCH和版本信息
// 输入:openjdk-10.0.1_windows-x64_bin.tar.gz
// 返回:Goos,Arch,Version,error
func (s *TWebHuawei) ParseWebFileName(filename string) (string, string, string, error) {
	filename = strings.TrimSpace(filename)
	if filename == "" {
		return "", "", "", fmt.Errorf("文件名不能为空")
	}

	// 解析文件名获取版本、GOOS和GOARCH
	filenameParts := strings.Split(filename, "_")
	if len(filenameParts) < 3 {
		return "", "", "", fmt.Errorf("文件名格式错误(%s)", filename)
	}

	version := strings.TrimPrefix(filenameParts[0], "openjdk-")
	filenameParts1 := strings.Split(filenameParts[1], "-")
	var goos, goarch string
	if len(filenameParts1) > 1 {
		goos = filenameParts1[0]
		goarch = filenameParts1[1]
	}

	// 标准化GOOS值
	goos = strings.ReplaceAll(goos, "osx", "darwin")
	goos = strings.ReplaceAll(goos, "macos", "darwin")
	switch goos {
	case "linux", "darwin", "windows":
		// 已经是标准值
	default:
		return "", "", "", fmt.Errorf("未知操作系统(%s)", goos)
	}

	// 标准化GOARCH值
	goarch = strings.ReplaceAll(goarch, "aarch64", "arm64")
	switch goarch {
	case "x64", "amd64", "arm64", "aarch64":
		if goarch == "x64" {
			goarch = "amd64" // Go标准中使用amd64而不是x64
		}
	default:
		return "", "", "", fmt.Errorf("未知系统架构(%s)", goarch)
	}

	return goos, goarch, version, nil
}

type TAzulJDK struct {
	PackageUUID        string `json:"package_uuid"`
	Name               string `json:"name"`
	JavaVersion        []int  `json:"java_version"`
	OpenjdkBuildNumber int    `json:"openjdk_build_number"`
	Latest             bool   `json:"latest"`
	DownloadURL        string `json:"download_url"`
	Product            string `json:"product"`
	DistroVersion      []int  `json:"distro_version"`
	AvailabilityType   string `json:"availability_type"`
}

// ParseURL 爬取所有 Azul JDK 下载地址
// Azul 提供 REST API，可以直接获取 JDK 下载地址
// API 文档: https://api.azul.com/metadata/v1/docs/swagger
//
// 返回:
//   - []TOpenJDK: 所有 JDK 下载条目
//   - error: 错误信息
func (s *TWebAzul) ParseURL() ([]TOpenJDK, error) {
	var allDownloads []TOpenJDK

	// 支持的操作系统和架构组合
	targets := []struct {
		os     string
		arch   string
		goos   string
		goarch string
	}{
		{"windows", "x86", "windows", "amd64"},
		{"windows", "arm", "windows", "arm64"},
		{"linux", "x86", "linux", "amd64"},
		{"linux", "arm", "linux", "arm64"},
		{"macos", "x86", "darwin", "amd64"},
		{"macos", "arm", "darwin", "arm64"},
	}

	// 遍历每个目标平台
	for _, target := range targets {
		logs.Debug("==================================================")
		logs.Debug("-= 获取 %s/%s 的 JDK 列表 =-", target.goos, target.goarch)
		logs.Debug("==================================================")

		// 构建 API URL
		apiURL := fmt.Sprintf("%s?os=%s&arch=%s&archive_type=zip&java_package_type=jdk&javafx_bundled=false&latest=true&release_status=ga&availability_types=CA&certifications=tck&page=1&page_size=100",
			s.BaseURL, target.os, target.arch)

		// 获取 JDK 列表
		downloads, err := s.FetchJDKList(apiURL, target.goos, target.goarch)
		if err != nil {
			logs.Warning(err.Error())
			continue
		}

		// 打印找到的 JDK 信息
		if len(downloads) > 0 {
			for _, jdk := range downloads {
				logs.Debug("%s", jdk.String())
			}
		}
		allDownloads = append(allDownloads, downloads...)
	}

	return allDownloads, nil
}

// FetchJDKList 从 Azul API 获取 JDK 列表
// 参数:
//   - apiURL: API 地址
//   - goos: 操作系统标准名称
//   - goarch: 架构标准名称
//
// 返回:
//   - []TOpenJDK: JDK 下载条目列表
//   - error: 错误信息
func (s *TWebAzul) FetchJDKList(apiURL string, goos string, goarch string) ([]TOpenJDK, error) {
	// 创建 HTTP 请求
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败，%v", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	// 忽略证书验证
	customTransport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	// 发送请求
	client := &http.Client{
		Transport: customTransport,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败，%v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败，%v", err)
	}

	// 解析 JSON
	var azulJDKs []TAzulJDK
	err = json.Unmarshal(body, &azulJDKs)
	if err != nil {
		return nil, fmt.Errorf("解析 JSON 失败，%v", err)
	}

	// 转换为 TOpenJDK 格式
	var downloads []TOpenJDK
	for _, azulJDK := range azulJDKs {
		// 跳过包含 crac 的版本（这是特殊版本）
		if strings.Contains(azulJDK.Name, "-crac-") {
			continue
		}

		// 提取版本号
		version := s.ExtractVersion(azulJDK.JavaVersion)

		// 提取文件大小（从 URL 获取，如果可能的话）
		size := "Unknown"

		// 使用当前时间作为最后修改时间
		lastModified := time.Now().Format("2006-01-02 15:04")

		downloads = append(downloads, TOpenJDK{
			Version:      version,
			Filename:     azulJDK.Name,
			URL:          azulJDK.DownloadURL,
			Size:         size,
			LastModified: lastModified,
			GOOS:         goos,
			GOARCH:       goarch,
		})
	}

	return downloads, nil
}

// ExtractVersion 从 JavaVersion 数组中提取版本号字符串
// 参数:
//   - javaVersion: Java 版本数组 [major, minor, patch]
//
// 返回:
//   - string: 版本号字符串
func (s *TWebAzul) ExtractVersion(javaVersion []int) string {
	if len(javaVersion) == 0 {
		return "unknown"
	}

	// 主版本号
	major := javaVersion[0]

	// 如果有次版本号和补丁版本号
	if len(javaVersion) >= 3 {
		minor := javaVersion[1]
		patch := javaVersion[2]
		if minor > 0 || patch > 0 {
			return fmt.Sprintf("%d.%d.%d", major, minor, patch)
		}
	}

	return fmt.Sprintf("%d", major)
}

// ============================================================
// Adoptium API 相关代码
// ============================================================

// TAdoptiumReleases 表示 Adoptium 的可用版本响应
type TAdoptiumReleases struct {
	AvailableReleases []int `json:"available_releases"`
}

// TAdoptiumAsset 表示 Adoptium 的资源响应
type TAdoptiumAsset struct {
	Binary struct {
		Architecture string `json:"architecture"`
		ImageType    string `json:"image_type"`
		OS           string `json:"os"`
		Package      struct {
			Name string `json:"name"`
			Link string `json:"link"`
			Size int64  `json:"size"`
		} `json:"package"`
		UpdatedAt string `json:"updated_at"`
	} `json:"binary"`
	Version struct {
		Major          int    `json:"major"`
		Minor          int    `json:"minor"`
		Security       int    `json:"security"`
		OpenjdkVersion string `json:"openjdk_version"`
	} `json:"version"`
}

// ParseURL 爬取所有 Adoptium JDK 下载地址
// Adoptium 提供 REST API，可以直接获取 JDK 下载地址
// API 文档: https://api.adoptium.net/q/swagger-ui/
//
// 返回:
//   - []TOpenJDK: 所有 JDK 下载条目
//   - error: 错误信息
func (s *TWebAdoptium) ParseURL() ([]TOpenJDK, error) {
	var allDownloads []TOpenJDK

	// 1. 获取所有可用的版本
	releases, err := s.GetAvailableReleases()
	if err != nil {
		return nil, fmt.Errorf("获取可用版本失败，%v", err)
	}

	logs.Debug("找到 %d 个可用版本", len(releases))

	// 支持的操作系统和架构组合
	targets := []struct {
		os     string
		arch   string
		goos   string
		goarch string
	}{
		{"windows", "x64", "windows", "amd64"},
		{"windows", "aarch64", "windows", "arm64"},
		{"linux", "x64", "linux", "amd64"},
		{"linux", "aarch64", "linux", "arm64"},
		{"mac", "x64", "darwin", "amd64"},
		{"mac", "aarch64", "darwin", "arm64"},
	}

	// 2. 遍历每个版本
	for _, version := range releases {
		logs.Debug("==================================================")
		logs.Debug("-= 处理 JDK %d 版本 =-", version)
		logs.Debug("==================================================")

		// 遍历每个目标平台
		for _, target := range targets {
			// 获取该版本和平台的 JDK
			downloads, err := s.FetchJDKForPlatform(version, target.os, target.arch, target.goos, target.goarch)
			if err != nil {
				continue // 某些版本可能不支持某些平台
			}

			// 打印找到的 JDK 信息
			for _, jdk := range downloads {
				logs.Debug("%s", jdk.String())
			}
			allDownloads = append(allDownloads, downloads...)
		}
	}

	return allDownloads, nil
}

// GetAvailableReleases 获取所有可用的版本
// 返回:
//   - []int: 版本号列表
//   - error: 错误信息
func (s *TWebAdoptium) GetAvailableReleases() ([]int, error) {
	apiURL := "https://api.adoptium.net/v3/info/available_releases"

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败，%v", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")

	// 创建自定义 HTTP 客户端，跳过证书验证
	customTransport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: customTransport,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败，%v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败，%v", err)
	}

	var releases TAdoptiumReleases
	err = json.Unmarshal(body, &releases)
	if err != nil {
		return nil, fmt.Errorf("解析 JSON 失败，%v", err)
	}

	return releases.AvailableReleases, nil
}

// FetchJDKForPlatform 获取指定版本和平台的 JDK
// 参数:
//   - version: JDK 版本号
//   - os: 操作系统 (windows, linux, mac)
//   - arch: 架构 (x64, aarch64)
//   - goos: Go 标准操作系统名称
//   - goarch: Go 标准架构名称
//
// 返回:
//   - []TOpenJDK: JDK 下载条目列表
//   - error: 错误信息
func (s *TWebAdoptium) FetchJDKForPlatform(version int, os, arch, goos, goarch string) ([]TOpenJDK, error) {
	// 构建 API URL
	apiURL := fmt.Sprintf("https://api.adoptium.net/v3/assets/latest/%d/hotspot?architecture=%s&image_type=jdk&os=%s&vendor=eclipse",
		version, arch, os)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败，%v", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")

	// 忽略证书验证
	customTransport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: customTransport,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败，%v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败，%v", err)
	}

	// 解析 JSON
	var assets []TAdoptiumAsset
	err = json.Unmarshal(body, &assets)
	if err != nil {
		return nil, fmt.Errorf("解析 JSON 失败，%v", err)
	}

	// 转换为 TOpenJDK 格式
	var downloads []TOpenJDK
	for _, asset := range assets {
		// 只处理 package (zip/tar.gz)，不处理 installer (msi/pkg)
		if asset.Binary.Package.Link == "" {
			continue
		}

		// 提取版本号
		versionStr := asset.Version.OpenjdkVersion
		if versionStr == "" {
			versionStr = fmt.Sprintf("%d.%d.%d", asset.Version.Major, asset.Version.Minor, asset.Version.Security)
		}

		// 格式化文件大小
		size := formatFileSize(fmt.Sprintf("%d", asset.Binary.Package.Size))

		// 格式化时间
		lastModified := parseTime(asset.Binary.UpdatedAt)

		downloads = append(downloads, TOpenJDK{
			Version:      versionStr,
			Filename:     asset.Binary.Package.Name,
			URL:          asset.Binary.Package.Link,
			Size:         size,
			LastModified: lastModified,
			GOOS:         goos,
			GOARCH:       goarch,
		})
	}

	return downloads, nil
}
