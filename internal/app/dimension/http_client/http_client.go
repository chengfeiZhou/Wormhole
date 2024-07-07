package httpclient

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/chengfeiZhou/Wormhole/internal/app/dimension"
	"github.com/chengfeiZhou/Wormhole/internal/pkg/structs"
	"github.com/chengfeiZhou/Wormhole/pkg/logger"
)

// nolint
type Adapter struct {
	log     logger.Logger
	bind    string        // 转发目标
	msgChan <-chan []byte // 报文转移通道
	client  *http.Client
}

type OptionFunc func(*Adapter)

// WithLogger 是一个函数选项，用于为Adapter设置日志记录器
//
// 参数：
//
//	log logger.Logger - 要设置的日志记录器
//
// 返回值：
//
//	OptionFunc - 一个函数选项，用于设置Adapter的日志记录器
func WithLogger(log logger.Logger) OptionFunc {
	return func(a *Adapter) {
		a.log = log
	}
}

// WithBind 返回一个OptionFunc类型的函数，该函数用于设置Adapter实例的bind字段
// 参数bind表示需要绑定的地址
func WithBind(bind string) OptionFunc {
	return func(a *Adapter) {
		a.bind = bind
	}
}

// WithHTTPTimeout 返回一个 OptionFunc 类型的函数，该函数用于设置 Adapter 中 HTTP 客户端的超时时间
//
// 参数：
// t time.Duration - 要设置的超时时间
//
// 返回值：
// OptionFunc - 一个函数，该函数接受一个指向 Adapter 的指针，并设置其 HTTP 客户端的超时时间
func WithHTTPTimeout(t time.Duration) OptionFunc {
	return func(a *Adapter) {
		a.client.Timeout = t
	}
}

// WithHTTPTransport 返回一个OptionFunc函数，该函数用于设置Adapter的HTTP传输层。
// trans参数是http.Transport类型的指针，表示要设置的HTTP传输层。
// 返回的OptionFunc函数接受一个*Adapter类型的指针参数a，用于设置其HTTP客户端的传输层。
func WithHTTPTransport(trans *http.Transport) OptionFunc {
	return func(a *Adapter) {
		a.client.Transport = trans
	}
}

// NewAdapter 创建一个新的 Adapter 实例
// msgChan 是一个只读的字节切片通道，用于接收消息
// ops 是一个可变参数列表，用于配置 Adapter 的选项
// 返回一个指向 Adapter 实例的指针
func NewAdapter(msgChan <-chan []byte, ops ...OptionFunc) *Adapter {
	ad := &Adapter{
		log:     logger.DefaultLogger(),
		msgChan: msgChan,
		bind:    "127.0.0.1:8080",
		client:  http.DefaultClient,
	}
	for _, op := range ops {
		op(ad)
	}
	return ad
}

// GetName 返回 Adapter 实例的名称，固定为 "http"。
func (a *Adapter) GetName() string {
	return "http"
}

// Setup 是Adapter类型的方法，用于设置Adapter的配置和依赖项
//
// 参数：
//
//	app *dimension.App：应用实例，用于获取Logger和配置信息
//	msgChan <-chan []byte：消息通道，用于接收消息
//
// 返回值：
//
//	error：如果设置成功，则返回nil；否则返回错误信息
func (a *Adapter) Setup(app *dimension.App, msgChan <-chan []byte) error {
	a.log = app.Logger
	a.bind = app.Config.Bind
	a.msgChan = msgChan
	a.client = &http.Client{
		Timeout:   time.Duration(app.Config.HttpTimeout) * time.Second,
		Transport: http.DefaultTransport,
	}
	return nil
}

// Help 是Adapter结构体上的一个方法，用于打印http模块的帮助信息
func (a *Adapter) Help() {
	// TODO: go模板语法
	fmt.Println("http module help")
}

// Transform 方法返回一个structs.TransMessage类型的值，表示Adapter的转换结果
func (a *Adapter) Transform() structs.TransMessage {
	return structs.TransHTTPMessage
}

// Run 是Adapter类型的方法，用于启动并运行适配器服务
// 它会监听消息通道a.msgChan，并处理其中的HTTP请求消息
// 如果通道关闭或上下文ctx被取消，该方法将结束运行
// 参数：
//
//	ctx: 上下文对象，用于控制Run方法的执行
//
// 返回值：
//
//	error: 如果在Run方法执行过程中发生错误，则返回非零的错误码；否则返回nil
func (a *Adapter) Run(ctx context.Context) error {
	a.log.Info("run service for proxy client for http", logger.MakeField("bind", a.bind))
	for {
		select {
		case data, ok := <-a.msgChan:
			if !ok {
				a.log.Info("搬运请求客户端数据无效")
				continue
			}
			dataE, err := structs.TransHTTPMessage(data)
			if err != nil {
				a.log.Error(logger.ErrorParam, "搬运数据类型错误,无法格式化", logger.ErrorField(err))
				continue
			}
			// TODO: 并发请求, 当前不做限制; 后期可以考虑增加并发池
			go a.sendRequest(ctx, dataE.(*structs.HTTPMessage))
		case <-ctx.Done():
			a.log.Info("搬运请求客户端模块执行结束")
			return nil
		}
	}
}

// sendRequest 是一个Adapter类型的方法，用于发送HTTP请求
//
// 参数：
// ctx: 上下文对象，用于控制请求的取消和超时
// dataE: 指向structs.HTTPMessage类型的指针，包含HTTP请求的所有信息
//
// 返回值：
// 无返回值，该函数为void类型
func (a *Adapter) sendRequest(ctx context.Context, dataE *structs.HTTPMessage) {
	req, err := dataE.MakeRequest(a.bind)
	if err != nil {
		a.log.Error(logger.ErrorRequestExecutor, "请求生成错误", logger.ErrorField(err),
			logger.MakeField("method", dataE.Method), logger.MakeField("URL", dataE.URL))
		return
	}
	resp, err := a.client.Do(req)
	if err != nil {
		a.log.Error(logger.ErrorRequestExecutor, "请求目标错误", logger.ErrorField(err),
			logger.MakeField("method", dataE.Method), logger.MakeField("URL", dataE.URL))
		return
	}
	defer resp.Body.Close()
	a.log.Debugf(func() string {
		data, err := io.ReadAll(resp.Body) // 将struct转成[]byte
		if err != nil {
			return err.Error()
		}
		return string(data)
	}())
	if resp.StatusCode > http.StatusBadRequest {
		a.log.Error(logger.ErrorRequestExecutor, "请求目标错误", logger.MakeField("respStatus", resp.Status),
			logger.MakeField("method", dataE.Method), logger.MakeField("URL", req.URL.String()))
		return
	}
	a.log.Info("请求转发成功", logger.MakeField("method", dataE.Method), logger.MakeField("URL", req.URL.String()),
		logger.MakeField("respStatus", resp.Status))
}
