package config

import (
	"log"

	"github.com/BurntSushi/toml"
)

type Config map[string]interface{}

func Load(path string) (Config, error) {
	cfg := make(Config)

	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		log.Printf("Config not found, using defaults")
		return make(Config), nil
	}

	return cfg, nil
}
