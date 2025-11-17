package main

import (
	"log"

	"github.com/joho/godotenv"
)

func main() {

	// Load env
	if err := godotenv.Load(); err != nil {
		log.Fatalf("load env: %v", err)
	}

}
