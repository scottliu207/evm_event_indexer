package erc20_test

import (
	"context"
	internalCnf "evm_event_indexer/internal/config"
	"evm_event_indexer/internal/erc20"
	"evm_event_indexer/internal/eth"
	"log"
	"math/big"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

var ctx = context.TODO()

func TestMain(m *testing.M) {
	internalCnf.LoadConfig()
	os.Exit(m.Run())
}

func Test_Deploy(t *testing.T) {

	priv := internalCnf.Get().PrivateKey
	if priv == "" {
		log.Fatalf("missing PRIVATE_KEY env (hex, no passphrase)")
	}

	client, err := eth.NewClient(ctx, internalCnf.Get().EthRpcHTTP)
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

	addr := internalCnf.Get().ContractAddress
	priv := internalCnf.Get().PrivateKey
	if priv == "" {
		log.Fatalf("missing PRIVATE_KEY env (hex, no passphrase)")
	}

	client, err := eth.NewClient(ctx, internalCnf.Get().EthRpcHTTP)
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

	tx, err := erc20Service.Transfer(ctx, common.HexToAddress(addr), big.NewInt(1))
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
