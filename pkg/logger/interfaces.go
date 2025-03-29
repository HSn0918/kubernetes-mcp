package logger

// Logger 定义了日志记录接口
type Logger interface {
	// Debug 记录调试级别日志
	Debug(msg string, keysAndValues ...interface{})

	// Info 记录信息级别日志
	Info(msg string, keysAndValues ...interface{})

	// Warn 记录警告级别日志
	Warn(msg string, keysAndValues ...interface{})

	// Error 记录错误级别日志
	Error(msg string, keysAndValues ...interface{})

	// With 返回带有额外字段的新日志记录器
	With(keysAndValues ...interface{}) Logger

	// Sync 刷新所有缓冲的日志
	Sync() error
}
