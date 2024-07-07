package files

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/chengfeiZhou/Wormhole/pkg/times"
)

const (
	B  = 1
	KB = 1024 * B
	MB = 1024 * KB
	GB = 1024 * MB
)

// GetExt 返回文件名的后缀部分（包括点）
func GetExt(fileName string) string {
	return path.Ext(fileName)
}

// CheckExist 检查文件/目录是否存在 存在返回true 不存在返回false
func CheckExist(src string) bool {
	_, err := os.Stat(src)
	if err != nil {
		return os.IsExist(err)
	}
	return true
}

// CheckPermission 检查文件权限
func CheckPermission(src string) bool {
	_, err := os.Stat(src)
	return os.IsPermission(err)
}

// IsNotExistMkDir 如果不存在则新建文件夹
func IsNotExistMkDir(src string) error {
	if !CheckExist(src) {
		fmt.Printf("路径[%s]不存在, 即将创建\n", src)
		if err := MkDir(src); err != nil {
			return err
		}
	}
	return nil
}

// MkDir 新建文件夹
func MkDir(src string) error {
	if err := os.MkdirAll(src, os.ModePerm); err != nil {
		return err
	}
	return nil
}

// Open 打开文件
func Open(name string, flag int, perm os.FileMode) (*os.File, error) {
	f, err := os.OpenFile(name, flag, perm)
	if err != nil {
		return nil, err
	}
	return f, nil
}

// GetType 获取文件类型
func GetType(p string) (string, error) {
	file, err := os.Open(p)
	if err != nil {
		return "", err
	}
	buff := make([]byte, 512)
	if _, err = file.Read(buff); err != nil {
		return "", err
	}
	filetype := http.DetectContentType(buff)
	return filetype, nil
}

// RootAbPathByCaller 获取项目绝对路径（go run）
func RootAbPathByCaller() string {
	var abPath string
	_, filename, _, ok := runtime.Caller(0)
	if ok {
		abPath = filepath.Dir(filename)
	}
	abPath = strings.Split(abPath, "pkg")[0]
	abPath = filepath.Dir(abPath)
	return abPath
}

// GetFileName 生成文件名 模块名称_时间_随机数
func GetFileName(fileName, ext string) string {
	tn := time.Now()
	return fmt.Sprintf("%s_%s_%d%s", fileName, tn.Format(times.TimeFormatFileName), (tn.UnixNano() % 100000), ext)
}

// SaveFile 保存文件
func SaveFile(path string, r io.Reader) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	if _, err = io.Copy(file, r); err != nil {
		return err
	}
	return nil
}

// JoinPath 拼接文件路径
func JoinPath(elem ...string) string {
	return path.Join(elem...)
}
