package files

import (
	"bufio"
	"context"
	"log"
	"os"
	"os/exec"
	"sync"
)

var filePass chan [2]string // [src, tag]文件关闭时， 调用操作系统mv命令的队列

// init 函数在程序启动时自动执行，用于初始化全局变量和启动后台goroutine
func init() {
	filePass = make(chan [2]string, 5000)
	go func(ctx context.Context) {
		for fileFlag := range filePass {
			src, tag := fileFlag[0], fileFlag[1]
			// err := os.Rename(sf.file.Name(), tarpath) // 移动文件: cache/data.hsxa => target/path/data.hsxa
			if _, err := exec.CommandContext(ctx, "mv", src, tag).Output(); err != nil { // nolint
				log.Printf("移动文件错误: error: %s, srcPath: %s,  targetPath: %s \n",
					err.Error(), src, tag)
			}
		}
	}(context.TODO())
}

// StreamFile 流式写文件
// nolint
type StreamFile struct {
	name  string
	path  string
	file  *os.File
	write *bufio.Writer
	lock  sync.Locker
}

// NewStreamFile 函数用于创建一个StreamFile对象，用于流式写入文件
//
// 参数：
// path: string类型，表示文件的目标路径（不包含文件名）
// filename: string类型，表示文件名
//
// 返回值：
// *StreamFile: StreamFile类型的指针，表示创建的StreamFile对象
// error: 错误对象，如果创建StreamFile对象过程中发生错误，则返回非零的错误码
func NewStreamFile(path, filename string) (*StreamFile, error) {
	cacheFile := JoinPath(os.TempDir(), filename)
	file, err := os.OpenFile(cacheFile, os.O_WRONLY|os.O_CREATE|os.O_SYNC, os.ModePerm)
	if err != nil {
		return nil, err
	}
	sf := &StreamFile{
		name:  filename,
		path:  path,
		file:  file,
		write: bufio.NewWriterSize(file, 10<<20), // 会初始化一个10MB的buffer
		lock:  new(sync.Mutex),
	}
	return sf, nil
}

// FileSize 返回StreamFile实例所代表的文件的大小（以字节为单位）
func (sf *StreamFile) FileSize() int64 {
	fi, _ := sf.file.Stat()
	return fi.Size()
}

// Name 返回StreamFile对象所代表的文件的名称
func (sf *StreamFile) Name() string {
	return sf.file.Name()
}

// Close 关闭StreamFile对象，并将文件添加到tar包中
//
// 参数：
//
//	sf: StreamFile对象指针
//
// 返回值：
//
//	返回值为一个包含两个nil值的tuple，表示没有返回值（实际开发中可能不返回StreamFile对象，而是直接返回error）
//	第二个返回值error表示关闭文件过程中出现的错误，如果成功关闭则为nil
func (sf *StreamFile) Close() (*StreamFile, error) {
	sf.lock.Lock()
	defer sf.lock.Unlock()
	sf.write.Flush()
	sf.file.Close()
	tarpath := JoinPath(sf.path, sf.name)

	filePass <- [2]string{sf.file.Name(), tarpath}
	return nil, nil
}

// Write 方法向StreamFile结构体中的write成员写入数据
// 参数data为待写入的字节切片
// 返回值为写入过程中出现的错误，如果没有错误则为nil
// nolint
func (sf *StreamFile) Write(data []byte) error {
	sf.lock.Lock()
	defer sf.lock.Unlock()
	_, err := sf.write.Write(data)
	defer sf.write.Write([]byte("\n"))
	return err
}

// Flush 方法将StreamFile中的缓存数据写入到底层存储，并清空缓存。
// 该方法是StreamFile类型的方法，通过指针接收器sf访问StreamFile对象。
func (sf *StreamFile) Flush() {
	sf.lock.Lock()
	defer sf.lock.Unlock()
	sf.write.Flush()
}

// WriteAt 实现io.WriterAt
// WriteAt 在指定的偏移量处写入字节切片 p 到 StreamFile 对应的文件中
//
// 参数：
//
//	sf: *StreamFile StreamFile 实例的指针
//	p: []byte 要写入的字节切片
//	off: int64 偏移量，表示从文件开头的字节数
//
// 返回值：
//
//	n: int 成功写入的字节数
//	err: error 如果发生错误，则为非零错误码；否则为零
func (sf *StreamFile) WriteAt(p []byte, off int64) (n int, err error) {
	return sf.file.WriteAt(p, off)
}
