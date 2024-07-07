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
	msgChan chan<- interface{} // 报文转移通道
}

type OptionFunc func(*Adapter)

func WithLogger(log logger.Logger) OptionFunc {
	return func(a *Adapter) {
		a.log = log
	}
}

func WithListen(bind string) OptionFunc {
	return func(a *Adapter) {
		a.listen = bind
	}
}

func NewAdapter(msgChan chan<- interface{}, ops ...OptionFunc) *Adapter {
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

// Setup 暴露给App的setup
func (a *Adapter) Setup(app *stargate.App, msgChan chan<- interface{}) error {
	a.msgChan = msgChan
	a.log = app.Logger
	a.listen = app.Config.Listen
	return nil
}

func (a *Adapter) GetName() string {
	return "http"
}

// Help 打印help
func (a *Adapter) Help() {
	// TODO: go模板语法
	fmt.Println("http module help")
}

func (a *Adapter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	msg, err := structs.ParseRequest(r)
	if err != nil {
		a.response(w, http.StatusInternalServerError, &types.HttpRespData{
			Code:    -1,
			Message: "请求转发处理失败",
			Data:    map[string]interface{}{"error": err.Error()},
		})
		return
	}
	a.log.Info("接收到请求", logger.MakeField("method", msg.Method), logger.MakeField("remote", msg.RemoteAddr),
		logger.MakeField("url", msg.URL), logger.MakeField("contentLength", msg.ContentLength))
	a.msgChan <- msg
	a.response(w, http.StatusOK, &types.HttpRespData{
		Code:    0,
		Message: "请求转发成功",
		Data:    nil,
	})
}

func (a *Adapter) response(w http.ResponseWriter, code int, rd *types.HttpRespData) {
	w.WriteHeader(code)
	msb, err := json.Marshal(rd)
	if err != nil {
		_, _ = w.Write([]byte(err.Error()))
		return
	}
	_, _ = w.Write(msb)
}

// Run 执行监听服务
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
