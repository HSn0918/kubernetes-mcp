package logger

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"
)

type slogLogger struct {
	logger *slog.Logger
}

var _ Logger = &slogLogger{}

var defaultLogger Logger

func (l *slogLogger) Debug(msg string, keysAndValues ...interface{}) {
	l.logger.Debug(msg, normalizeArgs(keysAndValues...)...)
}

func (l *slogLogger) Info(msg string, keysAndValues ...interface{}) {
	l.logger.Info(msg, normalizeArgs(keysAndValues...)...)
}

func (l *slogLogger) Warn(msg string, keysAndValues ...interface{}) {
	l.logger.Warn(msg, normalizeArgs(keysAndValues...)...)
}

func (l *slogLogger) Error(msg string, keysAndValues ...interface{}) {
	l.logger.Error(msg, normalizeArgs(keysAndValues...)...)
}

func (l *slogLogger) With(keysAndValues ...interface{}) Logger {
	return &slogLogger{logger: l.logger.With(normalizeArgs(keysAndValues...)...)}
}

func (l *slogLogger) Sync() error {
	return nil
}

func NewSlogLogger(level, format string) Logger {
	handlerOptions := &slog.HandlerOptions{Level: parseSlogLevel(level)}

	var handler slog.Handler
	switch strings.ToLower(format) {
	case "json":
		handler = slog.NewJSONHandler(os.Stdout, handlerOptions)
	default:
		handler = slog.NewTextHandler(os.Stdout, handlerOptions)
	}

	return &slogLogger{logger: slog.New(handler)}
}

func InitializeDefaultLogger(level, format string) {
	defaultLogger = NewSlogLogger(level, format)
}

func GetLogger() Logger {
	if defaultLogger == nil {
		InitializeDefaultLogger("info", "console")
	}
	return defaultLogger
}

func parseSlogLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func normalizeArgs(args ...interface{}) []any {
	if len(args) == 0 {
		return nil
	}

	normalized := make([]any, 0, len(args)+1)
	for i := 0; i < len(args); {
		if attr, ok := args[i].(slog.Attr); ok {
			normalized = append(normalized, attr)
			i++
			continue
		}

		if i == len(args)-1 {
			normalized = append(normalized, toAttr("value", args[i]))
			break
		}

		key, ok := args[i].(string)
		if !ok || strings.TrimSpace(key) == "" {
			normalized = append(normalized, toAttr(fmt.Sprintf("arg_%d", i), args[i]))
			i++
			continue
		}

		normalized = append(normalized, toAttr(key, args[i+1]))
		i += 2
	}

	return normalized
}

func toAttr(key string, value any) slog.Attr {
	switch v := value.(type) {
	case string:
		return slog.String(key, v)
	case bool:
		return slog.Bool(key, v)
	case int:
		return slog.Int(key, v)
	case int8:
		return slog.Int64(key, int64(v))
	case int16:
		return slog.Int64(key, int64(v))
	case int32:
		return slog.Int64(key, int64(v))
	case int64:
		return slog.Int64(key, v)
	case uint:
		return slog.Uint64(key, uint64(v))
	case uint8:
		return slog.Uint64(key, uint64(v))
	case uint16:
		return slog.Uint64(key, uint64(v))
	case uint32:
		return slog.Uint64(key, uint64(v))
	case uint64:
		return slog.Uint64(key, v)
	case float32:
		return slog.Float64(key, float64(v))
	case float64:
		return slog.Float64(key, v)
	case time.Duration:
		return slog.Duration(key, v)
	case time.Time:
		return slog.Time(key, v)
	default:
		return slog.Any(key, value)
	}
}
