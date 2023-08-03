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
	l := slog.New(slog.NewJSONHandler(os.Stdout, DefaultOptions))
	slog.SetDefault(l)
	return l
}

func NewText() *slog.Logger {
	l := slog.New(slog.NewTextHandler(os.Stdout, DefaultOptions))
	slog.SetDefault(l)
	return l
}

func NewJSONWithOptions(w io.Writer, opts *slog.HandlerOptions) *slog.Logger {
	l := slog.New(slog.NewJSONHandler(w, opts))
	slog.SetDefault(l)
	return l
}

func NewTextWithOptions(w io.Writer, opts *slog.HandlerOptions) *slog.Logger {
	l := slog.New(slog.NewTextHandler(w, opts))
	slog.SetDefault(l)
	return l
}
