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

func WithLogger(log logger.Logger) OptionFunc {
	return func(a *App) {
		a.Logger = log
	}
}

// NewApp 实例化stargate
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

func (app *App) Help() {
	// TODO: 研究一下go的模板怎么写
	fmt.Println("help")
}

func (app *App) GetModules() (modules []string) {
	for name := range app.modules {
		modules = append(modules, name)
	}
	return
}

func (app *App) GetBridges() (bridges []string) {
	for name := range app.bridges {
		bridges = append(bridges, name)
	}
	return
}

// AddModule 对指定模块进行初始化
func (app *App) AddModule(m Module) {
	app.modules[m.GetName()] = m
	app.Logger.Info("module registered", logger.MakeField("module", m.GetName()))
}

// AddModule 对指定模块进行初始化
func (app *App) AddBridge(b Bridge) {
	app.bridges[b.GetName()] = b
	app.Logger.Info("bridge module registered", logger.MakeField("bridge", b.GetName()))
}

// Setup 实例化
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

// Run 执行接收
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
