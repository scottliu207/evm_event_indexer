package config_test

import (
	"evm_event_indexer/internal/config"
	"evm_event_indexer/internal/testutil"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	testutil.SetupTestConfig()
	t.Log(config.Get())
}
