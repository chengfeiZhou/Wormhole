package stargate

import (
	"context"
	"fmt"

	config "github.com/chengfeiZhou/Wormhole/configs/stargate"
	"github.com/chengfeiZhou/Wormhole/pkg/logger"
)

// Module 接口定义了Stargate模块的接口。
type Module interface {
	GetName() string
	Setup(*App, chan<- []byte) error
	Run(context.Context) error
	Help()
}

// Bridge 接口定义了Stargate桥的接口。
type Bridge interface {
	GetName() string
	Setup(*App, <-chan []byte) error
	Run(context.Context) error
	Help()
}

// App 结构体定义了Stargate应用实例。
// nolint
type App struct {
	name    string            // 服务名称
	msg     chan []byte       // modules => bridge 传输数据的channel
	modules map[string]Module // 注册的 Module
	bridges map[string]Bridge // 注册的 Bridge
	module  Module            // 被实例化的Module
	bridge  Bridge            // 被实例化的Bridge
	Logger  logger.Logger
	Config  *config.Config
}

type OptionFunc func(*App)

// WithLogger 是一个函数选项，用于设置App实例的Logger属性。
// 它接受一个logger.Logger类型的参数log，并返回一个OptionFunc类型的函数。
// 当这个返回的函数被调用时，它会将其参数log赋值给传入的App实例的Logger属性。
func WithLogger(log logger.Logger) OptionFunc {
	return func(a *App) {
		a.Logger = log
	}
}

// NewApp 函数用于创建一个新的App实例
//
// 参数：
//
//	name string - App的名称
//	ops ...OptionFunc - 可选的OptionFunc函数列表，用于配置App实例
//
// 返回值：
//
//	*App - 指向新创建的App实例的指针
func NewApp(name string, ops ...OptionFunc) *App {
	app := &App{
		Logger: logger.DefaultLogger(),
		name:   name,
		// config: nil,
		// msg:    nil,
		modules: make(map[string]Module, 2),
		bridges: make(map[string]Bridge, 2),
		// module:  nil,
		// bridge:  nil,
	}
	for _, op := range ops {
		op(app)
	}
	return app
}

// Help 是App结构体中的一个方法，用于输出帮助信息
//
// 目前该方法仅输出字符串"help"，作为示例。
// TODO: 研究一下go的模板怎么写，以便未来能够输出更详细的帮助信息
func (app *App) Help() {
	fmt.Println("help")
}

// GetModules 返回应用程序中所有模块的名称列表
//
// 参数：
//   - 无
//
// 返回值：
//   - modules []string：包含所有模块名称的字符串切片
func (app *App) GetModules() (modules []string) {
	for name := range app.modules {
		modules = append(modules, name)
	}
	return
}

// GetBridges 方法从App结构体中获取所有的bridge名称并返回
// 返回值bridges是一个字符串切片，包含所有bridge的名称
func (app *App) GetBridges() (bridges []string) {
	for name := range app.bridges {
		bridges = append(bridges, name)
	}
	return
}

// AddModule 向App结构体中添加一个模块
//
// 参数：
// m - 要添加的模块，类型为Module
//
// 返回值：
// 无
func (app *App) AddModule(m Module) {
	app.modules[m.GetName()] = m
	app.Logger.Info("module registered", logger.MakeField("module", m.GetName()))
}

// AddBridge 向应用程序中添加一个桥接模块
//
// 参数：
//
//	app: *App - 应用程序实例的指针
//	b: Bridge - 要添加的桥接模块
//
// 返回值：
//
//	无返回值
//
// 功能：
//
//	将传入的桥接模块b添加到应用程序app的桥接模块集合中，并使用桥接模块的名称作为键。
//	在添加后，通过Logger记录一条信息，表示桥接模块已成功注册，并记录桥接模块的名称。
func (app *App) AddBridge(b Bridge) {
	app.bridges[b.GetName()] = b
	app.Logger.Info("bridge module registered", logger.MakeField("bridge", b.GetName()))
}

// Setup 函数用于初始化应用程序
// 参数：
//
//	ctx: 上下文对象
//	module: 接收模块名称
//	bridge: 转移模块名称
//	args: 命令行参数列表
//
// 返回值：
//
//	error: 错误信息，如果初始化成功则返回nil
func (app *App) Setup(ctx context.Context, module, bridge string, args []string) error {
	var (
		ok  bool
		err error
	)
	app.module, ok = app.modules[module]
	if !ok {
		app.Help()
		return fmt.Errorf("没有被注册过的接收模块: %s", module)
	}
	app.bridge, ok = app.bridges[bridge]
	if !ok {
		app.Help()
		return fmt.Errorf("没有被注册过的转移模块: %s", bridge)
	}
	if app.Config, err = config.InitConfig(args); err != nil {
		return err
	}
	if app.Config.IsDebug {
		if log, errL := logger.NewLogger(logger.SetDebug(app.Config.IsDebug)); errL != nil {
			app.Logger.Error(logger.ErrorMethod, "设置log", logger.ErrorField(errL))
		} else {
			app.Logger = log
		}
	}
	app.msg = make(chan []byte, app.Config.ChannelSize)
	if err := app.module.Setup(app, app.msg); err != nil {
		return err
	}
	if err := app.bridge.Setup(app, app.msg); err != nil {
		return err
	}
	return nil
}

// Run 方法启动应用程序的两个主要部分：bridge 和 module
// 在给定的上下文（ctx）中，它会并发地运行这两个部分，并等待它们中的任何一个完成
// 如果bridge或module中的任何一个返回错误，该错误将被返回
// 否则，将返回nil
func (app *App) Run(ctx context.Context) error {
	signal := make(chan error)
	go func() {
		signal <- app.bridge.Run(ctx)
	}()

	go func() {
		signal <- app.module.Run(ctx)
	}()
	return <-signal
}
