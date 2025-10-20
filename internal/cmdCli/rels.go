package cmdCli

import (
	"fmt"

	"github.com/spf13/pflag"
	"github.com/tea4go/jvms/internal/entity"
	"github.com/tea4go/jvms/utils/web"
)

// rlsCmd 执行显示可下载版本列表的命令
// 显示可供下载的JDK版本列表
// 参数:
//
//	args - 命令参数
//	cfx - 配置对象指针
//
// 返回值:
//
//	error - 执行错误
func rlsCmd(args []string, cfx *entity.TConfig) error {
	// 创建命令专用的 FlagSet
	fs := pflag.NewFlagSet("rls", pflag.ContinueOnError)

	showAll := fs.BoolP("all", "a", false, "列出所有版本")
	webType := fs.StringP("webtype", "t", "huawei", "切换软件源")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if cfx.Proxy != "" {
		web.SetProxy(cfx.Proxy)
	}

	cfx.WebAll = *showAll
	cfx.WebType = *webType

	versions, err := getJdkVersions(cfx)
	if err != nil {
		return err
	}

	for i, version := range versions {
		fmt.Printf("%3d) %s\n", i+1, version.Version)
		if !*showAll && i >= 9 {
			fmt.Println("\n使用 \"jvm rls -a\" 显示所有版本")
			break
		}
	}

	if len(versions) == 0 {
		fmt.Println("没有可供下载的 jdk 版本。")
	}

	return nil
}
