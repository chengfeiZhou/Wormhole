package skiphandler

import (
	"context"
	"fmt"

	"github.com/chengfeiZhou/Wormhole/internal/app/dimension"
	"github.com/chengfeiZhou/Wormhole/internal/pkg/structs"
	"github.com/chengfeiZhou/Wormhole/pkg/logger"
)

type Reader struct {
	log logger.Logger
}

type OptionFuncToReader func(*Reader)

func WithLoggerToReader(log logger.Logger) OptionFuncToReader {
	return func(r *Reader) {
		r.log = log
	}
}

func NewReader(ops ...OptionFuncToReader) *Reader {
	ad := &Reader{
		log: logger.DefaultLogger(),
	}
	for _, op := range ops {
		op(ad)
	}
	return ad
}

// nolint
func (r *Reader) GetName() string {
	return "skip"
}
func (r *Reader) Setup(app *dimension.App, msgChan chan<- interface{}, transform structs.TransMessage) error {
	r.log = app.Logger
	return nil
}

func (r *Reader) Help() {
	// TODO: go模板语法
	fmt.Println("skip bridge help")
}

// Run 读取搬运目录下的可以搬运的文件
func (r *Reader) Run(ctx context.Context) error {
	r.log.Info("空执行文件搬运读取...")
	<-ctx.Done()
	r.log.Info("空文件搬运读取模块退出")
	return nil
}
