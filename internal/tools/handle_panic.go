package tools

import (
	"evm_event_indexer/utils"
	"log/slog"
)

func Recovery(i any) {
	err := recover()
	if err != nil {
		slog.Error("recover 失敗", slog.Any("fn", utils.GetFuncName(i)), slog.Any("error", err))
	}
}
