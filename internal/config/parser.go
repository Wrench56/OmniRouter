package config

import (
	"fmt"
	"omnirouter/internal/logger"

	"github.com/BurntSushi/toml"
)

func ParseConfig(path string) (*Config, error) {
	var cfg Config
	meta, err := toml.DecodeFile(path, &cfg)
	if err != nil {
		logger.Error(fmt.Sprintf("Could not read configuration file \"%q\"", path))
		return nil, err
	}

	if cfg.Modules.Path == "" {
		logger.Error("Missing required setting: modules.path")
		return nil, fmt.Errorf("missing required setting: modules.path")
	}

	if undec := meta.Undecoded(); len(undec) > 0 {
		for _, k := range undec {
			logger.Warn(fmt.Sprintf("Unrecognized configuration key: %s", k.String()))
		}
	}

	logger.Info(fmt.Sprintf("Configuration loaded from %s (modules.path=%q)", path, cfg.Modules.Path))
	return &cfg, nil
}
