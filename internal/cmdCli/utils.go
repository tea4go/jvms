package cmdCli

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	logs "github.com/tea4go/gh/log4go"
	"github.com/tea4go/jvms/internal/entity"
	"github.com/tea4go/jvms/utils/file"
	"github.com/tea4go/jvms/utils/jdk"
)

// getJavaHome 从临时JDK文件目录中获取JAVA_HOME路径
// 参数:
//
//	jdkTempFile - JDK临时解压目录路径
//
// 返回值:
//
//	string - JAVA_HOME路径(包含javac.exe的父目录)
func getJavaHome(jdkTempFile string) string {
	var javaHome string
	fs.WalkDir(os.DirFS(jdkTempFile), ".", func(path string, d fs.DirEntry, err error) error {
		if filepath.Base(path) == "javac.exe" {
			temPath := strings.Replace(path, "bin/javac.exe", "", -1)
			javaHome = filepath.Join(jdkTempFile, temPath)
			return fs.SkipDir
		}
		return nil
	})
	return javaHome
}

// parseVersion 解析版本号字符串为数字切片
// 参数:
//
//	version - 版本号字符串，例如 "openjdk-12.0.1" 或 "12.0.1"
//
// 返回值:
//
//	[]int - 版本号数字切片，例如 [12, 0, 1]
func parseVersion(version string) []int {
	// 移除 "openjdk-" 前缀
	version = strings.TrimPrefix(version, "openjdk-")

	// 按点分割版本号
	parts := strings.Split(version, ".")
	numbers := make([]int, 0, len(parts))

	for _, part := range parts {
		// 转换为整数
		if num, err := strconv.Atoi(part); err == nil {
			numbers = append(numbers, num)
		} else {
			// 如果转换失败，视为0
			numbers = append(numbers, 0)
		}
	}

	return numbers
}

// compareVersions 比较两个版本号
// 参数:
//
//	v1 - 版本1
//	v2 - 版本2
//
// 返回值:
//
//	int - 如果 v1 < v2 返回 -1，v1 == v2 返回 0，v1 > v2 返回 1
func compareVersions(v1, v2 string) int {
	nums1 := parseVersion(v1)
	nums2 := parseVersion(v2)

	// 比较每个数字段
	maxLen := len(nums1)
	if len(nums2) > maxLen {
		maxLen = len(nums2)
	}

	for i := 0; i < maxLen; i++ {
		n1 := 0
		n2 := 0

		if i < len(nums1) {
			n1 = nums1[i]
		}
		if i < len(nums2) {
			n2 = nums2[i]
		}

		if n1 < n2 {
			return -1
		} else if n1 > n2 {
			return 1
		}
	}

	return 0
}

// getCacheFilePath 获取缓存文件路径
// 返回值:
//
//	string - 缓存文件的完整路径
func getCacheFilePath() string {
	currentPath := file.GetCurrentPath()
	return filepath.Join(currentPath, "jdkdlindex.json")
}

// isCacheValid 检查缓存文件是否有效（是否是当月创建的）
// 参数:
//
//	cacheFile - 缓存文件路径
//
// 返回值:
//
//	bool - 如果缓存文件存在且是当月创建的返回 true，否则返回 false
func isCacheValid(cacheFile string) bool {
	// 检查文件是否存在
	fileInfo, err := os.Stat(cacheFile)
	if err != nil {
		return false
	}

	// 获取文件修改时间
	modTime := fileInfo.ModTime()

	// 获取当前时间
	now := time.Now()

	// 检查是否在同一年和同一月
	return modTime.Year() == now.Year() && modTime.Month() == now.Month()
}

// loadCachedVersions 从缓存文件加载版本列表
// 参数:
//
//	cacheFile - 缓存文件路径
//
// 返回值:
//
//	[]entity.TJDKVersion - JDK 版本列表
//	error - 错误信息
func loadCachedVersions(cacheFile string) ([]entity.TJDKVersion, error) {
	data, err := os.ReadFile(cacheFile)
	if err != nil {
		return nil, err
	}

	var versions []entity.TJDKVersion
	if err := json.Unmarshal(data, &versions); err != nil {
		return nil, err
	}

	return versions, nil
}

// saveCachedVersions 保存版本列表到缓存文件
// 参数:
//
//	cacheFile - 缓存文件路径
//	versions - JDK 版本列表
//
// 返回值:
//
//	error - 错误信息
func saveCachedVersions(cacheFile string, versions []entity.TJDKVersion) error {
	data, err := json.MarshalIndent(versions, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(cacheFile, data, 0644)
}

// extractMajorVersion 提取主版本号
// 参数:
//
//	version - 完整版本号字符串，例如 "openjdk-12.0.1"
//
// 返回值:
//
//	string - 主版本号，例如 "12"
func extractMajorVersion(version string) string {
	// 移除 "openjdk-" 前缀
	version = strings.TrimPrefix(version, "openjdk-")

	// 按点分割，取第一段
	parts := strings.Split(version, ".")
	if len(parts) > 0 {
		return parts[0]
	}
	return version
}

// getJdkVersions 获取可供下载的JDK版本列表
// 参数:
//
//	cfx - 配置对象指针
//
// 返回值:
//
//	[]entity.TJDKVersion - JDK版本列表（按版本号从大到小排序，每个主版本号只显示最新的完整版本）
//	error - 错误信息
func getJdkVersions(cfx *entity.TConfig) ([]entity.TJDKVersion, error) {
	// 获取缓存文件路径
	cacheFile := getCacheFilePath()

	// 检查缓存是否有效
	if isCacheValid(cacheFile) {
		// 尝试从缓存加载
		versions, err := loadCachedVersions(cacheFile)
		if err == nil {
			fmt.Println("从本地缓存加载版本列表...")
			return versions, nil
		}
		// 如果加载失败，继续从网络获取
		fmt.Println("缓存文件加载失败，从网络获取...")
	}

	var versions []entity.TJDKVersion

	var downOpenJDKs []jdk.TOpenJDK
	var err error

	// 根据 webtype 参数选择不同的镜像源
	switch strings.ToLower(cfx.WebType) {
	case "tuna":
		logs.Debug("\n📦 使用镜像源: 清华大学 (Tsinghua University)")
		WebJDK := jdk.TWebTuna{}
		WebJDK.BaseURL = "https://mirrors.tuna.tsinghua.edu.cn/Adoptium/"
		logs.Debug("🔗 镜像地址: %s\n\n", WebJDK.BaseURL)
		downOpenJDKs, err = WebJDK.ParseURL()
	case "lzu":
		logs.Debug("\n📦 使用镜像源: 兰州大学 (Lanzhou University)")
		WebJDK := jdk.TWebLzu{}
		WebJDK.BaseURL = "https://mirror4.lzu.edu.cn/openjdk/"
		logs.Debug("🔗 镜像地址: %s\n\n", WebJDK.BaseURL)
		downOpenJDKs, err = WebJDK.ParseURL()
	case "huawei":
		logs.Debug("\n📦 使用镜像源: 华为云 (Huawei Cloud)")
		WebJDK := jdk.TWebHuawei{}
		WebJDK.BaseURL = "https://mirrors.huaweicloud.com/openjdk/"
		logs.Debug("🔗 镜像地址: %s\n\n", WebJDK.BaseURL)
		downOpenJDKs, err = WebJDK.ParseURL()
	case "injdk":
		logs.Debug("\n📦 使用镜像源: InJDK 网站")
		WebJDK := jdk.TWebInjdk{}
		WebJDK.BaseURL = "https://d10.injdk.cn/openjdk/openjdk/"
		logs.Debug("🔗 镜像地址: %s\n\n", WebJDK.BaseURL)
		downOpenJDKs, err = WebJDK.ParseURL()
	case "azul":
		logs.Debug("\n📦 使用镜像源: Azul Zulu")
		WebJDK := jdk.TWebAzul{}
		WebJDK.BaseURL = "https://api.azul.com/metadata/v1/zulu/packages"
		logs.Debug("🔗 镜像地址: %s\n\n", WebJDK.BaseURL)
		downOpenJDKs, err = WebJDK.ParseURL()
	case "adoptium":
		logs.Debug("\n📦 使用镜像源: Eclipse Adoptium")
		WebJDK := jdk.TWebAdoptium{}
		WebJDK.BaseURL = "https://api.adoptium.net/v3"
		logs.Debug("🔗 镜像地址: %s\n\n", WebJDK.BaseURL)
		downOpenJDKs, err = WebJDK.ParseURL()
	default:
		err = fmt.Errorf("❌ 错误: 未知的镜像源类型(%s)", cfx.WebType)
	}
	// 检查爬取是否出错
	if err != nil {
		return nil, err
	}

	if cfx.WebAll {
		for _, oneJdk := range downOpenJDKs {
			v := entity.TJDKVersion{}
			v.Version = fmt.Sprintf("openjdk-%s", oneJdk.Version)
			v.Url = oneJdk.URL
			versions = append(versions, v)
		}
	} else {
		// 使用 map 来去重，只保留每个主版本号的最新完整版本
		majorVersionMap := make(map[string]entity.TJDKVersion)

		for _, oneJdk := range downOpenJDKs {
			versionName := fmt.Sprintf("openjdk-%s", oneJdk.Version)
			majorVersion := extractMajorVersion(versionName)

			// 如果该主版本号还没有记录，或者当前版本更新，则保存完整版本号
			if existing, exists := majorVersionMap[majorVersion]; !exists {
				majorVersionMap[majorVersion] = entity.TJDKVersion{
					Version: versionName, // 保留完整版本号，例如 "openjdk-12.0.2"
					Url:     oneJdk.URL,
				}
			} else {
				// 比较版本，保留更新的版本（完整版本号）
				if compareVersions(versionName, existing.Version) > 0 {
					majorVersionMap[majorVersion] = entity.TJDKVersion{
						Version: versionName, // 保留完整版本号
						Url:     oneJdk.URL,
					}
				}
			}
		}

		// 将 map 转换为切片
		for _, v := range majorVersionMap {
			versions = append(versions, v)
		}
	}

	// 对版本进行排序（从大到小）
	sort.Slice(versions, func(i, j int) bool {
		return compareVersions(versions[i].Version, versions[j].Version) > 0
	})

	// 保存到缓存
	if err := saveCachedVersions(cacheFile, versions); err != nil {
		fmt.Printf("警告: 保存缓存失败: %v\n", err)
		// 不返回错误，因为主要功能已经完成
	} else {
		fmt.Println("版本列表已缓存到本地 -", cacheFile)
	}

	return versions, nil
}
