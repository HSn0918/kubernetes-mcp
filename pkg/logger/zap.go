package logger

import (
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// zapLogger 是基于zap的Logger实现
type zapLogger struct {
	logger *zap.SugaredLogger
}

// 确保zapLogger实现了Logger接口
var _ Logger = &zapLogger{}

// 默认日志记录器
var defaultLogger Logger

// Debug 实现接口方法
func (l *zapLogger) Debug(msg string, keysAndValues ...interface{}) {
	l.logger.Debugw(msg, keysAndValues...)
}

// Info 实现接口方法
func (l *zapLogger) Info(msg string, keysAndValues ...interface{}) {
	l.logger.Infow(msg, keysAndValues...)
}

// Warn 实现接口方法
func (l *zapLogger) Warn(msg string, keysAndValues ...interface{}) {
	l.logger.Warnw(msg, keysAndValues...)
}

// Error 实现接口方法
func (l *zapLogger) Error(msg string, keysAndValues ...interface{}) {
	l.logger.Errorw(msg, keysAndValues...)
}

// With 实现接口方法
func (l *zapLogger) With(keysAndValues ...interface{}) Logger {
	return &zapLogger{logger: l.logger.With(keysAndValues...)}
}

// Sync 实现接口方法
func (l *zapLogger) Sync() error {
	return l.logger.Sync()
}

// NewZapLogger 创建新的zap日志记录器
func NewZapLogger(level, format string) Logger {
	var zapLevel zapcore.Level

	switch strings.ToLower(level) {
	case "debug":
		zapLevel = zapcore.DebugLevel
	case "info":
		zapLevel = zapcore.InfoLevel
	case "warn", "warning":
		zapLevel = zapcore.WarnLevel
	case "error":
		zapLevel = zapcore.ErrorLevel
	default:
		zapLevel = zapcore.InfoLevel
	}

	var encoding string
	switch strings.ToLower(format) {
	case "json":
		encoding = "json"
	case "console":
		encoding = "console"
	default:
		encoding = "console"
	}

	config := zap.NewProductionConfig()
	config.Level = zap.NewAtomicLevelAt(zapLevel)
	config.Encoding = encoding
	config.OutputPaths = []string{"stdout"}
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncoderConfig.TimeKey = "time"

	if encoding == "console" {
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	logger, _ := config.Build()
	return &zapLogger{logger: logger.Sugar()}
}

// InitializeDefaultLogger 初始化默认日志记录器
func InitializeDefaultLogger(level, format string) {
	defaultLogger = NewZapLogger(level, format)
}

// GetLogger 获取默认日志记录器
func GetLogger() Logger {
	if defaultLogger == nil {
		InitializeDefaultLogger("info", "console")
	}
	return defaultLogger
}
