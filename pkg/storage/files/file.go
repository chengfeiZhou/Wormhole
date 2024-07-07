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

// CheckExist 函数用于检查给定路径src的文件或目录是否存在
// 如果存在，则返回true；否则返回false
// 参数：
//
//	src string - 需要检查的文件或目录的路径
//
// 返回值：
//
//	bool - 文件或目录是否存在的布尔值
func CheckExist(src string) bool {
	_, err := os.Stat(src)
	if err != nil {
		return os.IsExist(err)
	}
	return true
}

// CheckPermission 函数用于检查指定路径src的文件或目录的权限
// 如果src指定的文件或目录存在且当前用户没有足够的权限进行访问，则返回true
// 如果src指定的文件或目录不存在，或者当前用户有足够的权限进行访问，则返回false
// 参数：
//
//	src string - 要检查的文件或目录的路径
//
// 返回值：
//
//	bool - 表示是否没有足够的权限进行访问
func CheckPermission(src string) bool {
	_, err := os.Stat(src)
	return os.IsPermission(err)
}

// IsNotExistMkDir 检查给定路径是否存在，若不存在则创建该路径
//
// 参数：
//
//	src string - 需要检查的路径
//
// 返回值：
//
//	error - 如果路径不存在且创建失败，则返回非零错误码；否则返回nil
func IsNotExistMkDir(src string) error {
	if !CheckExist(src) {
		fmt.Printf("路径[%s]不存在, 即将创建\n", src)
		if err := MkDir(src); err != nil {
			return err
		}
	}
	return nil
}

// MkDir 函数用于创建一个目录（包括其父目录），如果目录已存在则不会报错
//
// 参数：
//
//	src string - 要创建的目录路径
//
// 返回值：
//
//	error - 如果创建目录过程中发生错误，则返回非零的错误码；否则返回nil
func MkDir(src string) error {
	if err := os.MkdirAll(src, os.ModePerm); err != nil {
		return err
	}
	return nil
}

// Open 函数尝试以指定的权限和模式打开名为name的文件，
// 并返回一个*os.File类型的指针和可能产生的错误。
// 如果文件成功打开，则返回一个指向文件的指针和一个nil错误。
// 如果打开文件时发生错误，则返回nil和一个非nil的错误。
//
// 参数：
//
//	name string - 要打开的文件名
//	flag int    - 打开文件的标志（如os.O_RDONLY, os.O_WRONLY等）
//	perm os.FileMode - 如果文件需要创建，则指定新文件的权限
//
// 返回值：
//
//	*os.File - 成功打开的文件指针，如果发生错误则为nil
//	error    - 如果在打开文件时发生错误，则为非nil的错误；否则为nil
func Open(name string, flag int, perm os.FileMode) (*os.File, error) {
	f, err := os.OpenFile(name, flag, perm)
	if err != nil {
		return nil, err
	}
	return f, nil
}

// GetType 根据文件路径 p 返回文件的 MIME 类型和可能发生的错误
//
// 参数：
//
//	p string - 文件路径
//
// 返回值：
//
//	string - 文件的 MIME 类型
//	error - 可能发生的错误
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

// RootAbPathByCaller 返回调用此函数的代码的根绝对路径
//
// 注意：此函数通过调用 runtime.Caller(0) 来获取调用者的文件名，然后
// 提取出该文件所在的绝对路径，并去掉 "pkg" 及其之后的部分（假设代码位于 pkg 目录下）。
// 最后，返回去掉 "pkg" 后的上一级目录的绝对路径。
//
// 返回值：
//
//	string：调用者的根绝对路径
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

// GetFileName 函数根据给定的文件名（fileName）和扩展名（ext）生成一个新的文件名
//
// 参数：
// fileName string - 要生成新文件名的原始文件名
// ext string - 文件扩展名
//
// 返回值：
// string - 生成的新文件名，格式为：fileName_时间戳_随机数字_ext
// 其中，时间戳按照times.TimeFormatFileName的格式进行格式化，随机数字是纳秒级时间戳取模100000的结果
func GetFileName(fileName, ext string) string {
	tn := time.Now()
	return fmt.Sprintf("%s_%s_%d%s", fileName, tn.Format(times.TimeFormatFileName), (tn.UnixNano() % 100000), ext)
}

// SaveFile 将读取器r中的数据保存到指定的文件路径path中
// 如果文件创建失败或写入过程中发生错误，则返回相应的错误
// 参数：
//   - path: 要保存的文件路径
//   - r: 读取器，用于读取要写入文件的数据
//
// 返回值：
//   - error: 如果保存文件成功，则返回nil；否则返回错误信息
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

// JoinPath 函数将多个路径元素合并成一个路径字符串。
// elem 是一个可变参数，表示要合并的路径元素。
// 函数返回合并后的路径字符串。
func JoinPath(elem ...string) string {
	return path.Join(elem...)
}
