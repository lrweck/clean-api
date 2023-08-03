package slogger

import (
	"context"

	"golang.org/x/exp/slog"
)

type sloggerKey string

var LoggerKey sloggerKey = "logger"

func FromContext(ctx context.Context) *slog.Logger {
	s, ok := ctx.Value(LoggerKey).(*slog.Logger)
	if !ok {
		return slog.Default()
	}
	return s
}
