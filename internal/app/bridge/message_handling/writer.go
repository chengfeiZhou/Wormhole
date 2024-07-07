package messagehandling

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/chengfeiZhou/Wormhole/internal/app/stargate"
	"github.com/chengfeiZhou/Wormhole/pkg/logger"
	"github.com/chengfeiZhou/Wormhole/pkg/storage/files"
)

const (
	targetExt = ".hole"     // 可用的搬运文件后缀
	bakExt    = ".hole_bak" // 正在读的搬运文件后缀
)

// TODO: 指定以下几个写入feate:
// 1. 每个message写入一个文件;
// 2. 多个message写入一个文件, 每1秒钟检查一次
// nolint
type Writer struct {
	log         logger.Logger
	writeTick   time.Duration
	bufMaxSize  int
	fileMaxSize int64
	path        string
	msgChan     <-chan []byte // 报文转移通道
}

type OptionFuncToWriter func(*Writer)

// WithLoggerToWriter 是一个函数选项，用于设置Writer实例的日志记录器
// 参数log是logger.Logger类型，表示要设置的日志记录器
// 返回值是OptionFuncToWriter类型，表示一个函数选项，用于配置Writer实例
func WithLoggerToWriter(log logger.Logger) OptionFuncToWriter {
	return func(w *Writer) {
		w.log = log
	}
}

// WithHandlingPathToWriter 返回一个函数选项，该函数选项用于设置文件搬运地址
// path：要处理的文件路径
// 返回值：一个OptionFuncToWriter类型的函数选项
func WithHandlingPathToWriter(path string) OptionFuncToWriter {
	return func(w *Writer) {
		w.path = path
	}
}

// WithWriterTicker 函数接收一个时间间隔参数t，并返回一个OptionFuncToWriter类型的函数
// 该函数用于设置文件写入间隔
func WithWriterTicker(t time.Duration) OptionFuncToWriter {
	return func(w *Writer) {
		w.writeTick = t
	}
}

// WithFileSize 是一个返回 OptionFuncToWriter 类型函数的函数，用于设置 Writer 的文件最大大小。
// 参数 s 是文件最大大小的 int64 类型的值。
// 返回的 OptionFuncToWriter 类型的函数接受一个指向 Writer 类型的指针 w，
// 并将其 fileMaxSize 字段设置为参数 s 的值。
func WithFileSize(s int64) OptionFuncToWriter {
	return func(w *Writer) {
		w.fileMaxSize = s
	}
}

// NewWriter 函数用于创建一个Writer实例，它接受一个消息通道msgChan作为输入，并通过该通道接收要写入的数据。
// 它还接受一个可变参数ops，这些参数为OptionFuncToWriter类型，用于配置Writer实例的选项。
// 返回值为指向Writer实例的指针。
func NewWriter(msgChan <-chan []byte, ops ...OptionFuncToWriter) *Writer {
	ad := &Writer{
		log:         logger.DefaultLogger(),
		msgChan:     msgChan,
		writeTick:   time.Second,
		fileMaxSize: 10 << 20, // 10MB
		path:        files.JoinPath(files.RootAbPathByCaller(), "tmp"),
	}
	for _, op := range ops {
		op(ad)
	}
	return ad
}

// GetName 方法返回Writer类型实例的名称，该名称固定为"file"
func (w *Writer) GetName() string {
	return "file"
}

// Setup 方法用于初始化Writer结构体中的相关字段
//
// app参数为stargate.App类型，表示应用程序的实例
// msgChan参数为<-chan []byte类型，表示接收消息的通道
//
// 该方法会执行以下操作：
// 1. 将app.Config.FileMaxSize的值赋给w.fileMaxSize，用于设置文件最大大小
// 2. 将app.Logger的值赋给w.log，用于设置日志记录器
// 3. 将msgChan通道赋值给w.msgChan，用于接收消息
// 4. 将app.Config.HandlingPath的值赋给w.path，用于设置处理路径
// 5. 将app.Config.WriterTicker的值转换为time.Duration类型，并乘以time.Second，然后赋值给w.writeTick，用于设置写入文件的计时器间隔
//
// 返回值为error类型，如果初始化成功则返回nil，否则返回相应的错误信息
func (w *Writer) Setup(app *stargate.App, msgChan <-chan []byte) error {
	w.fileMaxSize = app.Config.FileMaxSize
	w.log = app.Logger
	w.msgChan = msgChan
	w.path = app.Config.HandlingPath
	w.writeTick = time.Duration(app.Config.WriterTicker) * time.Second
	return nil
}

// Help 是Writer结构体中的方法，用于打印writer的帮助信息
func (w *Writer) Help() {
	fmt.Println("writer help")
}

// RunSimple 是Writer结构体中的方法，用于执行简单的文件写入操作
// ctx：上下文对象，用于控制goroutine的生命周期
// 返回值：
//
//	error：如果发生错误，则返回非零的错误码；否则返回nil
func (w *Writer) RunSimple(ctx context.Context) error {
	if err := files.IsNotExistMkDir(w.path); err != nil {
		w.log.Error(logger.ErrorNonExistsFolder, "创建搬运目录", logger.ErrorField(err))
		return err
	}
	w.log.Info("执行文件搬运写入", logger.MakeField("handlingPath", w.path))
	for {
		select {
		case data, ok := <-w.msgChan:
			if !ok {
				continue
			}
			filename, err := w.writFileOnce(data)
			if err != nil {
				w.log.Error(logger.ErrorWriteFile, "搬运报文数据写入", logger.ErrorField(err))
				continue
			}
			w.log.Info("写入搬运文件缓存", logger.MakeField("filename", filename))
		case <-ctx.Done():
			w.log.Info("文件搬运写入模块退出")
			return nil
		}
	}
}

// Run 是Writer类型的方法，用于在给定的上下文ctx中执行文件搬运写入操作。
// 如果写入过程中发生错误，将返回非零的错误码。
func (w *Writer) Run(ctx context.Context) error {
	if err := files.IsNotExistMkDir(w.path); err != nil {
		w.log.Error(logger.ErrorNonExistsFolder, "创建搬运目录", logger.ErrorField(err))
		return err
	}
	w.log.Info("执行文件搬运写入", logger.MakeField("handlingPath", w.path))
	tick := time.NewTicker(w.writeTick)
	var (
		sf  *files.StreamFile
		err error
	)
	for {
		select {
		case data, ok := <-w.msgChan:
			if !ok {
				continue
			}
			if sf == nil {
				if sf, err = files.NewStreamFile(w.path, files.GetFileName("message", targetExt)); err != nil {
					w.log.Error(logger.ErrorWriteFile, "创建搬运文件", logger.ErrorField(err))
					continue
				}
			}
			// nolint
			if err := sf.Write(data); err != nil {
				w.log.Error(logger.ErrorWriteFile, "写入搬运文件缓存", logger.ErrorField(err), logger.MakeField("filePath", sf.Name()))
				continue
			}
			w.log.Info("写入搬运文件缓存", logger.MakeField("filename", sf.Name()))
			if sf.FileSize() >= w.fileMaxSize {
				// 文件大于w.fileMaxSize, 写入文件并重置ticker
				sf, err = sf.Close()
				if err != nil {
					w.log.Error(logger.ErrorWriteFile, "文件关闭", logger.ErrorField(err))
				}
				tick.Reset(w.writeTick)
			}
		case <-tick.C:
			// 判断file对象是否存在
			if sf == nil {
				continue
			}
			w.log.Info("搬运文件写入", logger.MakeField("cachefile", sf.Name()))
			var err error
			sf, err = sf.Close()
			if err != nil {
				w.log.Error(logger.ErrorWriteFile, "文件关闭", logger.ErrorField(err))
			}
		case <-ctx.Done():
			w.log.Info("文件搬运写入模块退出")
			return nil
		}
	}
}

// writFileOnce 写入文件一次，并返回文件路径和可能发生的错误
//
// 参数：
//
//	w *Writer：Writer类型指针，用于获取文件写入路径
//	data []byte：要写入文件的数据
//
// 返回值：
//
//	string：写入文件的路径
//	error：写入文件过程中可能发生的错误
func (w *Writer) writFileOnce(data []byte) (string, error) {
	path := files.JoinPath(w.path, files.GetFileName("message", targetExt))
	reader := bytes.NewReader(data)
	return path, files.SaveFile(path, reader)
}
