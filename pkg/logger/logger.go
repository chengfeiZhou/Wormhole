package logger

import (
	"fmt"

	"github.com/chengfeiZhou/Wormhole/pkg/times"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	otputPaths = []string{"stderr"}
)

// Field 可以指定关键信息
type Field struct {
	Value interface{}
	Key   string
}

// MakeField 函数接收两个参数，一个字符串类型的key和一个空接口类型的value，
// 返回一个Field类型的结构体，该结构体包含Key和Value两个字段，分别对应传入的key和value。
func MakeField(k string, v interface{}) Field {
	return Field{Key: k, Value: v}
}

// ErrorField 函数接收一个 error 类型的 err 参数，返回一个 Field 结构体
// 如果 err 不为 nil，则返回的 Field 结构体中的 Key 字段为 "error"，Value 字段为 err
func ErrorField(err error) (f Field) {
	return Field{Key: "error", Value: err}
}

// Level 声明日志级别
type Level zapcore.Level

const (
	DebugLevel Level = iota - 1
	InfoLevel
	WarnLevel
	ErrorLevel
	dPanicLevel // nolint 占位, 不给外部使用
	panicLevel  // nolint 占位, 不给外部使用
	FatalLevel
)

// Logger 定义logger接口
type Logger interface {
	// SetLevel(Level)
	// GetLevel() string
	Log() *zap.Logger
	Debugf(string, ...interface{})
	Info(string, ...Field)
	Infof(string, ...interface{})
	Warn(AppError, string, ...Field)
	Warnf(AppError, string, ...interface{})
	Error(AppError, string, ...Field)
	Errorf(AppError, string, ...interface{})
	Fatal(AppError, string, ...Field)
	Fatalf(AppError, string, ...interface{})
	Println(...interface{})
}

// AppLogger 实现一个logger
type AppLogger struct {
	slg   *zap.SugaredLogger
	lg    *zap.Logger
	level zap.AtomicLevel
}

// //  RegisterAPI 是一个简单的JSON端点，可以报告或更改当前日志级别
// // https://pkg.go.dev/go.uber.org/zap#AtomicLevel.ServeHTTP
// // method: GET. PUT
// func (applg *AppLogger) RegisterAPI(mux *http.ServeMux) {
// 	mux.HandleFunc("/logger/level", applg.level.ServeHTTP)
// }

// SetLevel 方法用于设置 AppLogger 的日志级别
//
// 参数：
//
//	l Level - 要设置的日志级别
//
// 返回值：
//
//	无返回值
func (applg *AppLogger) SetLevel(l Level) {
	applg.level.SetLevel(zapcore.Level(l))
}

// GetLevel 返回AppLogger的日志级别，返回值为字符串类型
func (applg *AppLogger) GetLevel() string {
	return applg.level.Level().CapitalString()
}

// Log 方法返回AppLogger结构体中的zap.Logger实例指针
func (applg *AppLogger) Log() *zap.Logger {
	return applg.lg
}

// Debug logs a message at DebugLevel with SugaredLogger
func (applg *AppLogger) Debugf(temp string, args ...interface{}) {
	applg.slg.Debugf(temp, args...)
	go applg.slg.Sync() // nolint
}

// Info logs a message at InfoLevel with Logger
func (applg *AppLogger) Info(msg string, fields ...Field) {
	fieldsArr := make([]zap.Field, 0, len(fields))
	for _, field := range fields {
		fieldsArr = append(fieldsArr, zap.Any(field.Key, field.Value))
	}
	applg.lg.Info(msg, fieldsArr...)
	go applg.lg.Sync() // nolint
}

// Infof logs a message at InfoLevel with SugaredLogger
func (applg *AppLogger) Println(args ...interface{}) {
	applg.slg.Info(args...)
	go applg.slg.Sync() // nolint
}

func (applg *AppLogger) IsDebugLevel() bool {
	return applg.level.Level() == zapcore.Level(DebugLevel)
}

// Infof logs a message at InfoLevel with SugaredLogger
func (applg *AppLogger) Infof(temp string, args ...interface{}) {
	applg.slg.Infof(temp, args...)
	go applg.slg.Sync() // nolint
}

// Warn logs a message at WarnLevel as AppError with Logger
func (applg *AppLogger) Warn(err AppError, msg string, fields ...Field) {
	fieldsArr := make([]zap.Field, 0, len(fields))
	for _, field := range fields {
		fieldsArr = append(fieldsArr, zap.Any(field.Key, field.Value))
	}
	applg.lg.Warn(fmt.Sprintf("%d: %s", err.Code(), msg), fieldsArr...)
	go applg.lg.Sync() // nolint
}

// Warnf logs a message at WarnLevel with SugaredLogger
func (applg *AppLogger) Warnf(err AppError, temp string, args ...interface{}) {
	applg.slg.Warnf(fmt.Sprintf("%d: %s", err.Code(), temp), args...)
	go applg.slg.Sync() // nolint
}

// Error logs a message at ErrorLevel with Logger
func (applg *AppLogger) Error(err AppError, msg string, fields ...Field) {
	fieldsArr := make([]zap.Field, 0, len(fields))
	for _, field := range fields {
		fieldsArr = append(fieldsArr, zap.Any(field.Key, field.Value))
	}
	applg.lg.Error(fmt.Sprintf("%d: %s", err.Code(), msg), fieldsArr...)
	go applg.lg.Sync() // nolint
}

// Errorf logs a message at ErrorLevel with SugaredLogger
func (applg *AppLogger) Errorf(err AppError, temp string, args ...interface{}) {
	applg.slg.Errorf(fmt.Sprintf("%d: %s", err.Code(), temp), args...)
	go applg.slg.Sync() // nolint
}

// Fatal logs a message and exit at FatalLevel with Logger
// The logger then calls os.Exit(1), even if logging at FatalLevel is disabled
func (applg *AppLogger) Fatal(err AppError, msg string, fields ...Field) {
	fieldsArr := make([]zap.Field, 0, len(fields))
	for _, field := range fields {
		fieldsArr = append(fieldsArr, zap.Any(field.Key, field.Value))
	}
	applg.lg.Fatal(fmt.Sprintf("%d: %s", err.Code(), msg), fieldsArr...)
	go applg.lg.Sync() // nolint
}

// Fatalf logs a message at FatalLevel with SugaredLogger
// The logger then calls os.Exit(1), even if logging at FatalLevel is disabled
func (applg *AppLogger) Fatalf(err AppError, temp string, args ...interface{}) {
	applg.slg.Fatalf(fmt.Sprintf("%d: %s", err.Code(), temp), args...)
	go applg.slg.Sync() // nolint
}

type Option func(conf *zap.Config)

// SetDebug 用于创建logger时,设置是否处于开发环境
// ps: 会将Level设置成debug
func SetDebug(debug bool) Option {
	encoding := "console"
	level := zapcore.InfoLevel
	if debug {
		encoding = "json"
		level = zapcore.DebugLevel
	}
	return func(conf *zap.Config) {
		conf.Development = debug
		conf.Encoding = encoding
		conf.Level = zap.NewAtomicLevelAt(level)
	}
}

// SetLevel 创建Logger实例时, 设置日志等级
func SetLevel(lv Level) Option {
	return func(conf *zap.Config) {
		conf.Level = zap.NewAtomicLevelAt(zapcore.Level(lv))
	}
}

// NopLogger 一个空的log实例
func NopLogger() Logger {
	lg := zap.NewNop()
	return &AppLogger{slg: lg.Sugar(), lg: lg, level: zap.NewAtomicLevelAt(zapcore.Level(DebugLevel))}
}

// DefaultLogger 返回一个基础的默认logger实例
func DefaultLogger() Logger {
	lg := zap.NewExample()
	return &AppLogger{slg: lg.Sugar(), lg: lg, level: zap.NewAtomicLevelAt(zapcore.Level(DebugLevel))}
}

// NewLogger 创建logger实例对象:
// opts: logger包提供的对logger配置参数的配置方法
func NewLogger(opts ...Option) (Logger, error) {
	cfg := zap.Config{
		Level:       zap.NewAtomicLevelAt(zapcore.InfoLevel),
		Development: false,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding: "console",
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "T",
			LevelKey:       "L",
			NameKey:        "N",
			CallerKey:      "C",
			FunctionKey:    zapcore.OmitKey,
			MessageKey:     "M",
			StacktraceKey:  "S",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.CapitalLevelEncoder,                     // 等级显示标识
			EncodeTime:     zapcore.TimeEncoderOfLayout(times.TimeFormatMS), // 时间格式
			EncodeDuration: zapcore.StringDurationEncoder,                   // 调用时间序列化
			EncodeCaller:   zapcore.ShortCallerEncoder,                      // 调用文件路径
		},
		OutputPaths:      otputPaths,
		ErrorOutputPaths: otputPaths,
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	lg, err := cfg.Build(zap.AddCallerSkip(1))
	if err != nil {
		return nil, err
	}
	return &AppLogger{slg: lg.Sugar(), lg: lg, level: cfg.Level}, nil
}
