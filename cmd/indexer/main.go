package main

import (
	"context"
	"evm_event_indexer/background"
	"evm_event_indexer/internal/config"
	"evm_event_indexer/internal/slog"
	"evm_event_indexer/internal/storage"
	"os"
	"os/signal"
	"syscall"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

func main() {

	initInfra()

	// create background manager
	bgManager := background.NewBGManager()

	// register api server
	bgManager.AddWorker(background.NewAPIServer())

	// register subscription
	bgManager.AddWorker(background.NewSubscription())

	// register reorg consumer
	bgManager.AddWorker(background.NewReorgConsumer())

	// register scanners
	for _, scan := range config.Get().Scanner {

		topics := []common.Hash{}
		for _, topic := range scan.Topics {
			topics = append(topics, common.Hash(crypto.Keccak256([]byte(topic))))
		}

		bgManager.AddWorker(background.NewScanner(scan.Address, [][]common.Hash{topics}, scan.BatchSize))
	}

	// global context
	ctx, cancel := context.WithCancel(context.Background())

	// runs background services
	bgManager.Start(ctx)

	defer bgManager.Stop()

	// wait for signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// notify background services to stop
	cancel()
}

// initialize infrastructure
func initInfra() {
	config.LoadConfig("./config/config.yaml")
	slog.InitSlog()
	storage.InitDB()
}
