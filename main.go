package main

import (
	"omnirouter/internal/config"
	"omnirouter/internal/logger"
	"omnirouter/internal/modmgr"
	"omnirouter/internal/router"

	"github.com/rs/zerolog"
)

func main() {
	logger.Setup()
	logger.SetLevel(zerolog.DebugLevel)
	logger.Info("OmniRouter started!")
	config.ParseConfig("examples/c/hello_world/config.toml")
	modmgr.SetMirrorDir("./mirrordir/")
	modmgr.LookForChanges("examples/c/hello_world/")
	router.RunServer(":8080")
}
