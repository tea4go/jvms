package cmdCli

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/tea4go/jvms/internal/entity"
	"github.com/tea4go/jvms/utils/file"
	"github.com/urfave/cli/v2"
)

// initCmd 执行初始化命令
// 用于初始化配置文件和设置环境变量
// 参数:
//
//	args - 命令参数
//	cfx - 配置对象指针
//
// 返回值:
//
//	error - 执行错误
func initCmd(ctx *cli.Context, cfx *entity.TConfig) error {
	javaHome := ctx.String("java_home")
	if javaHome == "" {
		javaHome = defaultJavaHome()
	}

	// 设置 JAVA_HOME
	if ctx.IsSet("java_home") || cfx.JavaHome == "" {
		cfx.JavaHome = javaHome
	}

	cmd := exec.Command("cmd", "/C", "setx", "JAVA_HOME", cfx.JavaHome, "/M")
	err := cmd.Run()
	if err != nil {
		return errors.New("设置环境变量 `JAVA_HOME` 失败: 请以管理员身份运行")
	}
	fmt.Println("设置 `JAVA_HOME` 环境变量为 ", cfx.JavaHome)

	// 设置 PATH
	path := fmt.Sprintf(`%s/bin;%s;%s`, cfx.JavaHome, os.Getenv("PATH"), file.GetCurrentPath())
	cmd = exec.Command("cmd", "/C", "setx", "path", path, "/m")
	err = cmd.Run()
	if err != nil {
		return errors.New("设置环境变量 `PATH` 失败: 请以管理员身份运行")
	}
	fmt.Println("添加 jvms.exe 到 `path` 环境变量")

	return nil
}

func defaultJavaHome() string {
	return filepath.Join(os.Getenv("ProgramFiles"), "jdk")
}
