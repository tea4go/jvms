package cmdCli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

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
		return errors.New("无效的版本，输入 \"jvms rls\" 查看可供安装的版本")
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
			fmt.Printf("正在下载 JDK: %s\n", version.Url)
			dlzipfile, success := web.GetJDK(cfx.Download, v, version.Url)
			if success {
				fmt.Printf("正在安装 JDK %s ...\n", v)

				// 解压 JDK 到临时目录
				jdktempfile := filepath.Join(cfx.Download, fmt.Sprintf("%s_temp", v))
				if file.Exists(jdktempfile) {
					err := os.RemoveAll(jdktempfile)
					if err != nil {
						panic(err)
					}
				}
				err := file.Extract(dlzipfile, jdktempfile)
				if err != nil {
					return fmt.Errorf("解压失败: %w", err)
				}

				// 复制 JDK 文件到安装目录
				temJavaHome := getJavaHome(jdktempfile)
				if temJavaHome == "" {
					return fmt.Errorf("未找到有效的 JDK 目录，请检查%s", jdktempfile)
				}
				err = os.Rename(temJavaHome, filepath.Join(cfx.Store, v))
				if err != nil {
					return fmt.Errorf("移动文件失败: %w", err)
				}

				// 删除临时目录
				// 可以考虑保留临时文件
				err = os.RemoveAll(jdktempfile)
				if err != nil {
					fmt.Printf("警告: 清理临时目录失败: %v\n", err)
				}

				fmt.Printf("安装成功完成。如果您想使用此版本，请使用: jvms switch %v", v)
			} else {
				return fmt.Errorf("无法下载 JDK %s 可执行文件", v)
			}
			return nil
		}
	}

	return errors.New("无效的版本，输入 \"jvms rls\" 查看可供安装的版本")
}
