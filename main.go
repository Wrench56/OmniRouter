package main

import (
	"context"
	"omnirouter/internal/config"
	"omnirouter/internal/logger"
	"omnirouter/internal/modmgr"
	"omnirouter/internal/router"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	logger.Setup()
	logger.SetLevel(zerolog.DebugLevel)
	logger.Info("OmniRouter started!")
	conf, err := config.ParseConfig("examples/c/hello_world/config.toml")
	if err != nil {
		println("Missing required values in config, please see the log for further details")
		return
	}
	modmgr.InitMUID64Map()
	modmgr.SetMirrorDir(conf.Modules.Mirrorlib)
	modmgr.LookForChanges(ctx, "examples/c/hello_world/")
	router.RunServer(ctx, ":8080")

	<-ctx.Done()
}
