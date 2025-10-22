package cmdCli

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/tucnak/store"
	"github.com/urfave/cli/v2"

	logs "github.com/tea4go/gh/log4go"
	"github.com/tea4go/jvms/internal/entity"
	"github.com/tea4go/jvms/utils/file"
	"github.com/tea4go/jvms/utils/web"
)

// cfx 全局配置对象，存储 JVMS 的运行配置
var cfx entity.TConfig

// startup 在应用启动前执行
// 主要功能：
//  1. 注册 JSON 序列化/反序列化器
//  2. 加载配置文件 (jvms.json)
//  3. 初始化存储路径和下载路径
//  4. 设置代理（如果配置了）
//
// 返回:
//
//	error - 初始化失败时返回错误
func startup(c *cli.Context) error {
	logs.Debug("加载配置 jvms.json 文件")
	// 注册 JSON 格式的配置存储器
	store.Register(
		"json",
		func(v any) ([]byte, error) {
			return json.MarshalIndent(v, "", "    ")
		},
		json.Unmarshal)

	// 初始化配置存储
	store.Init("jvms")

	// 加载配置文件
	if err := store.Load("jvms.json", &cfx); err != nil {
		return errors.New("加载配置 jvms.json 失败，" + err.Error())
	}

	// 获取当前可执行文件所在路径
	s := file.GetCurrentPath()

	// 是否显示所有JDK
	cfx.WebAll = false

	// 下载源
	cfx.WebType = "lzu"

	// 设置 JDK 存储目录路径
	cfx.Store = filepath.Join(s, "store")

	// 设置下载临时目录路径
	cfx.Download = filepath.Join(s, "download")

	// 如果配置了代理，设置 HTTP 代理
	if cfx.Proxy != "" {
		web.SetProxy(cfx.Proxy)
	}

	return nil
}

// shutdown 在应用关闭后执行
// 主要功能：保存配置到 jvms.json 文件
func shutdown(c *cli.Context) error {
	if err := store.Save("jvms.json", &cfx); err != nil {
		return fmt.Errorf("警告: 保存配置失败，%s", err.Error())
	}
	return nil
}

// NewApp 创建 CLI 应用实例
func NewApp(appName, appVersion, buildTime string) *cli.App {
	app := cli.NewApp()
	app.Name = appName
	app.HideVersion = true
	app.HideHelp = true
	app.HideHelpCommand = true
	app.Version = appVersion
	app.Usage = "JDK Version Manager"
	app.Metadata = map[string]any{
		"build_time": buildTime,
	}
	app.Before = startup
	app.After = shutdown
	app.Action = func(ctx *cli.Context) error {
		printAppUsage(appVersion, buildTime)
		return nil
	}

	app.Flags = []cli.Flag{
		&cli.IntFlag{
			Name:    "loglevel",
			Aliases: []string{"l"},
			Value:   5,
			Usage:   "日志级别(1~7)",
			Action: func(ctx *cli.Context, v int) error {
				if v <= 7 || v > 0 {
					logs.SetLevel(v)
					return nil
				}
				return fmt.Errorf("日志级别错误(1~7) - %d", v)
			},
		},
	}

	app.Commands = []*cli.Command{
		newInitCommand(&cfx),
		newListCommand(&cfx),
		newInstallCommand(&cfx),
		newSwitchCommand(&cfx),
		newUseCommand(&cfx),
		newRemoveCommand(&cfx),
		newRlsCommand(&cfx),
		newProxyCommand(&cfx),
		newHelpCommand(),
		newVersionCommand(),
	}

	app.CommandNotFound = func(ctx *cli.Context, command string) {
		fmt.Fprintf(cli.ErrWriter, "未知命令: %s\n使用 'jvms help' 查看可用命令\n", command)
	}

	return app
}

func printAppUsage(version, buildTime string) {
	fmt.Println("NAME:")
	fmt.Println("   jvms - JDK Version Manager (JVMS)")
	fmt.Println("")
	fmt.Println("USAGE:")
	fmt.Println("   jvms [Global options] command [command options] [arguments...]")
	fmt.Println("")
	if buildTime != "" {
		fmt.Printf("VERSION:\n   %s - %s\n", version, buildTime)
	} else {
		fmt.Printf("VERSION:\n   %s\n", version)
	}
	fmt.Println("")
	fmt.Println("COMMANDS:")
	fmt.Println("   init        Initialize the configuration file")
	fmt.Println("   list, ls    List installed versions")
	fmt.Println("   install, i  Download and install a version")
	fmt.Println("   use, u      Switch to specified version")
	fmt.Println("   remove, rm  Uninstall a version")
	fmt.Println("   rls         List remote versions available for install")
	fmt.Println("   proxy       Set a proxy to use for downloads.")
	fmt.Println("   help, h     Shows a list of commands or help for one command，eg: help rls")
	fmt.Println("   version, v  Show version")
	fmt.Println("")
	fmt.Println("GLOBAL OPTIONS:")
	fmt.Println("   --loglevel,-l set log level (1~7)")
}

func newInitCommand(cfx *entity.TConfig) *cli.Command {
	defaultHome := defaultJavaHome()
	return &cli.Command{
		Name:  "init",
		Usage: "初始化配置文件",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "java_home",
				Usage: fmt.Sprintf("指定 JAVA_HOME 位置 (默认: %s)", defaultHome),
				Value: defaultHome,
			},
		},
		Action: func(ctx *cli.Context) error {
			return initCmd(ctx, cfx)
		},
	}
}

func newListCommand(cfx *entity.TConfig) *cli.Command {
	return &cli.Command{
		Name:    "list",
		Aliases: []string{"ls"},
		Usage:   "列出当前已安装的JDK",
		Action: func(ctx *cli.Context) error {
			return listCmd(ctx, cfx)
		},
	}
}

func newInstallCommand(cfx *entity.TConfig) *cli.Command {
	return &cli.Command{
		Name:    "install",
		Aliases: []string{"i"},
		Usage:   "安装可用的远程JDK",
		Action: func(ctx *cli.Context) error {
			return installCmd(ctx, cfx)
		},
	}
}

func newSwitchCommand(cfx *entity.TConfig) *cli.Command {
	return &cli.Command{
		Name:    "switch",
		Aliases: []string{"s"},
		Usage:   "切换使用指定的版本或索引号",
		Action: func(ctx *cli.Context) error {
			return switchCmd(ctx, cfx)
		},
	}
}

func newUseCommand(cfx *entity.TConfig) *cli.Command {
	return &cli.Command{
		Name:    "use",
		Aliases: []string{"u"},
		Usage:   "切换使用指定的版本或索引号",
		Action: func(ctx *cli.Context) error {
			return useCmd(ctx, cfx)
		},
	}
}

func newRemoveCommand(cfx *entity.TConfig) *cli.Command {
	return &cli.Command{
		Name:    "remove",
		Aliases: []string{"rm"},
		Usage:   "删除指定的版本",
		Action: func(ctx *cli.Context) error {
			return removeCmd(ctx, cfx)
		},
	}
}

func newRlsCommand(cfx *entity.TConfig) *cli.Command {
	return &cli.Command{
		Name:  "rls",
		Usage: "显示可供下载的版本列表",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "all",
				Aliases: []string{"a"},
				Usage:   "列出所有版本",
			},
			&cli.StringFlag{
				Name:    "webtype",
				Aliases: []string{"t"},
				Usage:   "设置 OpenJDK 下载源",
				Value:   "lzu",
			},
		},
		Action: func(ctx *cli.Context) error {
			return rlsCmd(ctx, cfx)
		},
	}
}

func newProxyCommand(cfx *entity.TConfig) *cli.Command {
	return &cli.Command{
		Name:  "proxy",
		Usage: "设置下载使用的代理",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "show",
				Usage: "显示当前代理",
			},
			&cli.StringFlag{
				Name:  "set",
				Usage: "设置代理",
			},
		},
		Action: func(ctx *cli.Context) error {
			return proxyCmd(ctx, cfx)
		},
	}
}

func newHelpCommand() *cli.Command {
	return &cli.Command{
		Name:    "help",
		Aliases: []string{"h"},
		Usage:   "显示命令列表或命令帮助",
		Action: func(ctx *cli.Context) error {
			return printHelp(ctx)
		},
	}
}

func newVersionCommand() *cli.Command {
	return &cli.Command{
		Name:    "version",
		Aliases: []string{"v"},
		Usage:   "显示版本号",
		Action: func(ctx *cli.Context) error {
			return printVersion(ctx)
		},
	}
}

func printVersion(ctx *cli.Context) error {
	// logs.Emergency("测试日志级别 1")
	// logs.Critical("测试日志级别 2")
	// logs.Error("测试日志级别 3")
	// logs.Warning("测试日志级别 4")
	// logs.Notice("测试日志级别 5")
	// logs.Info("测试日志级别 6")
	// logs.Debug("测试日志级别 7")
	fmt.Println(ctx.App.Version)
	return nil
}

func printHelp(ctx *cli.Context) error {
	if !ctx.Args().Present() {
		buildTime := ""
		if v, ok := ctx.App.Metadata["build_time"].(string); ok {
			buildTime = v
		}

		printAppUsage(ctx.App.Version, buildTime)
		return nil
	}

	cmd := ctx.Args().First()
	switch cmd {
	case "init":
		fmt.Println("init - 初始化配置文件")
		fmt.Println("")
		fmt.Println("用法: jvms init [选项]")
		fmt.Println("")
		fmt.Println("选项:")
		fmt.Printf("  --java_home <路径>      指定 JAVA_HOME 位置 (默认: %s)\n", defaultJavaHome())
	case "install", "i":
		fmt.Println("install - 安装可用的远程JDK")
		fmt.Println("")
		fmt.Println("用法: jvms install <版本>")
	case "list", "ls":
		fmt.Println("list - 列出当前已安装的JDK")
		fmt.Println("")
		fmt.Println("用法: jvms list")
	case "switch", "s":
		fmt.Println("switch - 切换使用指定的版本或索引号")
		fmt.Println("")
		fmt.Println("用法: jvms switch <版本|索引号>")
	case "use", "u":
		fmt.Println("use - 切换使用指定的版本或索引号")
		fmt.Println("")
		fmt.Println("用法: jvms use <版本|索引号>")
	case "remove", "rm":
		fmt.Println("remove - 删除指定的版本")
		fmt.Println("")
		fmt.Println("用法: jvms remove <版本>")
	case "rls":
		fmt.Println("rls - 显示可供下载的版本列表")
		fmt.Println("")
		fmt.Println("用法: jvms rls [选项]")
		fmt.Println("")
		fmt.Println("选项:")
		fmt.Println("  --all, -a        列出所有版本")
		fmt.Println("  --webtype, -t    设置 OpenJDK 下载源")
		fmt.Println("")
		fmt.Println("OpenJDK 下载源:")
		fmt.Println("  lzu      - 兰州大学开源软件镜像站")
		fmt.Println("  tuna     - 清华大学开源软件镜像站")
		fmt.Println("  injdk    - InJDK 网站")
		fmt.Println("  huawei   - 华为云镜像站")
		fmt.Println("  azul     - Azul Zulu OpenJDK")
		fmt.Println("  adoptium - Eclipse Adoptium")
	case "proxy":
		fmt.Println("proxy - 设置下载使用的代理")
		fmt.Println("")
		fmt.Println("用法: jvms proxy [选项]")
		fmt.Println("")
		fmt.Println("选项:")
		fmt.Println("  --show          显示当前代理")
		fmt.Println("  --set <代理>    设置代理")
	default:
		fmt.Printf("未知命令: %s\n", cmd)
	}
	return nil
}
