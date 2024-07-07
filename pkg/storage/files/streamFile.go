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

func (sf *StreamFile) FileSize() int64 {
	fi, _ := sf.file.Stat()
	return fi.Size()
}

func (sf *StreamFile) Name() string {
	return sf.file.Name()
}

func (sf *StreamFile) Close() (*StreamFile, error) {
	sf.lock.Lock()
	defer sf.lock.Unlock()
	sf.write.Flush()
	sf.file.Close()
	tarpath := JoinPath(sf.path, sf.name)

	filePass <- [2]string{sf.file.Name(), tarpath}
	return nil, nil
}

// nolint
func (sf *StreamFile) Write(data []byte) error {
	sf.lock.Lock()
	defer sf.lock.Unlock()
	_, err := sf.write.Write(data)
	defer sf.write.Write([]byte("\n"))
	return err
}

func (sf *StreamFile) Flush() {
	sf.lock.Lock()
	defer sf.lock.Unlock()
	sf.write.Flush()
}

// WriteAt 实现io.WriterAt
func (sf *StreamFile) WriteAt(p []byte, off int64) (n int, err error) {
	return sf.file.WriteAt(p, off)
}
