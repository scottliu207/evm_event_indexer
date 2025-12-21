package tools

import (
	"evm_event_indexer/utils"
	"log/slog"
)

func Recovery(i any) {
	err := recover()
	if err != nil {
		slog.Error("recover failed", slog.Any("fn", utils.GetFuncName(i)), slog.Any("error", err))
	}
}
