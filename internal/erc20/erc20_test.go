package erc20_test

import (
	"context"
	internalCnf "evm_event_indexer/internal/config"
	"evm_event_indexer/internal/erc20"
	"evm_event_indexer/internal/eth"
	"evm_event_indexer/internal/testutil"
	"log"
	"math/big"
	"os"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
)

var ctx = context.TODO()

func TestMain(m *testing.M) {
	testutil.SetupTestConfig()
	os.Exit(m.Run())
}

func Test_Deploy(t *testing.T) {

	priv := "0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"

	client, err := eth.NewClient(ctx, internalCnf.Get().Scanners[0].RpcHTTP)
	if err != nil {
		t.Fatal(err)
	}

	defer client.Close()

	erc20Service := erc20.NewERC20Service(client, priv)

	res, err := erc20Service.Deploy(new(big.Int).Mul(big.NewInt(100), big.NewInt(1_000_000_000_000_000_000)))
	if err != nil {
		t.Fatalf("deploy failed: %v", err)
	}

	log.Printf("contract address: %s", res.Address)

	pk, err := crypto.HexToECDSA(strings.TrimPrefix(priv, "0x"))
	if err != nil {
		t.Fatalf("hex to ecdsa failed: %v", err)
	}
	assert.NoError(t, err)
	owner := crypto.PubkeyToAddress(pk.PublicKey)

	balance, err := erc20Service.GetBalance(ctx, res.Address, owner)
	if err != nil {
		t.Fatalf("get balance failed: %v", err)
	}
	log.Printf("balance: %s", balance)

}

func Test_GetBalance(t *testing.T) {

	addr := internalCnf.Get().Scanners[0].Addresses[0].Address
	priv := "0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"

	client, err := eth.NewClient(ctx, internalCnf.Get().Scanners[0].RpcHTTP)
	if err != nil {
		t.Fatal(err)
	}

	defer client.Close()

	erc20Service := erc20.NewERC20Service(client, priv)

	pk, err := crypto.HexToECDSA(strings.TrimPrefix(priv, "0x"))
	if err != nil {
		t.Fatalf("hex to ecdsa failed: %v", err)
	}

	owner := crypto.PubkeyToAddress(pk.PublicKey)
	balance, err := erc20Service.GetBalance(ctx, common.HexToAddress(addr), owner)
	if err != nil {
		t.Fatalf("get balance failed: %v", err)
	}
	log.Printf("balance: %s", balance)

}

func Test_Transfer(t *testing.T) {

	addr := internalCnf.Get().Scanners[0].Addresses[0].Address
	priv := "0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"

	client, err := eth.NewClient(ctx, internalCnf.Get().Scanners[0].RpcHTTP)
	if err != nil {
		t.Fatal(err)
	}

	defer client.Close()

	erc20Service := erc20.NewERC20Service(client, priv)

	pk, err := crypto.HexToECDSA(strings.TrimPrefix(priv, "0x"))
	if err != nil {
		t.Fatalf("hex to ecdsa failed: %v", err)
	}
	owner := crypto.PubkeyToAddress(pk.PublicKey)
	balance, err := erc20Service.GetBalance(ctx, common.HexToAddress(addr), owner)
	if err != nil {
		t.Fatalf("get balance failed: %v", err)
	}
	log.Printf("balance: %s", balance)

	tx, err := erc20Service.Transfer(ctx, common.HexToAddress(addr), big.NewInt(1))
	if err != nil {
		t.Fatalf("transfer failed: %v", err)
	}
	log.Printf("tx: %s", tx.Hash())

	balance, err = erc20Service.GetBalance(ctx, common.HexToAddress(addr), owner)
	if err != nil {
		t.Fatalf("get balance failed: %v", err)
	}
	log.Printf("after transfer: %s", balance)

}
