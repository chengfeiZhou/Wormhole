package skiphandler

import (
	"context"
	"fmt"

	"github.com/chengfeiZhou/Wormhole/internal/app/dimension"
	"github.com/chengfeiZhou/Wormhole/pkg/logger"
)

type Reader struct {
	log logger.Logger
}

type OptionFuncToReader func(*Reader)

// WithLoggerToReader 是一个函数选项，用于配置Reader结构体实例的日志记录器
// 参数：
//
//	log logger.Logger - 日志记录器实例
//
// 返回值：
//
//	OptionFuncToReader - 一个函数选项，该函数接受一个*Reader指针作为参数，用于配置该Reader实例的日志记录器
func WithLoggerToReader(log logger.Logger) OptionFuncToReader {
	return func(r *Reader) {
		r.log = log
		// NewReader 创建一个新的Reader对象
		// ops 参数是一个可变参数列表，每个参数都是一个用于配置Reader的函数（OptionFuncToReader）
		// 返回值为指向新创建的Reader对象的指针
	}
}

// NewReader 函数返回一个新的Reader指针，该Reader使用默认的日志记录器logger.DefaultLogger()。
//
// ops是一个可变参数列表，其中的元素是类型为OptionFuncToReader的函数，这些函数用于配置Reader的选项。
// 每个OptionFuncToReader函数都会接受一个指向Reader的指针作为参数，并对其进行修改。
//
// 示例用法：
// reader := NewReader(WithLogger(customLogger))
func NewReader(ops ...OptionFuncToReader) *Reader {
	ad := &Reader{
		log: logger.DefaultLogger(),
	}
	for _, op := range ops {
		op(ad)
	}
	return ad
}

// GetName 返回Reader结构体中存储的name字段的值，
// 但在此示例中，它直接返回了字符串"skip"，
// 实际情况下，GetName方法可能需要根据Reader结构体中的具体字段或状态来返回相应的字符串。
// nolint
func (r *Reader) GetName() string {
	return "skip"
}

// Setup 是Reader类型的方法，用于设置Reader对象的配置
//
// app：一个dimension.App类型的指针，用于提供应用上下文和依赖
// msgChan：一个只写channel，用于传递消息给其他goroutine
//
// 返回值：
// error：如果设置成功，则返回nil；否则返回非nil的错误信息
func (r *Reader) Setup(app *dimension.App, msgChan chan<- []byte) error {
	r.log = app.Logger
	return nil
}

// Help 方法是Reader类型的方法，用于打印帮助信息
// 但目前该方法仅打印固定字符串"skip bridge help"，后续可能需要使用go模板语法进行扩展
// 注意：目前该方法是一个占位符，实际功能可能需要后续实现
func (r *Reader) Help() {
	fmt.Println("skip bridge help")
}

// Run 是Reader类型的方法，用于启动文件读取操作
//
// 参数：
//
//	ctx context.Context：上下文对象，用于控制函数执行流程
//
// 返回值：
//
//	error：如果函数执行过程中发生错误，则返回非零的错误码；否则返回nil
func (r *Reader) Run(ctx context.Context) error {
	r.log.Info("空执行文件搬运读取...")
	<-ctx.Done()
	r.log.Info("空文件搬运读取模块退出")
	return nil
}
