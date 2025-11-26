package main

import (
	"evm_event_indexer/api"
	"evm_event_indexer/background"
	internalCnf "evm_event_indexer/internal/config"
	internalSlog "evm_event_indexer/internal/slog"
	internalStorage "evm_event_indexer/internal/storage"
)

func main() {

	internalCnf.LoadConfig()
	internalSlog.InitSlog()

	internalStorage.InitDB()

	go background.Subscription()
	go background.ReorgConsumer()
	go background.LogScanner(internalCnf.Get().ContractAddress)
	api.Listen()

}
