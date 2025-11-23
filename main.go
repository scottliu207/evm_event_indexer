package main

import (
	"evm_event_indexer/background"
	"os"

	"github.com/joho/godotenv"
)

// TODO: config
const CONTRACT_ADDRESS = "0x5FbDB2315678afecb367f032d93F642f64180aa3"

func main() {

	// Load env
	if err := godotenv.Load(); err != nil {
		panic(err)
	}

	go background.Subscription()
	go background.ReorgConsumer()

	background.LogScanner(os.Getenv("CONTRACT_ADDRESS"))
}
