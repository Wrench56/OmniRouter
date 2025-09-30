package main

import (
	logger "omnirouter/internal"
	"omnirouter/internal/config"
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
