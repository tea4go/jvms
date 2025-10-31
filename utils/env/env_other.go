//go:build !windows

package env

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	logs "github.com/tea4go/gh/log4go"
	"github.com/tea4go/jvms/utils/file"
)

// Unix 系统（macOS/Linux）下设置环境变量
func SetJavaHome(javaHome string) error {
	if javaHome == "" {
		return fmt.Errorf("JavaHome目录没配置，请先执行 init 命令")
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("获取用户主目录失败，%v", err)
	}
	logs.Debug("开始设置环境变量")

	// 需要更新的配置文件列表（.profile 优先，因为它更通用）
	configFiles := []string{
		filepath.Join(homeDir, ".profile"),
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
		// .profile 文件如果不存在则创建，其他文件只处理已存在的
		if configFile == filepath.Join(homeDir, ".profile") {
			if !file.Exists(configFile) {
				logs.Debug("创建 .profile 文件")
				// 创建空文件
				if err := os.WriteFile(configFile, []byte(""), 0644); err != nil {
					logs.Warning("⚠️ 警告: 创建 %s 失败，%v", configFile, err)
					continue
				}
			}
		} else if !file.Exists(configFile) {
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
			if !strings.HasSuffix(content, "\n") && len(content) > 0 {
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
		return fmt.Errorf("未找到任何配置文件")
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
