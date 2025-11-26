package config_test

import (
	"evm_event_indexer/internal/config"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	config.LoadConfig()
	t.Log(config.Get())
}
