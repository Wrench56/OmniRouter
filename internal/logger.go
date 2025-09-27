package logger

import (
	"os"

	"log/slog"
	"github.com/lmittmann/tint"
	"github.com/mattn/go-isatty"
)

func SetupLogger() {
	logger := slog.New(tint.NewHandler(os.Stderr, &tint.Options{
		NoColor: !isatty.IsTerminal(os.Stderr.Fd()),
	}))
	slog.SetDefault(logger)
}
