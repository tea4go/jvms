package cmdCli

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	logs "github.com/tea4go/gh/log4go"
	"github.com/tea4go/jvms/internal/entity"
	"github.com/tea4go/jvms/utils/file"
	"github.com/tea4go/jvms/utils/jdk"
	"github.com/urfave/cli/v2"
)

// switchCmd 执行切换JDK版本的命令
// 切换到指定版本或索引号的JDK
// 参数:
//
//	args - 命令参数
//	cfx - 配置对象指针
//
// 返回值:
//
//	error - 执行错误
func switchCmd(ctx *cli.Context, cfx *entity.TConfig) error {
	return switchFunc(ctx, cfx)
}

// switchFunc 切换JDK版本的处理函数
// 参数:
//
//	args - 命令参数
//	cfx - 配置对象指针
//
// 返回值:
//
//	error - 执行错误
func switchFunc(ctx *cli.Context, cfx *entity.TConfig) error {
	logs.Debug("切换JDK版本 - %s", cfx.Store)
	if ctx.NArg() == 0 {
		return errors.New("您应该输入版本或索引号，输入 jvms list 查看已安装的版本")
	}
	if cfx.JavaHome == "" {
		return errors.New("您先执行 init 命令，初始化 JAVA_HOME 目录")
	}
	v := ctx.Args().First()

	// 检查输入是否为数字（索引）
	index, err := strconv.Atoi(v)
	if err == nil && index > 0 {
		// 输入是有效的数字，获取已安装的JDK列表
		installedJDKs := jdk.GetInstalled(cfx.Store)
		if len(installedJDKs) == 0 {
			return errors.New("no JDK installations found")
		}

		if index > len(installedJDKs) {
			return fmt.Errorf("invalid index %d,should be between 1 and %d", index, len(installedJDKs))
		}

		v = installedJDKs[index-1]
		fmt.Printf("Using index %d to select %s\n", index, v)
	}

	if !jdk.IsVersionInstalled(cfx.Store, v) {
		fmt.Printf("%s is not installed.\n", v)
		return nil
	}

	// 创建或更新符号链接
	if file.Exists(cfx.JavaHome) {
		err := os.Remove(cfx.JavaHome)
		if err != nil {
			return errors.New("switch jdk failed, please manually remove " + cfx.JavaHome)
		}
	}

	err = setJavaHome(cfx.JavaHome)
	if err != nil {
		return err
	}

	err = os.Symlink(filepath.Join(cfx.Store, v), cfx.JavaHome)
	if err != nil {
		return fmt.Errorf("switch jdk failed (%s), %s", cfx.JavaHome, err.Error())
	}

	fmt.Println("Switch success.\nNow using " + v)
	cfx.CurrentJDKVersion = v
	return nil
}

// 跨平台设置 JAVA_HOME 环境变量
func setJavaHome(javaHome string) error {
	switch runtime.GOOS {
	case "windows":
		return setJavaHomeWindows(javaHome)
	case "darwin", "linux":
		return setJavaHomeUnix(javaHome)
	default:
		return fmt.Errorf("不支持的操作系统 (%s)", runtime.GOOS)
	}
}

// setJavaHomeWindows Windows 下设置环境变量
func setJavaHomeWindows(javaHome string) error {
	logs.Debug("setx JAVA_HOME=%s /M", javaHome)
	cmd := exec.Command("cmd", "/C", "setx", "JAVA_HOME", javaHome, "/M")
	err := cmd.Run()
	if err != nil {
		return errors.New("设置环境变量 JAVA_HOME 失败，请以管理员身份运行")
	}
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
			fmt.Printf("⚠️ 警告: 读取 %s 失败，%v\n", configFile, err)
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
			fmt.Printf("⚠️ 警告: 写入 %s 失败，%v\n", configFile, err)
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

	fmt.Println("✓ 已更新以下配置文件:")
	for k, f := range updatedFiles {
		fmt.Printf("%d - %s\n", k, f)
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

	fmt.Printf("\n💡 提示: 运行以下命令使配置立即生效:\n")
	fmt.Printf("   source %s\n", sourceFile)

	return nil
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
