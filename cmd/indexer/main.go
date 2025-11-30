package main

import (
	"evm_event_indexer/api"
	internalCnf "evm_event_indexer/internal/config"
	internalSlog "evm_event_indexer/internal/slog"
	internalStorage "evm_event_indexer/internal/storage"
)

func main() {

	internalCnf.LoadConfig("./config/config.yaml")
	internalSlog.InitSlog()

	internalStorage.InitDB()

	// go background.Subscription()
	// go background.ReorgConsumer()
	// for _, addr := range internalCnf.Get().ContractsAddress {
	// 	go background.LogScanner(addr)
	// }
	api.Listen()

}
