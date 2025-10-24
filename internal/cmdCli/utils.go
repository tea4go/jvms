package cmdCli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
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

	// 确定要查找的文件名
	javacName := jdk.GetJavacName()

	fs.WalkDir(os.DirFS(jdkTempFile), ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// 跳过非文件
		if d.IsDir() {
			return nil
		}

		// 查找 javac 文件
		if filepath.Base(path) == javacName {
			logs.Debug("找到 javac 所在目录 - %s", path)

			// 获取 bin 目录的父目录（即 JAVA_HOME）
			absPath := filepath.Join(jdkTempFile, path)

			// 向上查找，直到找到包含 bin 目录的父目录
			binDir := filepath.Dir(absPath)
			if filepath.Base(binDir) == "bin" {
				javaHome = filepath.Dir(binDir)
				logs.Debug("确定 JAVA_HOME = %s\n", javaHome)
				return fs.SkipAll
			}
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
		if err == nil && len(versions) > 0 {
			// 本地缓存文件存在，且有数据，则用缓存数据
			fmt.Println("Load the version list from the local cache ...")
			return versions, nil
		}
		// 如果加载失败，继续从网络获取
		fmt.Println("Cache file fails to load and is retrieved from the network ...")
	}

	var versions []entity.TJDKVersion

	var downOpenJDKs []jdk.TOpenJDK
	var err error

	// 根据 webtype 参数选择不同的镜像源
	switch strings.ToLower(cfx.WebType) {
	case "tuna":
		logs.Info("📦 使用镜像源: 清华大学 (Tsinghua University)")
		WebJDK := jdk.TWebTuna{}
		WebJDK.BaseURL = "https://mirrors.tuna.tsinghua.edu.cn/Adoptium/"
		logs.Info("🔗 镜像地址: %s", WebJDK.BaseURL)
		downOpenJDKs, err = WebJDK.ParseURL()
	case "lzu":
		logs.Info("📦 使用镜像源: 兰州大学 (Lanzhou University)")
		WebJDK := jdk.TWebLzu{}
		WebJDK.BaseURL = "https://mirror4.lzu.edu.cn/openjdk/"
		logs.Info("🔗 镜像地址: %s", WebJDK.BaseURL)
		downOpenJDKs, err = WebJDK.ParseURL()
	case "injdk":
		logs.Info("📦 使用镜像源: InJDK 网站")
		WebJDK := jdk.TWebInjdk{}
		WebJDK.BaseURL = "https://d10.injdk.cn/openjdk/openjdk/"
		logs.Info("🔗 镜像地址: %s", WebJDK.BaseURL)
		downOpenJDKs, err = WebJDK.ParseURL()
	case "azul":
		logs.Info("📦 使用镜像源: Azul Zulu")
		WebJDK := jdk.TWebAzul{}
		WebJDK.BaseURL = "https://api.azul.com/metadata/v1/zulu/packages"
		logs.Info("🔗 镜像地址: %s", WebJDK.BaseURL)
		downOpenJDKs, err = WebJDK.ParseURL()
	case "adoptium":
		logs.Info("📦 使用镜像源: Eclipse Adoptium")
		WebJDK := jdk.TWebAdoptium{}
		WebJDK.BaseURL = "https://api.adoptium.net/v3"
		logs.Info("🔗 镜像地址: %s", WebJDK.BaseURL)
		downOpenJDKs, err = WebJDK.ParseURL()
	case "huawei":
	default:
		// 默认使用huawei镜像源
		logs.Info("📦 使用镜像源: 华为云 (Huawei Cloud)")
		WebJDK := jdk.TWebHuawei{}
		WebJDK.BaseURL = "https://mirrors.huaweicloud.com/openjdk/"
		logs.Info("🔗 镜像地址: %s", WebJDK.BaseURL)
		downOpenJDKs, err = WebJDK.ParseURL()
	}
	// 检查爬取是否出错
	if err != nil {
		logs.Error("❌ 错误: 获取JDK版本镜像源失败，请检查网络连接或镜像源地址是否正确\n%v", err)
		return nil, err
	}

	// 获取当前系统和架构信息
	currentOS := runtime.GOOS     // linux, darwin, windows
	currentArch := runtime.GOARCH // amd64, arm64

	// 过滤符合当前系统和架构的JDK
	filteredJDKs := filterJDKsByPlatform(downOpenJDKs, currentOS, currentArch)

	if len(filteredJDKs) == 0 {
		return nil, fmt.Errorf("❌ 错误: 未找到适合当前系统(%s-%s)的JDK版本", currentOS, currentArch)
	}

	if cfx.WebAll {
		for _, oneJdk := range filteredJDKs {
			v := entity.TJDKVersion{}
			v.Version = fmt.Sprintf("openjdk-%s", oneJdk.Version)
			v.Url = oneJdk.URL
			versions = append(versions, v)
		}
	} else {
		// 使用 map 来去重，只保留每个主版本号的最新完整版本
		majorVersionMap := make(map[string]entity.TJDKVersion)

		for _, oneJdk := range filteredJDKs {
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
		logs.Warning("警告: 保存缓存失败，%v", err.Error())
		// 不返回错误，因为主要功能已经完成
	} else {
		logs.Debug("版本列表已缓存到本地 -", cacheFile)
	}

	return versions, nil
}

// 根据系统和架构过滤JDK列表
func filterJDKsByPlatform(jdks []jdk.TOpenJDK, osType, arch string) []jdk.TOpenJDK {
	var filtered []jdk.TOpenJDK

	// 构建平台标识符
	platformSuffix := getPlatformSuffix(osType, arch)

	for _, oneJdk := range jdks {
		// 检查URL是否包含对应的平台标识
		if strings.Contains(oneJdk.URL, platformSuffix) {
			filtered = append(filtered, oneJdk)
		}
	}

	return filtered
}

// 获取平台后缀标识
func getPlatformSuffix(osType, arch string) string {
	var osSuffix, archSuffix string

	// 转换操作系统名称
	switch osType {
	case "linux":
		osSuffix = "linux"
	case "darwin":
		osSuffix = "macos"
	case "windows":
		osSuffix = "windows"
	default:
		osSuffix = osType
	}

	// 转换架构名称
	switch arch {
	case "amd64":
		archSuffix = "x64"
	case "arm64":
		archSuffix = "aarch64"
	default:
		archSuffix = arch
	}

	// 返回格式: linux-x64, macos-aarch64, windows-x64 等
	return fmt.Sprintf("%s-%s", osSuffix, archSuffix)
}

// 跨平台设置 JAVA_HOME 环境变量
func setJavaHome(javaHome string) error {
	switch runtime.GOOS {
	case "windows":
		return setJavaHomeWindows(javaHome)
	case "darwin", "linux":
		return setJavaHomeUnix(javaHome)
	default:
		return fmt.Errorf("unsupported operating systems (%s)", runtime.GOOS)
	}
}

// setJavaHomeWindows Windows 下设置环境变量
func setJavaHomeWindows(javaHome string) error {
	// 设置 JAVA_HOME
	logs.Debug("setx JAVA_HOME=%s", javaHome)
	cmd := exec.Command("setx", "JAVA_HOME", javaHome)
	err := cmd.Run()
	if err != nil {
		return errors.New("设置用户环境变量 JAVA_HOME 失败")
	}
	logs.Debug("设置 JAVA_HOME 环境变量为 ", javaHome)

	// 设置 PATH
	javaBinPath := filepath.Join(javaHome, "bin")
	currentExePath := file.GetCurrentPath()
	currentPath := os.Getenv("PATH")

	// 检查是否已存在，不存在才添加
	newPath := currentPath
	if !containsPath(currentPath, javaBinPath) {
		newPath = javaBinPath + ";" + newPath
	}
	if !containsPath(currentPath, currentExePath) {
		newPath = currentExePath + ";" + newPath
	}
	cmd = exec.Command("setx", "PATH", newPath)
	err = cmd.Run()
	if err != nil {
		return errors.New("设置用户环境变量 PATH 失败")
	}
	fmt.Println("已将 jvms.exe 添加到用户 PATH 环境变量")
	return nil
}

// setJavaHomeUnix Unix 系统（macOS/Linux）下设置环境变量
func setJavaHomeUnix(javaHome string) error {
	if javaHome == "" {
		return fmt.Errorf("JavaHome目录没配置，请先执行 init 命令")
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("获取用户主目录失败，%v", err)
	}
	logs.Debug("开始设置环境变量")

	// 需要更新的配置文件列表
	configFiles := []string{
		filepath.Join(homeDir, ".zshrc"),
		filepath.Join(homeDir, ".bashrc"),
		filepath.Join(homeDir, ".bash_profile"),
	}

	// JVMS 配置块的标记
	jvmsStart := "# JVMS - Java Version Manager - START"
	jvmsEnd := "# JVMS - Java Version Manager - END"

	// 新的配置内容
	newConfig := fmt.Sprintf("%s\nexport JAVA_HOME=\"%s\"\nexport PATH=\"$JAVA_HOME/bin:$PATH\"\n%s",
		jvmsStart, javaHome, jvmsEnd)

	updatedFiles := []string{}

	// 遍历所有配置文件
	for _, configFile := range configFiles {
		// 只处理已存在的文件
		if !file.Exists(configFile) {
			continue
		}
		logs.Debug("加载文件 %s ......", configFile)

		// 读取现有配置
		data, err := os.ReadFile(configFile)
		if err != nil {
			logs.Warning("⚠️ 警告: 读取 %s 失败，%v", configFile, err)
			continue
		}
		content := string(data)

		var newContent string

		// 检查是否已存在 JVMS 配置
		if strings.Contains(content, jvmsStart) {
			// 替换现有配置
			startIdx := strings.Index(content, jvmsStart)
			endIdx := strings.Index(content, jvmsEnd)
			logs.Debug("发现残留配置 (%d~%d)", startIdx, endIdx)

			if endIdx > startIdx {
				// 找到完整的配置块，替换它
				endIdx += len(jvmsEnd)
				newContent = content[:startIdx] + newConfig + content[endIdx:]
			} else {
				// 配置块不完整，删除旧的开始标记，追加新配置
				newContent = strings.Replace(content, jvmsStart, "", 1) + "\n" + newConfig
			}
		} else {
			logs.Debug("尾部追加新配置")
			// 确保文件末尾有换行符
			if !strings.HasSuffix(content, "\n") {
				content += "\n"
			}
			newContent = content + "\n" + newConfig + "\n"
		}

		// 写回配置文件
		err = os.WriteFile(configFile, []byte(newContent), 0644)
		if err != nil {
			logs.Warning("⚠️ 警告: 写入 %s 失败，%v", configFile, err)
			continue
		}

		updatedFiles = append(updatedFiles, configFile)
	}

	// 设置当前会话的环境变量
	logs.Debug("export JAVA_HOME=%s", javaHome)
	os.Setenv("JAVA_HOME", javaHome)

	// 输出更新结果
	if len(updatedFiles) == 0 {
		return fmt.Errorf("未找到任何配置文件 (.zshrc, .bashrc, .bash_profile)")
	}

	logs.Debug("✓ 已更新以下配置文件:")
	for k, f := range updatedFiles {
		logs.Debug("%d - %s", k, f)
	}

	// 检测当前使用的 shell 并给出提示
	shell := os.Getenv("SHELL")
	shellName := filepath.Base(shell)

	var sourceFile string
	switch shellName {
	case "zsh":
		sourceFile = filepath.Join(homeDir, ".zshrc")
	case "bash":
		// bash 优先使用 .bash_profile，如果不存在则使用 .bashrc
		if file.Exists(filepath.Join(homeDir, ".bash_profile")) {
			sourceFile = filepath.Join(homeDir, ".bash_profile")
		} else {
			sourceFile = filepath.Join(homeDir, ".bashrc")
		}
	default:
		sourceFile = updatedFiles[0] // 使用第一个更新的文件
	}

	fmt.Printf("💡 The run command takes effect:\n")
	fmt.Printf("   source %s\n", sourceFile)

	return nil
}

// 检查 PATH 中是否包含指定路径,避免重复添加
func containsPath(pathEnv, targetPath string) bool {
	lowerTarget := strings.ToLower(filepath.Clean(targetPath))
	paths := strings.Split(pathEnv, ";")
	for _, p := range paths {
		if strings.ToLower(filepath.Clean(strings.TrimSpace(p))) == lowerTarget {
			return true
		}
	}
	return false
}
