package cmdCli

import (
	"fmt"

	"github.com/tea4go/jvms/internal/entity"
	"github.com/tea4go/jvms/utils/jdk"
	"github.com/urfave/cli"
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
	fmt.Println("已安装的 jdk (* 标记正在使用):")
	v := jdk.GetInstalled(cfx.Store)
	for i, version := range v {
		str := ""
		if cfx.CurrentJDKVersion == version {
			str = fmt.Sprintf("%s  * %d) %s", str, i+1, version)
		} else {
			str = fmt.Sprintf("%s    %d) %s", str, i+1, version)
		}
		fmt.Printf(str + "\n")
	}
	if len(v) == 0 {
		fmt.Println("未识别到已安装的版本。")
	}
	return nil
}
