package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	msghandling "github.com/chengfeiZhou/Wormhole/internal/app/bridge/message_handling"
	skiphandler "github.com/chengfeiZhou/Wormhole/internal/app/bridge/skip_handler"
	"github.com/chengfeiZhou/Wormhole/internal/app/stargate"
	httpserver "github.com/chengfeiZhou/Wormhole/internal/app/stargate/http_server"
	"github.com/chengfeiZhou/Wormhole/pkg/logger"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
)

func main() {
	// 接收系统信号, 通过context关系退出服务
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT)
	defer stop()
	logging, err := logger.NewLogger()
	if err != nil {
		log.Panic("实例化日志实例错误")
		return
	}
	// 匿名函数内部初始化了一个 Gin 服务器，并注册了 pprof 性能分析工具，最后让服务器在 :3000 端口上运行
	go func() {
		app := gin.Default()
		pprof.Register(app) // 性能
		if err = app.Run(":3000"); err != nil {
			panic(err)
		}
	}()
	app := stargate.NewApp(filepath.Base(os.Args[0]), stargate.WithLogger(logging))
	// 注册 Module
	app.AddModule(new(httpserver.Adapter))
	// app.AddModule(new(kafkaconsumer.Adapter))

	// 注册Bridge
	app.AddBridge(new(msghandling.Writer))
	app.AddBridge(new(skiphandler.Writer))
	// 启动参数
	if len(os.Args) < 3 {
		app.Help()
		return
	}
	if err = app.Setup(ctx, os.Args[1], os.Args[2], os.Args[3:]); err != nil {
		app.Help()
		app.Logger.Fatal(logger.ErrorParam, "参数不正确", logger.ErrorField(err))
		return
	}
	if err := app.Run(ctx); err != nil {
		app.Logger.Fatal(logger.ErrorAgentStart, "agent", logger.ErrorField(err))
	}
}
