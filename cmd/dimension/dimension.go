package main

import (
	"context"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	msghandling "github.com/chengfeiZhou/Wormhole/internal/app/bridge/message_handling"
	skiphandler "github.com/chengfeiZhou/Wormhole/internal/app/bridge/skip_handler"
	"github.com/chengfeiZhou/Wormhole/internal/app/dimension"
	httpclient "github.com/chengfeiZhou/Wormhole/internal/app/dimension/dimension/http_client"
	"github.com/chengfeiZhou/Wormhole/pkg/logger"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
)

func main() {
	// 接收系统信号, 通过context关系退出服务
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT)
	defer stop()
	go func() {
		app := gin.Default()
		pprof.Register(app) // 性能
		app.Run(":3000")
	}()
	app := dimension.NewApp(filepath.Base(os.Args[0]))
	// 注册 Module
	app.AddModule(new(httpclient.Adapter))
	// app.AddModule(new(kafkaproducer.Adapter))
	// 注册Bridge
	app.AddBridge(new(msghandling.Reader))
	app.AddBridge(new(skiphandler.Reader))
	// 启动参数
	if len(os.Args) < 3 {
		app.Help()
		return
	}
	if err := app.Setup(ctx, os.Args[1], os.Args[2], os.Args[3:]); err != nil {
		app.Help()
		app.Logger.Fatal(logger.ErrorParam, "参数不正确", logger.ErrorField(err))
		return
	}
	if err := app.Run(ctx); err != nil {
		app.Logger.Fatal(logger.ErrorAgentStart, "agent", logger.ErrorField(err))
	}
}
