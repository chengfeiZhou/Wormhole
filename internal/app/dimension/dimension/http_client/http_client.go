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

func WithLogger(log logger.Logger) OptionFunc {
	return func(a *Adapter) {
		a.log = log
	}
}

func WithBind(bind string) OptionFunc {
	return func(a *Adapter) {
		a.bind = bind
	}
}

func WithHTTPTimeout(t time.Duration) OptionFunc {
	return func(a *Adapter) {
		a.client.Timeout = t
	}
}

func WithHTTPTransport(trans *http.Transport) OptionFunc {
	return func(a *Adapter) {
		a.client.Transport = trans
	}
}

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
func (a *Adapter) GetName() string {
	return "http"
}
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
func (a *Adapter) Help() {
	// TODO: go模板语法
	fmt.Println("http module help")
}

func (a *Adapter) Transform() structs.TransMessage {
	return structs.TransHTTPMessage
}

// Run 执行监听服务
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
