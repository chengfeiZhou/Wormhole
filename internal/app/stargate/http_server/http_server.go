package httpserver

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/chengfeiZhou/Wormhole/internal/app/stargate"
	"github.com/chengfeiZhou/Wormhole/internal/pkg/structs"
	"github.com/chengfeiZhou/Wormhole/pkg/logger"
	types "github.com/chengfeiZhou/Wormhole/pkg/structs"
)

// nolint
type Adapter struct {
	log     logger.Logger
	listen  string
	msgChan chan<- []byte // 报文转移通道
}

type OptionFunc func(*Adapter)

// WithLogger 是一个函数选项，用于设置Adapter的日志记录器
//
// log 参数表示要设置的日志记录器
//
// 返回值是一个OptionFunc类型的函数，该函数接收一个指向Adapter的指针作为参数
// 在函数内部，将参数a的log字段设置为传入的log参数
func WithLogger(log logger.Logger) OptionFunc {
	return func(a *Adapter) {
		a.log = log
	}
}

// WithListen 是一个返回 OptionFunc 类型的函数，该函数接受一个字符串参数 bind，
// 用于设置 Adapter 实例的监听地址。
func WithListen(bind string) OptionFunc {
	return func(a *Adapter) {
		a.listen = bind
	}
}

// NewAdapter 函数用于创建一个新的Adapter实例
// msgChan 是一个通道，用于发送字节切片类型的消息
// ops 是一个可变参数列表，用于传递多个OptionFunc类型的函数选项
// 函数返回一个指向新创建的Adapter实例的指针
func NewAdapter(msgChan chan<- []byte, ops ...OptionFunc) *Adapter {
	ad := &Adapter{
		log:     logger.DefaultLogger(),
		msgChan: msgChan,
		listen:  "0.0.0.0:8080",
	}
	for _, op := range ops {
		op(ad)
	}
	return ad
}

// Setup 是Adapter类型的方法，用于设置Adapter的属性。
// 它接收两个参数：
//   - app：类型为*stargate.App的指针，表示应用程序的实例
//   - msgChan：类型为chan<- []byte的通道，用于接收消息
//
// 函数将msgChan赋值给a的msgChan属性，
// 将app的Logger赋值给a的log属性，
// 将app.Config.Listen赋值给a的listen属性，
// 函数返回值为error类型，但在本例中始终返回nil，表示设置成功。
func (a *Adapter) Setup(app *stargate.App, msgChan chan<- []byte) error {
	a.msgChan = msgChan
	a.log = app.Logger
	a.listen = app.Config.Listen
	return nil
}

// GetName 方法返回 Adapter 对象的名称，这里是 "http"。
// 这是一个属于 Adapter 结构体类型的方法。
func (a *Adapter) GetName() string {
	return "http"
}

// Help 是Adapter结构体中的一个方法，用于打印http模块的帮助信息
func (a *Adapter) Help() {
	// TODO: go模板语法
	fmt.Println("http module help")
}

// ServeHTTP 是Adapter类型的方法，用于处理HTTP请求
// 它接受两个参数：
//
//	w: http.ResponseWriter，用于向客户端发送响应
//	r: *http.Request，表示客户端发送的HTTP请求
func (a *Adapter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	msg, err := structs.ParseRequest(r)
	if err != nil {
		a.response(w, http.StatusInternalServerError, &types.HttpRespData{
			Code:    -1,
			Message: "请求转发处理失败",
			Data:    map[string]any{"error": err.Error()},
		})
		return
	}
	msgB, err := msg.Marshal()
	if err != nil {
		a.response(w, http.StatusInternalServerError, &types.HttpRespData{
			Code:    -1,
			Message: "请求转发处理失败",
			Data:    map[string]any{"error": err.Error()},
		})
		return
	}
	a.msgChan <- msgB
	a.log.Info("接收到请求", logger.MakeField("method", msg.Method), logger.MakeField("remote", msg.RemoteAddr),
		logger.MakeField("url", msg.URL), logger.MakeField("contentLength", msg.ContentLength))

	a.response(w, http.StatusOK, &types.HttpRespData{
		Code:    0,
		Message: "请求转发成功",
		Data:    nil,
	})
}

// response 是Adapter结构体的方法，用于将HTTP响应数据写入到http.ResponseWriter中
//
// 参数：
//
//	w http.ResponseWriter：HTTP响应的写入对象
//	code int：HTTP响应的状态码
//	rd *types.HttpRespData：要写入HTTP响应的HTTP响应数据结构体指针
//
// 返回值：
//
//	无返回值
func (a *Adapter) response(w http.ResponseWriter, code int, rd *types.HttpRespData) {
	w.WriteHeader(code)
	msb, err := json.Marshal(rd)
	if err != nil {
		_, _ = w.Write([]byte(err.Error()))
		return
	}
	_, _ = w.Write(msb)
}

// Run 是Adapter类型的方法，用于启动代理服务
// 接收一个context.Context类型的参数ctx，用于控制服务的生命周期
// 返回值为error类型，表示服务启动过程中出现的错误，如果没有错误则为nil
func (a *Adapter) Run(ctx context.Context) error {
	a.log.Info("run service for proxy", logger.MakeField("listen", a.listen))
	svc := &http.Server{
		Addr:           a.listen,
		Handler:        a,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20}
	if err := svc.ListenAndServe(); err != nil {
		a.log.Fatal(logger.ErrorHTTPHandle, "conversion of Entry Service", logger.ErrorField(err))
		return err
	}
	return nil
}
