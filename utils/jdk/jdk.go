package jdk

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/tea4go/jvms/utils/file"
)

// TOpenJDK 表示一个 OpenJDK 下载条目
type TOpenJDK struct {
	Version      string `json:"version"`       // 版本号
	Filename     string `json:"filename"`      // 文件名
	URL          string `json:"url"`           // 下载链接
	Size         string `json:"size"`          // 文件大小
	LastModified string `json:"last_modified"` // 最后修改时间
	GOOS         string `json:"goos"`          // 操作系统类型
	GOARCH       string `json:"goarch"`        // 系统架构
}

// String 返回 TOpenJDK 的字符串表示
// 返回:
//   - string: 格式化的 OpenJDK 信息字符串
func (jdk TOpenJDK) String() string {
	javaname := fmt.Sprintf("OpenJDK%s.%s-%s", jdk.Version, jdk.GOOS, jdk.GOARCH)
	sizetext := strings.ReplaceAll(jdk.Size, "i", "")
	return fmt.Sprintf("%-35s | %-10s | %s",
		javaname, sizetext, jdk.Filename)
}

// GetInstalled 获取已安装的 JDK 版本列表
// 参数:
//
//	root - JDK 安装的根目录路径
//
// 返回值:
//
//	[]string - 已安装的 JDK 版本名称列表（按倒序排列）
func GetInstalled(root string) []string {
	list := make([]string, 0)
	files, _ := os.ReadDir(root)
	for i := len(files) - 1; i >= 0; i-- {
		if files[i].IsDir() {
			list = append(list, files[i].Name())
		}
	}
	return list
}

// IsVersionInstalled 检查指定版本的 JDK 是否已安装
// 参数:
//
//	root - JDK 安装的根目录路径
//	version - 要检查的 JDK 版本号
//
// 返回值:
//
//	bool - 已安装返回 true，否则返回 false
func IsVersionInstalled(root string, version string) bool {
	javacPath := filepath.Join(root, version, "bin", GetJavacName())
	return file.Exists(javacPath)
}

// GetJavacName 根据操作系统返回 javac 的文件名
func GetJavacName() string {
	if runtime.GOOS == "windows" {
		return "javac.exe"
	}
	return "javac"
}
