package slogerxexp

import (
	slog "golang.org/x/exp/slog"
)

var (
	DefaultLogger *slog.Logger
)

func SetDefaultLogger(l *slog.Logger) {
	DefaultLogger = l
}

func With(args ...any) *slog.Logger {
	if DefaultLogger == nil {
		return slog.With(args...)
	}
	return DefaultLogger.With(args...)
}
