package config_test

import (
	"evm_event_indexer/internal/config"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	config.LoadConfig("../../config/config.yaml")
	t.Log(config.Get())
}
