package slog

import (
	internalCnf "evm_event_indexer/internal/config"
	"log/slog"
	"os"
	"strings"
)

func InitSlog() {
	var slogLevel slog.Level

	// 設定 Log Level, 參數吃 Config 檔中的設定值
	// Note: 因爲這邊要使用到 Config, 所以初始化的步驟需要在 Config 之後
	switch strings.ToLower(internalCnf.Get().LogLevel) {
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
