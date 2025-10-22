package cmdCli

import (
	"fmt"

	"github.com/tea4go/jvms/internal/entity"
	"github.com/tea4go/jvms/utils/jdk"
	"github.com/urfave/cli/v2"
)

// listCmd 执行列出已安装JDK的命令
// 显示所有已安装的JDK版本，并标记当前正在使用的版本
// 参数:
//
//	args - 命令参数
//	cfx - 配置对象指针
//
// 返回值:
//
//	error - 执行错误
func listCmd(ctx *cli.Context, cfx *entity.TConfig) error {
	fmt.Println("Installed JDK (* marks in use)")
	v := jdk.GetInstalled(cfx.Store)
	for i, version := range v {
		if cfx.CurrentJDKVersion == version {
			fmt.Printf("* %d - %snvm \n", i+1, version)
		} else {
			fmt.Printf("  %d - %s\n", i+1, version)
		}
	}
	if len(v) == 0 {
		fmt.Println("No installations recognized.")
	}
	return nil
}
