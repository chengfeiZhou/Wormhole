package skiphandler

import (
	"context"
	"fmt"

	"github.com/chengfeiZhou/Wormhole/internal/app/stargate"
	"github.com/chengfeiZhou/Wormhole/pkg/logger"
)

type Writer struct {
	log logger.Logger
}

type OptionFuncToWriter func(*Writer)

func WithLoggerToWriter(log logger.Logger) OptionFuncToWriter {
	return func(w *Writer) {
		w.log = log
	}
}

func NewWriter(ops ...OptionFuncToWriter) *Writer {
	ad := &Writer{
		log: logger.DefaultLogger(),
	}
	for _, op := range ops {
		op(ad)
	}
	return ad
}

func (w *Writer) GetName() string {
	return "skip"
}

// Setup 暴露给App的setup
func (w *Writer) Setup(app *stargate.App, msgChan <-chan interface{}) error {
	w.log = app.Logger
	return nil
}
func (w *Writer) Help() {
	fmt.Println("skip writer help")
}

func (w *Writer) Run(ctx context.Context) error {
	w.log.Info("空执行文件搬运写入...")
	<-ctx.Done()
	w.log.Info("空文件搬运写入模块退出")
	return nil
}
