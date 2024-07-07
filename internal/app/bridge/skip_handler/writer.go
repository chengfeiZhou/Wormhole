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

// WithLoggerToWriter 是一个函数选项，用于设置Writer实例的日志记录器。
// 它接收一个实现了logger.Logger接口的对象作为参数，并返回一个OptionFuncToWriter类型的函数。
// 这个返回的函数在被调用时，会将传入的日志记录器赋值给Writer实例的log字段。
func WithLoggerToWriter(log logger.Logger) OptionFuncToWriter {
	return func(w *Writer) {
		w.log = log
	}
}

// NewWriter 返回一个指向Writer类型结构体的指针，该结构体用于写入日志
// ops参数是一个可变参数，类型为OptionFuncToWriter的函数切片，用于对Writer结构体进行配置
func NewWriter(ops ...OptionFuncToWriter) *Writer {
	ad := &Writer{
		log: logger.DefaultLogger(),
	}
	for _, op := range ops {
		op(ad)
	}
	return ad
}

// GetName 返回Writer对象的名称，始终返回字符串"skip"
func (w *Writer) GetName() string {
	return "skip"
}

// Setup 方法用于初始化Writer结构体中的日志记录器
//
// 参数：
//
//	w *Writer：Writer结构体的指针，表示当前Writer实例
//	app *stargate.App：stargate.App结构体的指针，表示应用实例
//	msgChan <-chan []byte：接收消息的通道，这里未使用
//
// 返回值：
//
//	error：返回nil表示初始化成功，否则返回错误信息
func (w *Writer) Setup(app *stargate.App, msgChan <-chan []byte) error {
	w.log = app.Logger
	return nil
}

// Help 是Writer类型的方法，用于打印帮助信息
func (w *Writer) Help() {
	fmt.Println("skip writer help")
}

// Run 方法在给定的上下文 ctx 中运行 Writer 结构体的实例。
// 它打印一条日志消息 "空执行文件搬运写入..."，然后等待 ctx 被取消或超时。
// 当 ctx 完成时，它会打印一条日志消息 "空文件搬运写入模块退出"，并返回 nil 表示没有错误发生。
func (w *Writer) Run(ctx context.Context) error {
	w.log.Info("空执行文件搬运写入...")
	<-ctx.Done()
	w.log.Info("空文件搬运写入模块退出")
	return nil
}
