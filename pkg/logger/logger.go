package logger

import (
	"os"

	"transfer/pkg/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var Log *zap.Logger
var sugar *zap.SugaredLogger

func init() {
	// 创建默认的开发环境配置
	defaultConfig := zap.NewDevelopmentConfig()
	defaultConfig.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	defaultConfig.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000")

	// 创建默认logger
	defaultLogger, _ := defaultConfig.Build(zap.AddCallerSkip(1))

	// 初始化全局变量
	Log = defaultLogger
	sugar = defaultLogger.Sugar()
}

// Init 初始化日志
func Init() {
	cfg := config.GlobalConfig.Logger

	// 设置日志级别
	atomicLevel := zap.NewAtomicLevel()
	switch cfg.Level {
	case "debug":
		atomicLevel.SetLevel(zapcore.DebugLevel)
	case "info":
		atomicLevel.SetLevel(zapcore.InfoLevel)
	case "warn":
		atomicLevel.SetLevel(zapcore.WarnLevel)
	case "error":
		atomicLevel.SetLevel(zapcore.ErrorLevel)
	default:
		atomicLevel.SetLevel(zapcore.InfoLevel)
	}

	// 设置编码器配置
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02T15:04:05.000+0800")
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	var cores []zapcore.Core

	// 控制台输出
	if cfg.Output == "console" || cfg.Output == "both" {
		consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)
		consoleCore := zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), atomicLevel)
		cores = append(cores, consoleCore)
	}

	// 文件输出
	if cfg.Output == "file" || cfg.Output == "both" {
		// 确保日志目录存在
		if err := os.MkdirAll("logs", 0755); err != nil {
			panic(err)
		}

		fileWriter := &lumberjack.Logger{
			Filename:   cfg.File.Path,
			MaxSize:    cfg.File.MaxSize,
			MaxAge:     cfg.File.MaxAge,
			MaxBackups: cfg.File.MaxBackups,
			Compress:   cfg.File.Compress,
		}

		fileEncoder := zapcore.NewJSONEncoder(encoderConfig)
		fileCore := zapcore.NewCore(fileEncoder, zapcore.AddSync(fileWriter), atomicLevel)
		cores = append(cores, fileCore)
	}

	// 创建logger
	core := zapcore.NewTee(cores...)
	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))

	// 更新全局变量
	Log = logger
	sugar = logger.Sugar()
}

// Debug 输出debug级别日志
func Debug(msg string, fields ...zap.Field) {
	Log.Debug(msg, fields...)
}

// Debugf 输出debug级别日志（格式化）
func Debugf(template string, args ...interface{}) {
	sugar.Debugf(template, args...)
}

// Info 输出info级别日志
func Info(msg string, fields ...zap.Field) {
	Log.Info(msg, fields...)
}

// Infof 输出info级别日志（格式化）
func Infof(template string, args ...interface{}) {
	sugar.Infof(template, args...)
}

// Warn 输出warn级别日志
func Warn(msg string, fields ...zap.Field) {
	Log.Warn(msg, fields...)
}

// Warnf 输出warn级别日志（格式化）
func Warnf(template string, args ...interface{}) {
	sugar.Warnf(template, args...)
}

// Error 输出error级别日志
func Error(msg string, fields ...zap.Field) {
	Log.Error(msg, fields...)
}

// Errorf 输出error级别日志（格式化）
func Errorf(template string, args ...interface{}) {
	sugar.Errorf(template, args...)
}

// Fatal 输出fatal级别日志并退出程序
func Fatal(msg string, fields ...zap.Field) {
	Log.Fatal(msg, fields...)
}

// Fatalf 输出fatal级别日志并退出程序（格式化）
func Fatalf(template string, args ...interface{}) {
	sugar.Fatalf(template, args...)
}

// Sync 同步日志缓冲区
func Sync() error {
	return Log.Sync()
}
