package main

import (
	"log/slog"
	logger "omnirouter/internal"
)

func main() {
	logger.SetupLogger()
	slog.Info("Hello World!")
}
