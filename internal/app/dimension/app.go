package dimension

import (
	"context"
	"fmt"

	config "github.com/chengfeiZhou/Wormhole/configs/dimension"
	"github.com/chengfeiZhou/Wormhole/internal/pkg/structs"
	"github.com/chengfeiZhou/Wormhole/pkg/logger"
)

type Module interface {
	GetName() string
	Setup(*App, <-chan []byte) error
	Run(context.Context) error
	Transform() structs.TransMessage
	Help()
}

type Bridge interface {
	GetName() string
	Setup(*App, chan<- []byte) error
	Run(context.Context) error
	Help()
}

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

// WithLogger 是一个函数选项，用于设置App实例的Logger字段。
// 它接收一个logger.Logger类型的参数log，并返回一个OptionFunc类型的函数。
// 这个返回的OptionFunc函数在调用时会将传入的log设置到App实例的Logger字段中。
func WithLogger(log logger.Logger) OptionFunc {
	return func(a *App) {
		a.Logger = log
	}
}

// NewApp 创建一个新的App实例
//
// 参数：
//
//	name: string - 应用名称
//	ops:  ...OptionFunc - 可选参数列表，用于配置App实例
//
// 返回值：
//
//	*App - 返回一个指向新创建的App实例的指针
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

// Help 方法用于输出帮助信息
func (app *App) Help() {
	// TODO: 研究一下go的模板怎么写
	fmt.Println("help")
}

// GetModules 返回App结构体中modules字段中所有模块的名称切片
// 参数：
//
//	app *App - App结构体指针
//
// 返回值：
//
//	modules []string - 包含所有模块名称的字符串切片
func (app *App) GetModules() (modules []string) {
	for name := range app.modules {
		modules = append(modules, name)
	}
	return
}

// GetBridges 返回App结构体中所有桥接器的名称列表
//
// 参数:
//   - app: App结构体指针，表示要获取桥接器名称列表的应用实例
//
// 返回值:
//   - bridges: []string类型，表示桥接器名称的列表
func (app *App) GetBridges() (bridges []string) {
	for name := range app.bridges {
		bridges = append(bridges, name)
	}
	return
}

// AddModule 方法用于向App实例中添加一个模块
//
// 参数：
//
//	m Module：要添加的模块，需要实现Module接口
//
// 返回值：
//
//	无
func (app *App) AddModule(m Module) {
	app.modules[m.GetName()] = m
	app.Logger.Info("module registered", logger.MakeField("module", m.GetName()))
}

// AddBridge 向应用程序中添加一个桥接模块
//
// 参数：
//
//	app *App - 应用程序实例指针
//	b Bridge - 要添加的桥接模块
//
// 返回值：
//
//	无
//
// 功能：
//  1. 将传入的桥接模块 b 存储到应用程序实例 app 的 bridges 字段中，键为桥接模块的名称（通过 b.GetName() 获取）
//  2. 使用 Logger 输出一条日志信息，记录桥接模块已经注册，包含桥接模块的名称作为日志字段
func (app *App) AddBridge(b Bridge) {
	app.bridges[b.GetName()] = b
	app.Logger.Info("bridge module registered", logger.MakeField("bridge", b.GetName()))
}

// Setup 用于初始化App对象，并设置发送模块、转移模块以及配置信息
//
// 参数：
//
//	ctx context.Context：上下文对象
//	module string：发送模块名称
//	bridge string：转移模块名称
//	args []string：配置参数列表
//
// 返回值：
//
//	error：如果设置过程中发生错误，则返回非零的错误码；否则返回nil
func (app *App) Setup(ctx context.Context, module, bridge string, args []string) error {
	var (
		ok  bool
		err error
	)
	app.module, ok = app.modules[module]
	if !ok {
		app.Help()
		return fmt.Errorf("没有被注册过的发送模块: %s", module)
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

// Run 启动App的运行
// ctx 用于控制应用程序的生命周期，可用来在App运行期间传递信号或取消操作
// 返回值error表示启动过程中是否出现错误
func (app *App) Run(ctx context.Context) error {
	signal := make(chan error)
	go func() {
		signal <- app.module.Run(ctx)
	}()
	go func() {
		signal <- app.bridge.Run(ctx)
	}()
	return <-signal
}
