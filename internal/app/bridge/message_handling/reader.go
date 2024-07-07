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
	"github.com/chengfeiZhou/Wormhole/pkg/logger"
	"github.com/chengfeiZhou/Wormhole/pkg/storage/files"
)

// TODO: 指定以下几个写入feate:
// 1. 每次读一个文件;
// 2. 每次读多个文件
// nolint
type Reader struct {
	log          logger.Logger
	path         string        // 缓存目录
	scanInterval time.Duration // 目录扫描定时
	msgChan      chan<- []byte // 报文转移通道
}

type OptionFuncToRead func(*Reader)

// WithLoggerToRead 返回一个用于配置 Reader 日志选项的函数。
// log 参数是 logger.Logger 类型的日志记录器。
// 该函数返回一个 OptionFuncToRead 类型的函数，该函数接受一个 Reader 指针作为参数，
// 并将其日志记录器设置为提供的 log 参数。
func WithLoggerToRead(log logger.Logger) OptionFuncToRead {
	return func(r *Reader) {
		r.log = log
	}
}

// WithHandlingPathToRead 是一个返回OptionFuncToRead的函数，用于设置Reader实例的读取路径
//
// 参数：
//
//	path string - 要读取的文件的路径
//
// 返回值：
//
//	OptionFuncToRead - 一个函数选项，用于设置Reader实例的读取路径
func WithHandlingPathToRead(path string) OptionFuncToRead {
	return func(r *Reader) {
		r.path = path
	}
}

// WithScanIntervalToRead 返回一个OptionFuncToRead类型的函数，设置搬运目录遍历间隔
// 参数t表示扫描间隔时间，类型为time.Duration
// 返回的OptionFuncToRead类型的函数用于修改Reader结构体的scanInterval字段
func WithScanIntervalToRead(t time.Duration) OptionFuncToRead {
	return func(r *Reader) {
		r.scanInterval = t
	}
}

// NewReader 创建一个新的Reader实例
//
// msgChan：用于发送消息的通道
// ops：可变参数列表，用于配置Reader的选项函数
//
// 返回值：
// *Reader：指向新创建的Reader实例的指针
func NewReader(msgChan chan<- []byte, ops ...OptionFuncToRead) *Reader {
	ad := &Reader{
		log:          logger.DefaultLogger(),
		msgChan:      msgChan,
		scanInterval: 5 * time.Second,
		path:         files.JoinPath(files.RootAbPathByCaller(), "tmp"),
	}
	for _, op := range ops {
		op(ad)
	}
	return ad
}

// GetName 是Reader结构体上的一个方法，它返回一个字符串表示的名称。
//
// 对于这个特定的Reader实现，它总是返回"file"。
//
// 参数：
//
//	r *Reader - 指向Reader结构体的指针，表示要获取名称的Reader实例。
//
// 返回值：
//
//	string - 返回一个表示Reader名称的字符串，此处总是"file"。
//
// nolint
func (r *Reader) GetName() string {
	return "file"
}

// Setup 方法为 Reader 类型接收者 r 提供一个配置函数，用于设置相关属性和初始化工作。
// 参数：
//
//	app: *dimension.App - 应用程序实例，用于获取日志记录器和配置信息。
//	msgChan: chan<- []byte - 消息通道，用于传递处理后的消息。
//
// 返回值：
//
//	error - 如果没有错误则返回 nil，否则返回相应的错误信息。
func (r *Reader) Setup(app *dimension.App, msgChan chan<- []byte) error {
	r.log = app.Logger
	r.msgChan = msgChan
	r.path = app.Config.HandlingPath
	r.scanInterval = time.Duration(app.Config.ScanInterval) * time.Second
	return nil
}

// Help 是Reader类型的方法，用于打印帮助信息
func (r *Reader) Help() {
	// TODO: go模板语法
	fmt.Println("file bridge help")
}

// Run 读取搬运目录下的可以搬运的文件
// Run 方法在给定的context.Context中运行Reader结构体实例的文件搬运读取操作
// 如果Reader结构体实例的path字段指定的目录不存在，则尝试创建该目录
// 若目录创建失败，则记录错误信息并返回error
// Run方法启动一个周期性检查器（定时器），时间间隔由Reader结构体实例的scanInterval字段指定
// 在每个时间间隔内，Run方法会调用handling方法执行文件搬运操作
// 若handling方法执行失败，则记录错误信息
// 当context.Context被取消时，Run方法会停止执行并返回nil
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

// handling 是Reader类型的方法，用于处理目录中的文件
// 如果目录不存在或处理文件时发生错误，则返回非零错误码
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

// readFileAndSend 读取指定文件，并将文件内容处理后发送至消息通道
//
// 参数：
//
//	r *Reader - Reader结构体指针，包含日志记录器和消息通道
//	filename string - 要读取的文件名
//
// 返回值：
//
//	error - 读取文件过程中出现的错误，如果成功则为nil
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
		r.msgChan <- line
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
