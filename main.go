package main

import (
	"log/slog"
	logger "omnirouter/internal"
	"omnirouter/internal/config"
)

func main() {
	logger.SetupLogger()
	slog.Info("Hello World!")
	config.ParseConfig("examples/config.toml")
}
