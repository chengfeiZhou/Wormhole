package messagehandling

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/chengfeiZhou/Wormhole/internal/app/dimension"
	"github.com/chengfeiZhou/Wormhole/internal/pkg/structs"
	"github.com/chengfeiZhou/Wormhole/pkg/logger"
	"github.com/chengfeiZhou/Wormhole/pkg/storage/files"
)

// TODO: 指定以下几个写入feate:
// 1. 每次读一个文件;
// 2. 每次读多个文件
// nolint
type Reader struct {
	log          logger.Logger
	path         string               // 缓存目录
	scanInterval time.Duration        // 目录扫描定时
	msgChan      chan<- interface{}   // 报文转移通道
	transform    structs.TransMessage // 数据转换函数
}

type OptionFuncToRead func(*Reader)

func WithLoggerToRead(log logger.Logger) OptionFuncToRead {
	return func(r *Reader) {
		r.log = log
	}
}

// WithHandlingPath 设置文件搬运地址
func WithHandlingPathToRead(path string) OptionFuncToRead {
	return func(r *Reader) {
		r.path = path
	}
}

// WithScanIntervalToRead 设置搬运目录遍历间隔
func WithScanIntervalToRead(t time.Duration) OptionFuncToRead {
	return func(r *Reader) {
		r.scanInterval = t
	}
}

func NewReader(transform structs.TransMessage, msgChan chan<- interface{}, ops ...OptionFuncToRead) *Reader {
	ad := &Reader{
		log:          logger.DefaultLogger(),
		msgChan:      msgChan,
		scanInterval: 5 * time.Second,
		path:         files.JoinPath(files.RootAbPathByCaller(), "tmp"),
		transform:    transform,
	}
	for _, op := range ops {
		op(ad)
	}
	return ad
}

// nolint
func (r *Reader) GetName() string {
	return "file"
}
func (r *Reader) Setup(app *dimension.App, msgChan chan<- interface{}, transform structs.TransMessage) error {
	r.log = app.Logger
	r.msgChan = msgChan
	r.path = app.Config.HandlingPath
	r.scanInterval = time.Duration(app.Config.ScanInterval) * time.Second
	r.transform = transform
	return nil
}

func (r *Reader) Help() {
	// TODO: go模板语法
	fmt.Println("file bridge help")
}

// Run 读取搬运目录下的可以搬运的文件
func (r *Reader) Run(ctx context.Context) error {
	if err := files.IsNotExistMkDir(r.path); err != nil {
		r.log.Error(logger.ErrorNonExistsFolder, "创建搬运目录", logger.ErrorField(err))
		return err
	}
	tick := time.NewTicker(r.scanInterval)
	r.log.Info("执行文件搬运读取", logger.MakeField("handlingPath", r.path))
	for {
		select {
		case <-tick.C:
			if err := r.handling(); err != nil {
				r.log.Error(logger.ErrorMethod, "文件搬运操作", logger.ErrorField(err))
			}
		case <-ctx.Done():
			r.log.Info("文件搬运读取模块退出")
			return nil
		}
	}
}

func (r *Reader) handling() error {
	rd, err := os.ReadDir(r.path)
	if err != nil {
		r.log.Error(logger.ErrorNonExistsFolder, "目录获取", logger.ErrorField(err))
		return err
	}
	for _, fi := range rd {
		if fi.IsDir() {
			r.log.Warn(logger.ErrorNonExistsFile, "不应该存在的目录", logger.MakeField("folder", fi.Name()))
			continue
		}
		if strings.HasSuffix(fi.Name(), targetExt) {
			go func(filename string) {
				r.log.Info("处理搬运文件", logger.MakeField("filename", filename))
				// 文件是ready的文件
				newName, err := renameReadyFile(files.JoinPath(r.path, filename))
				if err != nil {
					r.log.Error(logger.ErrorWriteFile, "ready文件更名", logger.ErrorField(err), logger.MakeField("ready file", filename))
					return
				}
				if err := r.readFileAndSend(newName); err != nil {
					// 恢复文件名, 等待下次读取
					if _, errF := restoreFileToReady(newName); errF != nil {
						r.log.Error(logger.ErrorWriteFile, "错误文件命名恢复", logger.ErrorField(err), logger.MakeField("restore file", newName))
					}
				}
				// 删除文件
				_ = os.RemoveAll(newName)
			}(fi.Name())
		}
	}
	return nil
}

func (r *Reader) readFileAndSend(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		r.log.Error(logger.ErrorWriteFile, "文件打开失败", logger.ErrorField(err), logger.MakeField("filename", filename))
		return err
	}
	defer f.Close()
	fs, err := f.Stat()
	if err != nil {
		r.log.Error(logger.ErrorWriteFile, "文件信息获取错误", logger.ErrorField(err), logger.MakeField("filename", filename))
		return err
	}
	br := bufio.NewReaderSize(f, int(fs.Size()))
	for {
		// TODO: @zcf 如何优雅的读取大文件?
		line, _, err := br.ReadLine()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			r.log.Error(logger.ErrorReadFile, "read file error", logger.ErrorField(err), logger.MakeField("filename", filename))
			break
		}
		data, err := r.transform(line)
		if err != nil {
			r.log.Error(logger.ErrorWriteFile, "ready文件解析", logger.ErrorField(err), logger.MakeField("line", line))
			continue
		}
		r.msgChan <- data
	}
	return nil
}

// renameReadyFile 对允许搬运的文件重命名
// .hsxa ==> .hsxa_bak
func renameReadyFile(abFilePath string) (string, error) {
	fileTmp := strings.TrimSuffix(abFilePath, targetExt)
	newFileName := fmt.Sprintf("%s%s", fileTmp, bakExt)
	err := os.Rename(abFilePath, newFileName)
	return newFileName, err
}

// restoreFileToReady 将搬运中的文件的文件名恢复到ready
// .hsxa_bak ==> .hsxa
func restoreFileToReady(abFilePath string) (string, error) {
	fileTmp := strings.TrimSuffix(abFilePath, bakExt)
	newFileName := fmt.Sprintf("%s%s", fileTmp, targetExt)
	err := os.Rename(abFilePath, newFileName)
	return newFileName, err
}
