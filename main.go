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
	config.ParseConfig("examples/c/hello_world/config.toml")
	modmgr.SetMirrorDir("./mirrordir/")
	modmgr.LookForChanges(ctx, "examples/c/hello_world/")
	router.RunServer(":8080")

	<-ctx.Done()
}
