package cmdCli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	logs "github.com/tea4go/gh/log4go"
	"github.com/tea4go/jvms/internal/entity"
	"github.com/tea4go/jvms/utils/file"
	"github.com/tea4go/jvms/utils/jdk"
	"github.com/tea4go/jvms/utils/web"
	"github.com/urfave/cli/v2"
)

// installCmd 执行安装JDK的命令
// 从远程源下载并安装指定版本的JDK
// 参数:
//
//	args - 命令参数
//	cfx - 配置对象指针
//
// 返回值:
//
//	error - 执行错误
func installCmd(ctx *cli.Context, cfx *entity.TConfig) error {
	if cfx.Proxy != "" {
		web.SetProxy(cfx.Proxy)
	}

	if ctx.NArg() == 0 {
		return errors.New("无效的版本，输入 git jvms rls 查看可供安装的版本")
	}

	v := ctx.Args().First()

	if jdk.IsVersionInstalled(cfx.Store, v) {
		fmt.Println("版本 " + v + " 已经安装。")
		return nil
	}

	versions, err := getJdkVersions(cfx)
	if err != nil {
		return err
	}

	if !file.Exists(cfx.Download) {
		os.MkdirAll(cfx.Download, 0777)
	}
	if !file.Exists(cfx.Store) {
		os.MkdirAll(cfx.Store, 0777)
	}

	for _, version := range versions {
		if version.Version == v {
			fmt.Printf("正在下载 %s - %s\n", v, version.Url)
			dlzipfile, success := web.GetJDK(cfx.Download, v, version.Url)
			if success {
				fmt.Printf("正在安装 %s ...\n", v)

				// 解压 JDK 到临时目录
				jdktempfile := filepath.Join(cfx.Download, fmt.Sprintf("%s_temp", v))
				logs.Debug("解压 %s 到临时目录 - %s", v, jdktempfile)
				if file.Exists(jdktempfile) {
					err := os.RemoveAll(jdktempfile)
					if err != nil {
						panic(err)
					}
				}
				err := file.Extract(dlzipfile, jdktempfile)
				if err != nil {
					return fmt.Errorf("解压失败，%s", err.Error())
				}

				// 复制 JDK 文件到安装目录
				temJavaHome := getJavaHome(jdktempfile)
				if temJavaHome == "" {
					return fmt.Errorf("当前下载的 %s 为无效版本，请手工检查 %s", v, jdktempfile)
				}

				logs.Debug("配置所在目录：%s", cfx.Store)
				destJavaHome := filepath.Join(cfx.Store, v)
				logs.Debug("移动目录 %s -> %s", temJavaHome, destJavaHome)
				err = os.Rename(temJavaHome, destJavaHome)
				if err != nil {
					return fmt.Errorf("移动目录失败，%s", err.Error())
				}

				// 删除临时目录
				// 可以考虑保留临时文件
				err = os.RemoveAll(jdktempfile)
				if err != nil {
					fmt.Printf("警告: 清理临时目录失败，%v\n", err)
				}

				fmt.Println("安装成功完成。")
				fmt.Printf("如果您想使用此版本，请执行 jvms switch %v\n", v)
			} else {
				return fmt.Errorf("无法下载 %s 版本", v)
			}
			return nil
		}
	}

	return errors.New("无效的版本，输入 jvms rls 查看可供安装的版本")
}
