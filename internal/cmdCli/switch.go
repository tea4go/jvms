package cmdCli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	logs "github.com/tea4go/gh/log4go"
	"github.com/tea4go/jvms/internal/entity"
	"github.com/tea4go/jvms/utils/file"
	"github.com/tea4go/jvms/utils/jdk"
	"github.com/urfave/cli/v2"
)

// switchCmd 执行切换JDK版本的命令
// 切换到指定版本或索引号的JDK
// 参数:
//
//	args - 命令参数
//	cfx - 配置对象指针
//
// 返回值:
//
//	error - 执行错误
func switchCmd(ctx *cli.Context, cfx *entity.TConfig) error {
	return switchFunc(ctx, cfx)
}

// switchFunc 切换JDK版本的处理函数
// 参数:
//
//	args - 命令参数
//	cfx - 配置对象指针
//
// 返回值:
//
//	error - 执行错误
func switchFunc(ctx *cli.Context, cfx *entity.TConfig) error {
	logs.Debug("切换JDK版本 - %s", cfx.Store)
	if ctx.NArg() == 0 {
		return errors.New("您应该输入版本或索引号，输入 jvms list 查看已安装的版本")
	}
	if cfx.JavaHome == "" {
		return errors.New("您先执行 init 命令，初始化 JAVA_HOME 目录")
	}
	v := ctx.Args().First()

	// 检查输入是否为数字（索引）
	index, err := strconv.Atoi(v)
	if err == nil && index > 0 {
		// 输入是有效的数字，获取已安装的JDK列表
		installedJDKs := jdk.GetInstalled(cfx.Store)
		if len(installedJDKs) == 0 {
			return errors.New("没有可用的JDK版本")
		}

		if index > len(installedJDKs) {
			return fmt.Errorf("无效的索引 %d，应该为(1~%d)", index, len(installedJDKs))
		}

		v = installedJDKs[index-1]
		fmt.Printf("选择 %d - %s\n", index, v)
	}

	if !jdk.IsVersionInstalled(cfx.Store, v) {
		fmt.Printf("未安装JDK版本为 %s\n", v)
		return nil
	}

	// 创建或更新符号链接
	if file.Exists(cfx.JavaHome) {
		err := os.Remove(cfx.JavaHome)
		if err != nil {
			return errors.New("切换 JDK版本 失败，请手动删除 " + cfx.JavaHome)
		}
	}

	// 切换JDK不需要再重新设置环境变量
	// err = setJavaHome(cfx.JavaHome)
	// if err != nil {
	// 	return err
	// }

	err = os.Symlink(filepath.Join(cfx.Store, v), cfx.JavaHome)
	if err != nil {
		return fmt.Errorf("切换 jdk 失败(%s)，%s", cfx.JavaHome, err.Error())
	}

	fmt.Println("切换成功。\n当前使用 " + v)
	cfx.CurrentJDKVersion = v
	return nil
}
