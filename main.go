package main

import (
	"omnirouter/internal/config"
	"omnirouter/internal/logger"
	"omnirouter/internal/modmgr"

	"github.com/rs/zerolog"
)

func main() {
	logger.Setup()
	logger.SetLevel(zerolog.DebugLevel)
	logger.Info("OmniRouter started!")
	config.ParseConfig("examples/config.toml")
	modmgr.LookForChanges("examples/modules/")
}
