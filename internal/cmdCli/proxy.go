package cmdCli

import (
	"fmt"

	"github.com/tea4go/jvms/internal/entity"
	"github.com/urfave/cli"
)

// proxyCmd 执行代理设置命令
// 设置或显示用于下载的代理服务器
// 参数:
//
//	args - 命令参数
//	cfx - 配置对象指针
//
// 返回值:
//
//	error - 执行错误
func proxyCmd(ctx *cli.Context, cfx *entity.TConfig) error {
	if ctx.Bool("show") {
		fmt.Printf("当前代理: %s\n", cfx.Proxy)
		return nil
	}

	if ctx.IsSet("set") {
		cfx.Proxy = ctx.String("set")
		fmt.Printf("代理已设置为: %s\n", cfx.Proxy)
	}

	return nil
}
