// Package main 是 JVMS (JDK Version Manager for Windows) 的主入口
// JVMS 是一个用于在 Windows 系统上管理多个 JDK 版本的命令行工具
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

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
	logs.StartLogger()

	// 初始化配置
	if err := startup(); err != nil {
		fmt.Println(err.Error())
		return
	}
	defer shutdown()

	// 创建命令参数
	cmdParams := &cmdCli.TCommandParams{
		Config: &cfx,
	}

	app := cmdCli.NewApp(AppName, AppVersion, BuildTime, cmdParams)
	if err := app.Run(os.Args); err != nil {
		fmt.Println(err.Error())
	}
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
func shutdown() {
	if err := store.Save("jvms.json", &cfx); err != nil {
		fmt.Println("警告: 保存配置失败，", err.Error())
	}
}
