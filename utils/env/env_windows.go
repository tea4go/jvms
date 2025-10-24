//go:build windows
// +build windows

package env

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/tea4go/jvms/utils/file"
	"golang.org/x/sys/windows/registry"
)

// Windows 下设置环境变量
func SetJavaHome(javaHome string) error {
	// 1. 设置 JAVA_HOME
	if err := setUserEnvVar("JAVA_HOME", javaHome); err != nil {
		return fmt.Errorf("设置 JAVA_HOME 失败: %v", err)
	}
	fmt.Println("已设置 JAVA_HOME 环境变量为:", javaHome)

	// 2. 获取当前用户的 PATH（从注册表读取）
	currentPath, err := getUserEnvVar("PATH")
	if err != nil {
		return fmt.Errorf("读取用户 PATH 失败: %v", err)
	}

	// 3. 准备要添加的路径
	javaBinPath := filepath.Join(javaHome, "bin")
	currentExePath := file.GetCurrentPath()

	// 4. 构建新的 PATH
	newPath := currentPath
	if !containsPath(currentPath, javaBinPath) {
		newPath = javaBinPath + ";" + newPath
	}
	if !containsPath(currentPath, currentExePath) {
		newPath = currentExePath + ";" + newPath
	}

	// 5. 设置新的 PATH
	if err := setUserEnvVar("PATH", newPath); err != nil {
		return fmt.Errorf("设置 PATH 失败: %v", err)
	}

	fmt.Println("已将 jvms.exe 添加到用户 PATH 环境变量")

	// 6. 广播环境变量变更消息
	// broadcastEnvironmentChange()

	return nil
}

// 从注册表获取用户环境变量
func getUserEnvVar(name string) (string, error) {
	key, err := registry.OpenKey(registry.CURRENT_USER,
		`Environment`, registry.QUERY_VALUE)
	if err != nil {
		return "", err
	}
	defer key.Close()

	value, _, err := key.GetStringValue(name)
	if err != nil {
		// 如果变量不存在，返回空字符串而不是错误
		if err == registry.ErrNotExist {
			return "", nil
		}
		return "", err
	}
	return value, nil
}

// 设置用户环境变量到注册表
func setUserEnvVar(name, value string) error {
	key, err := registry.OpenKey(registry.CURRENT_USER,
		`Environment`, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer key.Close()

	return key.SetStringValue(name, value)
}

// 检查 PATH 中是否包含指定路径,避免重复添加
func containsPath(pathEnv, targetPath string) bool {
	// lowerTarget := strings.ToLower(filepath.Clean(targetPath))
	// paths := strings.Split(pathEnv, ";")
	// for _, p := range paths {
	// 	if strings.ToLower(filepath.Clean(strings.TrimSpace(p))) == lowerTarget {
	// 		return true
	// 	}
	// }
	// return false
	paths := strings.Split(pathEnv, ";")
	targetPath = strings.ToLower(filepath.Clean(targetPath))

	for _, p := range paths {
		if strings.ToLower(filepath.Clean(p)) == targetPath {
			return true
		}
	}
	return false
}

// broadcastEnvironmentChange 广播环境变量变更消息
func broadcastEnvironmentChange() {
	// 通知系统环境变量已更改
	cmd := exec.Command("cmd", "/C",
		`setx DUMMY_VAR "" >nul 2>&1 & REG DELETE "HKCU\Environment" /F /V DUMMY_VAR >nul 2>&1`)
	cmd.Run()
}
