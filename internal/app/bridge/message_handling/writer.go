package messagehandling

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/chengfeiZhou/Wormhole/internal/app/stargate"
	"github.com/chengfeiZhou/Wormhole/internal/pkg/structs"
	"github.com/chengfeiZhou/Wormhole/pkg/logger"
	"github.com/chengfeiZhou/Wormhole/pkg/storage/files"
)

const (
	targetExt = ".hsxa"     // 可用的搬运文件后缀
	bakExt    = ".hsxa_bak" // 正在读的搬运文件后缀
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
	msgChan     <-chan interface{} // 报文转移通道
}

type OptionFuncToWriter func(*Writer)

func WithLoggerToWriter(log logger.Logger) OptionFuncToWriter {
	return func(w *Writer) {
		w.log = log
	}
}

// WithHandlingPath 设置文件搬运地址
func WithHandlingPathToWriter(path string) OptionFuncToWriter {
	return func(w *Writer) {
		w.path = path
	}
}

// WithWriterTicker 设置文件写入间隔
func WithWriterTicker(t time.Duration) OptionFuncToWriter {
	return func(w *Writer) {
		w.writeTick = t
	}
}

func WithFileSize(s int64) OptionFuncToWriter {
	return func(w *Writer) {
		w.fileMaxSize = s
	}
}

func NewWriter(msgChan <-chan interface{}, ops ...OptionFuncToWriter) *Writer {
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

func (w *Writer) GetName() string {
	return "file"
}

// Setup 暴露给App的setup
func (w *Writer) Setup(app *stargate.App, msgChan <-chan interface{}) error {
	w.fileMaxSize = app.Config.FileMaxSize
	w.log = app.Logger
	w.msgChan = msgChan
	w.path = app.Config.HandlingPath
	w.writeTick = time.Duration(app.Config.WriterTicker) * time.Second
	return nil
}
func (w *Writer) Help() {
	fmt.Println("writer help")
}

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
			dataB, err := data.(structs.Message).Marshal()
			if err != nil {
				w.log.Error(logger.ErrorWriteFile, "搬运报文数据解析", logger.ErrorField(err))
				continue
			}
			filename, err := w.writFileOnce(dataB)
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

// Run 将需要搬运的数据写入指定目录下
func (w *Writer) Run(ctx context.Context) error {
	if err := files.IsNotExistMkDir(w.path); err != nil {
		w.log.Error(logger.ErrorNonExistsFolder, "创建搬运目录", logger.ErrorField(err))
		return err
	}
	w.log.Info("执行文件搬运写入", logger.MakeField("handlingPath", w.path))
	tick := time.NewTicker(w.writeTick)
	var (
		sf *files.StreamFile
	)
	for {
		select {
		case data, ok := <-w.msgChan:
			if !ok {
				continue
			}
			dataB, err := data.(structs.Message).Marshal()
			if err != nil {
				w.log.Error(logger.ErrorWriteFile, "搬运报文数据解析", logger.ErrorField(err))
				continue
			}
			if sf == nil {
				if sf, err = files.NewStreamFile(w.path, files.GetFileName("message", targetExt)); err != nil {
					w.log.Error(logger.ErrorWriteFile, "创建搬运文件", logger.ErrorField(err))
					continue
				}
			}
			// nolint
			if err := sf.Write(dataB); err != nil {
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

// writFileOnce 报文单文件写入
func (w *Writer) writFileOnce(data []byte) (string, error) {
	path := files.JoinPath(w.path, files.GetFileName("message", targetExt))
	reader := bytes.NewReader(data)
	return path, files.SaveFile(path, reader)
}
