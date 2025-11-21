package erc20_test

import (
	"context"
	"evm_event_indexer/internal/erc20"
	"evm_event_indexer/internal/eth"
	"log"
	"math/big"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/joho/godotenv"
)

var ctx = context.TODO()

func Test_Deploy(t *testing.T) {
	if err := godotenv.Load("../../.env"); err != nil {
		panic(err)
	}

	priv := os.Getenv("PRIVATE_KEY")
	if priv == "" {
		log.Fatalf("missing PRIVATE_KEY env (hex, no passphrase)")
	}

	client, err := eth.NewClient(ctx, os.Getenv("ETH_RPC"))
	if err != nil {
		t.Fatal(err)
	}

	defer client.Close()

	erc20Service := erc20.NewERC20Service(client, priv)

	res, err := erc20Service.Deploy()
	if err != nil {
		t.Fatalf("deploy failed: %v", err)
	}

	log.Printf("contract address: %s", res.Address)

	balance, err := erc20Service.GetBalance(ctx, res.Address)
	if err != nil {
		t.Fatalf("get balance failed: %v", err)
	}
	log.Printf("balance: %s", balance)

}

func Test_Transfer(t *testing.T) {
	if err := godotenv.Load("../../.env"); err != nil {
		panic(err)
	}

	addr := "0x5FbDB2315678afecb367f032d93F642f64180aa3"

	priv := os.Getenv("PRIVATE_KEY")
	if priv == "" {
		log.Fatalf("missing PRIVATE_KEY env (hex, no passphrase)")
	}

	client, err := eth.NewClient(ctx, os.Getenv("ETH_RPC"))
	if err != nil {
		t.Fatal(err)
	}

	defer client.Close()

	erc20Service := erc20.NewERC20Service(client, priv)

	balance, err := erc20Service.GetBalance(ctx, common.HexToAddress(addr))
	if err != nil {
		t.Fatalf("get balance failed: %v", err)
	}
	log.Printf("balance: %s", balance)

	tx, err := erc20Service.Transfer(ctx, common.HexToAddress(addr), big.NewInt(3))
	if err != nil {
		t.Fatalf("transfer failed: %v", err)
	}
	log.Printf("tx: %s", tx.Hash())

	balance, err = erc20Service.GetBalance(ctx, common.HexToAddress(addr))
	if err != nil {
		t.Fatalf("get balance failed: %v", err)
	}
	log.Printf("after transfer: %s", balance)

}
