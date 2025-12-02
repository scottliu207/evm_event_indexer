package main

import (
	"evm_event_indexer/api"
	"evm_event_indexer/background"
	"evm_event_indexer/internal/config"
	internalSlog "evm_event_indexer/internal/slog"
	internalStorage "evm_event_indexer/internal/storage"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

func main() {

	config.LoadConfig("./config/config.yaml")
	internalSlog.InitSlog()

	internalStorage.InitDB()

	go background.Subscription()
	go background.ReorgConsumer()
	for _, scan := range config.Get().Scanner {

		topics := []common.Hash{}
		for _, topic := range scan.Topics {
			topics = append(topics, common.Hash(crypto.Keccak256([]byte(topic))))
		}

		go background.LogScanner(scan.Address, [][]common.Hash{topics}, scan.BatchSize)
	}
	api.Listen()

}
