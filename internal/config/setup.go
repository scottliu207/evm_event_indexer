package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

func LoadConfig(path string) {
	if path == "" {
		path = "./config/config.yaml"
	}

	data, err := os.ReadFile(path)
	if err != nil {
		panic(fmt.Errorf("read config file failed: %w", err))
	}

	if err := yaml.Unmarshal(data, &config); err != nil {
		panic(fmt.Errorf("yaml unmarshal failed: %w", err))
	}

	if err := config.Validate(); err != nil {
		panic(fmt.Errorf("config validation failed: %w", err))
	}
}

func Get() *Config {
	return config
}
