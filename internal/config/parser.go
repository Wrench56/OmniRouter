package config

import (
	"fmt"
	"log/slog"

	"github.com/BurntSushi/toml"
)

func ParseConfig(path string) (*Config, error) {
	var cfg Config
	meta, err := toml.DecodeFile(path, &cfg)
	if err != nil {
		slog.Error(fmt.Sprintf("Could not read configuration file \"%q\"", path))
		return nil, err
	}

	if cfg.Modules.Path == "" {
		slog.Error("Missing required setting: modules.path")
		return nil, fmt.Errorf("missing required setting: modules.path")
	}

	if undec := meta.Undecoded(); len(undec) > 0 {
		for _, k := range undec {
			slog.Warn(fmt.Sprintf("Unrecognized configuration key: %s", k.String()))
		}
	}

	slog.Info(fmt.Sprintf("Configuration loaded from %s (modules.path=%q)", path, cfg.Modules.Path))
	return &cfg, nil
}
