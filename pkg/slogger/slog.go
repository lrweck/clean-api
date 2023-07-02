package slogger

import (
	"io"
	"os"

	"golang.org/x/exp/slog"
)

var DefaultOptions = &slog.HandlerOptions{
	AddSource: true,
}

func NewJSON() *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, DefaultOptions))
}

func NewText() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, DefaultOptions))
}

func NewJSONWithOptions(w io.Writer, opts *slog.HandlerOptions) *slog.Logger {
	return slog.New(slog.NewJSONHandler(w, opts))
}

func NewTextWithOptions(w io.Writer, opts *slog.HandlerOptions) *slog.Logger {
	return slog.New(slog.NewTextHandler(w, opts))
}
