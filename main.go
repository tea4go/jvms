// Package main 是 JVMS (JDK Version Manager for Windows) 的主入口
// JVMS 是一个用于在 Windows 系统上管理多个 JDK 版本的命令行工具
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	logs "github.com/tea4go/gh/log4go"

	"github.com/tea4go/jvms/internal/cmdCli"
	"github.com/tea4go/jvms/internal/entity"
	"github.com/tea4go/jvms/utils/file"
	"github.com/tea4go/jvms/utils/web"
	"github.com/tucnak/store"
)

// version 定义当前 JVMS 的版本号
var AppName string = "jvms"
var AppVersion = "3.0.7"
var BuildTime = ""
var IsBeta string = "false"

// cfx 全局配置对象，存储 JVMS 的运行配置
var cfx entity.TConfig

// main 是程序的入口函数
func main() {
	flag.Usage = func() {
		printUsage()
	}

	// 初始化配置
	if err := startup(); err != nil {
		logs.Emergency(err.Error())
		return
	}
	defer shutdown()

	// 使用 os.Args 直接解析，避免全局 pflag 干扰命令参数
	args := os.Args[1:] // 跳过程序名

	// 检查是否有 --version 或 -v
	for _, arg := range args {
		if arg == "--version" || arg == "-v" {
			fmt.Println(AppVersion)
			return
		}
		if arg == "--help" || arg == "-h" {
			printUsage()
			return
		}
	}

	// 如果没有参数，显示帮助
	if len(args) == 0 {
		printUsage()
		return
	}

	// 获取命令和参数
	command := args[0]
	cmdArgs := args[1:]

	// 创建命令参数
	cmdParams := &cmdCli.TCommandParams{
		Config: &cfx,
	}

	// 执行命令
	if err := cmdCli.Execute(command, cmdArgs, cmdParams); err != nil {
		logs.Emergency(err.Error())
	}
}

func filepathJoin(elem ...string) string {
	path := filepath.Join(elem...)
	if runtime.GOOS == "windows" {
		return strings.ReplaceAll(path, "\\", "/")
	}
	return path
}

// printUsage 打印使用说明
func printUsage() {
	fmt.Println("NAME:")
	fmt.Println("   jvms - JDK Version Manager (JVMS) for Windows")
	fmt.Println("")
	fmt.Println("USAGE:")
	fmt.Println("   jvms [全局选项] 命令 [命令选项] [参数...]")
	fmt.Println("")
	fmt.Printf("VERSION:\n   %s - %s\n", AppVersion, BuildTime)
	fmt.Println("")
	fmt.Println("COMMANDS:")
	fmt.Println("   init        初始化配置文件")
	fmt.Println("   list, ls    列出当前已安装的JDK")
	fmt.Println("   install, i  安装可用的远程JDK")
	fmt.Println("   switch, s   切换使用指定的版本或索引号")
	fmt.Println("   use, u      切换使用指定的版本或索引号")
	fmt.Println("   remove, rm  删除指定的版本")
	fmt.Println("   rls         显示可供下载的版本列表")
	fmt.Println("   proxy       设置下载使用的代理")
	fmt.Println("   help, h     显示命令列表或命令帮助")
	fmt.Println("")
	fmt.Println("全局选项:")
	fmt.Println("   --help, -h     显示帮助")
	fmt.Println("   --version, -v  显示版本")
}

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
func startup() error {
	// 注册 JSON 格式的配置存储器
	store.Register(
		"json",
		func(v interface{}) ([]byte, error) {
			return json.MarshalIndent(v, "", "    ")
		},
		json.Unmarshal)

	// 初始化配置存储
	store.Init("jvms")

	// 加载配置文件
	if err := store.Load("jvms.json", &cfx); err != nil {
		return errors.New("failed to load the config:" + err.Error())
	}

	// 获取当前可执行文件所在路径
	s := file.GetCurrentPath()

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
func shutdown() {
	if err := store.Save("jvms.json", &cfx); err != nil {
		logs.Warning("警告: 保存配置失败: %s\n", err.Error())
	}
}
