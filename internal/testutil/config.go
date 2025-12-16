package testutil

import (
	"evm_event_indexer/internal/config"
	"os"
	"path/filepath"
	"runtime"
)

// only for test file
func SetupTestConfig() {
	root := getProjectRoot()
	// load scanner.json to env first, so viper won't override it
	os.Setenv("SCANNER_PATH", filepath.Join(root, "config", "scanner.json"))
	config.LoadConfig(filepath.Join(root, "config", "config.yaml"))
}

func getProjectRoot() string {
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)
	return filepath.Join(dir, "..", "..")
}
