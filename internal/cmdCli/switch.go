package cmdCli

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/codegangsta/cli"
	"github.com/tea4go/jvms/internal/entity"
	"github.com/tea4go/jvms/utils/file"
	"github.com/tea4go/jvms/utils/jdk"
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
	if ctx.NArg() == 0 {
		return errors.New("您应该输入版本或索引号，输入 \"jvms list\" 查看已安装的版本")
	}

	v := ctx.Args().First()

	// 检查输入是否为数字（索引）
	index, err := strconv.Atoi(v)
	if err == nil && index > 0 {
		// 输入是有效的数字，获取已安装的JDK列表
		installedJDKs := jdk.GetInstalled(cfx.Store)
		if len(installedJDKs) == 0 {
			return errors.New("未找到已安装的JDK")
		}

		if index > len(installedJDKs) {
			return fmt.Errorf("无效的索引: %d，应该在 1 到 %d 之间", index, len(installedJDKs))
		}

		v = installedJDKs[index-1]
		fmt.Printf("使用索引 %d 选择 JDK %s\n", index, v)
	}

	if !jdk.IsVersionInstalled(cfx.Store, v) {
		fmt.Printf("jdk %s 未安装。", v)
		return nil
	}

	// 创建或更新符号链接
	if file.Exists(cfx.JavaHome) {
		err := os.Remove(cfx.JavaHome)
		if err != nil {
			return errors.New("切换 jdk 失败，请手动删除 " + cfx.JavaHome)
		}
	}

	cmd := exec.Command("cmd", "/C", "setx", "JAVA_HOME", cfx.JavaHome, "/M")
	err = cmd.Run()
	if err != nil {
		return errors.New("设置环境变量 `JAVA_HOME` 失败: 请以管理员身份运行")
	}

	err = os.Symlink(filepath.Join(cfx.Store, v), cfx.JavaHome)
	if err != nil {
		return errors.New("切换 jdk 失败, " + err.Error())
	}

	fmt.Println("切换成功。\n当前使用 JDK " + v)
	cfx.CurrentJDKVersion = v
	return nil
}
