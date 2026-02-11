package logger

import "log/slog"

func Any(key string, value any) slog.Attr {
	return slog.Any(key, value)
}

func String(key, value string) slog.Attr {
	return slog.String(key, value)
}

func Int(key string, value int) slog.Attr {
	return slog.Int(key, value)
}

func Bool(key string, value bool) slog.Attr {
	return slog.Bool(key, value)
}
