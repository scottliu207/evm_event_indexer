package main

import (
	"context"
	"evm_event_indexer/background"
	"evm_event_indexer/internal/config"
	"evm_event_indexer/internal/slog"
	"evm_event_indexer/internal/storage"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

func main() {
	fmt.Println("Infrastructure initialization started.")
	initInfra()
	fmt.Println("Infrastructure initialization completed, waiting for 10 seconds...")
	time.Sleep(10 * time.Second)

	fmt.Println("Initializing database...")
	storage.InitDB()
	fmt.Println("Database initialized")

	// create background manager
	bgManager := background.NewBGManager()

	// register api server
	bgManager.AddWorker(background.NewAPIServer())

	// register reorg consumer
	bgManager.AddWorker(background.NewReorgConsumer())

	// register scanners and subscriptions
	for _, scan := range config.Get().Scanners {

		addresses := []string{}
		for _, address := range scan.Addresses {
			addresses = append(addresses, address.Address)
			topics := []common.Hash{}

			for _, topic := range address.Topics {
				topics = append(topics, common.Hash(crypto.Keccak256([]byte(topic))))
			}

			// register scanner, each contract has its own scanner
			bgManager.AddWorker(background.NewScanner(scan.RpcHTTP, address.Address, [][]common.Hash{topics}, scan.BatchSize))
		}

		// register subscription, addresses on the same chain share the same subscription
		bgManager.AddWorker(background.NewSubscription(scan.RpcHTTP, scan.RpcWS, addresses))
	}

	// global context
	ctx, cancel := context.WithCancel(context.Background())

	// runs background services
	fmt.Println("Starting background services...")
	bgManager.Start(ctx)

	fmt.Println("Background services started")
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
}
