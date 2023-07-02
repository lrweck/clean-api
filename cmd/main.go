package main

import (
	"golang.org/x/exp/slog"

	"github.com/lrweck/clean-api/internal"
)

func main() {

	app := internal.NewApplication()

	go func() {
		if err := app.Start(5010); err != nil {
			app.Common.Logger.Error("webserver closed", slog.String("error", err.Error()))
		}
	}()

	if err := app.WaitSignal(); err != nil {
		app.Common.Logger.Error("failed to wait for signal", slog.String("error", err.Error()))
	}
}
