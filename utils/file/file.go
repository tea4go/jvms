package file

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Unzip 解压缩 zip 文件到指定目录
// 参数:
//
//	src - zip 文件的源路径
//	dest - 解压缩的目标目录路径
//
// 返回值:
//
//	error - 解压过程中的错误，成功则返回 nil
//
// 函数来源: http://stackoverflow.com/users/1129149/swtdrgn
func Unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer rc.Close()

		fpath := filepath.Join(dest, f.Name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, f.Mode())
		} else {
			var fdir string
			if lastIndex := strings.LastIndex(fpath, string(os.PathSeparator)); lastIndex > -1 {
				fdir = fpath[:lastIndex]
			}

			err = os.MkdirAll(fdir, f.Mode())
			if err != nil {
				return err
			}
			f, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer f.Close()

			_, err = io.Copy(f, rc)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// Untar 解压 .tar.gz 文件
func Untar(src, dest string) error {
	file, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("创建 gzip reader 失败: %w", err)
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			fmt.Printf("解压完成\n")
			break // 读取完成
		}
		if err != nil {
			return fmt.Errorf("读取 tar header 失败: %w", err)
		}

		target := filepath.Join(dest, header.Name)

		// 安全检查：防止路径穿越攻击
		cleanDest := filepath.Clean(dest) + string(os.PathSeparator)
		cleanTarget := filepath.Clean(target)
		if !strings.HasPrefix(cleanTarget, cleanDest) && cleanTarget != filepath.Clean(dest) {
			fmt.Printf("警告: 跳过非法路径: %s\n", header.Name)
			continue
		}

		switch header.Typeflag {
		case tar.TypeDir:
			// 创建目录
			if err := os.MkdirAll(target, 0755); err != nil {
				return fmt.Errorf("创建目录失败 %s: %w", target, err)
			}

		case tar.TypeReg:
			// 创建文件
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return fmt.Errorf("创建父目录失败 %s: %w", filepath.Dir(target), err)
			}

			outFile, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return fmt.Errorf("创建文件失败 %s: %w", target, err)
			}

			written, err := io.Copy(outFile, tr)
			outFile.Close()

			if err != nil {
				return fmt.Errorf("写入文件失败 %s: %w", target, err)
			}

			if written != header.Size {
				return fmt.Errorf("文件大小不匹配 %s: 期望 %d, 实际 %d", target, header.Size, written)
			}

		case tar.TypeSymlink:
			// 创建符号链接
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return fmt.Errorf("创建符号链接父目录失败 %s: %w", filepath.Dir(target), err)
			}

			// 如果符号链接已存在，先删除
			if _, err := os.Lstat(target); err == nil {
				if err := os.Remove(target); err != nil {
					return fmt.Errorf("删除已存在的符号链接失败 %s: %w", target, err)
				}
			}

			if err := os.Symlink(header.Linkname, target); err != nil {
				fmt.Printf("警告: 创建符号链接失败 %s -> %s: %v\n", target, header.Linkname, err)
			}

		default:
			fmt.Printf("警告: 跳过不支持的文件类型 %c: %s\n", header.Typeflag, header.Name)
		}
	}

	return nil
}

// Extract 自动识别并解压文件（支持 .zip 和 .tar.gz）
func Extract(src, dest string) error {
	if strings.HasSuffix(src, ".zip") {
		return Unzip(src, dest)
	} else if strings.HasSuffix(src, ".tar.gz") || strings.HasSuffix(src, ".tgz") {
		return Untar(src, dest)
	}
	return fmt.Errorf("不支持的压缩格式: %s", src)
}

// Exists 检查文件或目录是否存在
// 参数:
//
//	filename - 要检查的文件或目录路径
//
// 返回值:
//
//	bool - 存在返回 true，否则返回 false
func Exists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

// GetCurrentPath 获取当前可执行文件所在的目录路径
// 返回值:
//
//	string - 当前可执行文件的目录路径，获取失败则返回空字符串
func GetCurrentPath() string {
	currentDir, err := os.Executable()
	if err != nil {
		return ""
	}
	return filepath.Dir(currentDir)
}
