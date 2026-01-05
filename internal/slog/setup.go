package slog

import (
	"evm_event_indexer/internal/config"
	"log/slog"
	"os"
	"strings"
)

// InitSlog initializes the slog logger, must be called after config is loaded
func InitSlog() {
	var slogLevel slog.Level

	switch strings.ToLower(config.Get().LogLevel) {
	case "panic", "fatal", "error":
		slogLevel = slog.LevelError
	case "warn":
		slogLevel = slog.LevelWarn
	case "info":
		slogLevel = slog.LevelInfo
	case "debug", "trace":
		slogLevel = slog.LevelDebug
	default:
		slogLevel = slog.LevelError
	}

	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slogLevel})))
}
