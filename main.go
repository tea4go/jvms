// Package main 是 JVMS (JDK Version Manager for Windows) 的主入口
// JVMS 是一个用于在 Windows 系统上管理多个 JDK 版本的命令行工具
package main

import (
	"os"

	logs "github.com/tea4go/gh/log4go"

	"github.com/tea4go/jvms/internal/cmdCli"
)

var AppName string = "jvms"
var AppVersion = "3.0.7"
var BuildTime = ""
var IsBeta string = "false"

// main 是程序的入口函数
func main() {
	logs.StartLogger()
	logs.SetLevel(1)
	app := cmdCli.NewApp(AppName, AppVersion, BuildTime)
	app.Run(os.Args)
}
