package cmdCli

import (
	"os"
	"path/filepath"

	"github.com/tea4go/jvms/internal/entity"
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

	// 初始化设置环境变量
	return setJavaHome(cfx.JavaHome)
}

func defaultJavaHome() string {
	return filepath.Join(os.Getenv("ProgramFiles"), "jdk")
}
