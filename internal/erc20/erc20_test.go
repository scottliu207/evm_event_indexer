package erc20_test

import (
	"context"
	"evm_event_indexer/internal/config"
	"evm_event_indexer/internal/erc20"
	"evm_event_indexer/internal/eth"
	"evm_event_indexer/internal/testutil"
	"log"
	"math/big"
	"os"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

var ctx = context.TODO()

func TestMain(m *testing.M) {
	testutil.SetupTestConfig()
	os.Exit(m.Run())
}

func Test_Deploy(t *testing.T) {

	ownerPK := "0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"

	client, err := eth.NewClient(ctx, config.Get().Scanners[0].RpcHTTP)
	if err != nil {
		t.Fatal(err)
	}

	defer client.Close()

	service := erc20.NewERC20Service(client, ownerPK)

	res, err := service.Deploy(new(big.Int).Mul(big.NewInt(100), big.NewInt(1_000_000_000_000_000_000)))
	if err != nil {
		t.Fatalf("deploy failed: %v", err)
	}

	_, err = bind.WaitMined(ctx, client.Client, res.Tx)
	if err != nil {
		t.Fatalf("wait deploy mined failed: %v", err)
	}

	log.Printf("contract address: %s", res.Address)

	balance, err := service.GetBalance(ctx, res.Address)
	if err != nil {
		t.Fatalf("get balance failed: %v", err)
	}
	log.Printf("balance: %s", balance)

}

func Test_GetBalance(t *testing.T) {

	ownerPK := "0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"
	addr := config.Get().Scanners[0].Addresses[0].Address

	client, err := eth.NewClient(ctx, config.Get().Scanners[0].RpcHTTP)
	if err != nil {
		t.Fatal(err)
	}

	defer client.Close()

	service := erc20.NewERC20Service(client, ownerPK)

	balance, err := service.GetBalance(ctx, common.HexToAddress(addr))
	if err != nil {
		t.Fatalf("get balance failed: %v", err)
	}
	log.Printf("balance: %s", balance)

}

func Test_Transfer(t *testing.T) {

	priv := "0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"
	to := common.HexToAddress("0x70997970C51812dc3A010C7d01b50e0d17dc79C8")

	client, err := eth.NewClient(ctx, config.Get().Scanners[0].RpcHTTP)
	if err != nil {
		t.Fatal(err)
	}

	defer client.Close()

	erc20Service := erc20.NewERC20Service(client, priv)

	deployRes, err := erc20Service.Deploy(big.NewInt(1000))
	if err != nil {
		t.Fatalf("deploy failed: %v", err)
	}

	// wait deploy mined
	_, err = bind.WaitMined(ctx, client.Client, deployRes.Tx)
	if err != nil {
		t.Fatalf("wait deploy mined failed: %v", err)
	}

	balance, err := erc20Service.GetBalance(ctx, deployRes.Address)
	if err != nil {
		t.Fatalf("get balance failed: %v", err)
	}
	log.Printf("balance: %s", balance)

	tx, err := erc20Service.Transfer(ctx, deployRes.Address, to, big.NewInt(10))
	if err != nil {
		t.Fatalf("transfer failed: %v", err)
	}
	log.Printf("tx: %s", tx.Hash())

	// wait approve mined
	_, err = bind.WaitMined(ctx, client.Client, tx)
	if err != nil {
		t.Fatalf("wait approve mined failed: %v", err)
	}

	balance, err = erc20Service.GetBalance(ctx, deployRes.Address)
	if err != nil {
		t.Fatalf("get balance failed: %v", err)
	}
	log.Printf("after transfer: %s", balance)

}

func Test_DeplayAndTransfer(t *testing.T) {
	ownerPK := "0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"

	client, err := eth.NewClient(ctx, config.Get().Scanners[0].RpcHTTP)
	if err != nil {
		t.Fatal(err)
	}

	defer client.Close()

	service := erc20.NewERC20Service(client, ownerPK)

	deployRes, err := service.Deploy(big.NewInt(1000))
	if err != nil {
		t.Fatalf("deploy failed: %v", err)
	}

	// wait deploy mined
	_, err = bind.WaitMined(ctx, client.Client, deployRes.Tx)
	if err != nil {
		t.Fatalf("wait deploy mined failed: %v", err)
	}

	balance, err := service.GetBalance(ctx, deployRes.Address)
	if err != nil {
		t.Fatalf("get balance failed: %v", err)
	}

	// owner should has 1000 uints in balance
	assert.Equal(t, decimal.NewFromInt(1000).String(), balance.String())
	log.Printf("owner balance: %s", balance)

	receiverAddress := common.HexToAddress("0x70997970C51812dc3A010C7d01b50e0d17dc79C8")

	// transfer 10 units into another address
	tx, err := service.Transfer(ctx, deployRes.Address, receiverAddress, big.NewInt(10))
	if err != nil {
		t.Fatalf("transfer failed: %v", err)
	}

	// wait for mined
	_, err = bind.WaitMined(ctx, client.Client, tx)
	if err != nil {
		t.Fatalf("wait approve mined failed: %v", err)
	}

	// get owner balance after transfer
	balance, err = service.GetBalance(ctx, deployRes.Address)
	if err != nil {
		t.Fatalf("get balance failed: %v", err)
	}
	log.Printf("owner balance after transfer: %s", balance)

	// get receiver balance after transfer
	balance, err = service.GetBalanceOf(ctx, deployRes.Address, receiverAddress)
	if err != nil {
		t.Fatalf("get balance failed: %v", err)
	}
	log.Printf("receiver balance after transfer: %s", balance)

}

func Test_ApproveAndTransferFrom(t *testing.T) {
	ownerPK := "0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"
	spenderPK := "0x59c6995e998f97a5a0044966f0945389dc9e86dae88c7a8412f4603b6b78690d"
	receiverPK := "0x5de4111afa1a4b94908f83103eb1f1706367c2e68ca870fc3fb9a804cdab365a"

	client, err := eth.NewClient(ctx, config.Get().Scanners[0].RpcHTTP)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	ownerService := erc20.NewERC20Service(client, ownerPK)

	// deploy contract
	deployRes, err := ownerService.Deploy(big.NewInt(1000))
	if err != nil {
		t.Fatalf("deploy failed: %v", err)
	}

	// wait deploy mined
	_, err = bind.WaitMined(ctx, client.Client, deployRes.Tx)
	if err != nil {
		t.Fatalf("wait deploy mined failed: %v", err)
	}

	ownerBalance, err := ownerService.GetBalance(ctx, deployRes.Address)
	if err != nil {
		t.Fatalf("get owner balance failed: %v", err)
	}

	assert.Equal(t, "1000", ownerBalance.String())
	t.Logf("owner balance: %s", ownerBalance.String())

	// get owner address
	pk, err := crypto.HexToECDSA(strings.TrimPrefix(ownerPK, "0x"))
	if err != nil {
		t.Fatalf("hex to ecdsa failed: %v", err)
	}
	owner := crypto.PubkeyToAddress(pk.PublicKey)

	// get spender address
	pk, err = crypto.HexToECDSA(strings.TrimPrefix(spenderPK, "0x"))
	if err != nil {
		t.Fatalf("hex to ecdsa failed: %v", err)
	}
	spender := crypto.PubkeyToAddress(pk.PublicKey)

	// approve spender to spend tokens
	amount := big.NewInt(100)
	tx, err := ownerService.Approve(ctx, deployRes.Address, spender, amount)
	if err != nil {
		t.Fatalf("approve failed: %v", err)
	}

	// wait for mined
	_, err = bind.WaitMined(ctx, client.Client, tx)
	if err != nil {
		t.Fatalf("wait approve mined failed: %v", err)
	}

	// get allowance
	allowance, err := ownerService.GetAllowance(ctx, deployRes.Address, spender)
	if err != nil {
		t.Fatalf("allowance failed: %v", err)
	}

	assert.Equal(t, amount.String(), allowance.String())
	t.Logf("owner allowance: %s", allowance)

	// get spender address
	pk, err = crypto.HexToECDSA(strings.TrimPrefix(receiverPK, "0x"))
	if err != nil {
		t.Fatalf("hex to ecdsa failed: %v", err)
	}
	receiver := crypto.PubkeyToAddress(pk.PublicKey)

	// spender execute transfer from owner to receiver
	tx, err = erc20.NewERC20Service(client, spenderPK).TransferFrom(ctx, deployRes.Address, owner, receiver, amount)
	if err != nil {
		t.Fatalf("transfer from failed: %v", err)
	}

	// wait for mined
	_, err = bind.WaitMined(ctx, client.Client, tx)
	if err != nil {
		t.Fatalf("wait approve mined failed: %v", err)
	}

	ownerBalance, err = ownerService.GetBalance(ctx, deployRes.Address)
	if err != nil {
		t.Fatalf("get owner balance failed: %v", err)
	}
	assert.Equal(t, "900", ownerBalance.String())
	t.Logf("owner balance after TransferFrom: %s", ownerBalance)

	// get owner allowance
	allowance, err = ownerService.GetAllowance(ctx, deployRes.Address, spender)
	if err != nil {
		t.Fatalf("allowance failed: %v", err)
	}

	assert.Equal(t, "0", allowance.String())
	t.Logf("owner allowance after TransferFrom: %s", allowance)

	receiverBalance, err := ownerService.GetBalanceOf(ctx, deployRes.Address, receiver)
	if err != nil {
		t.Fatalf("get receiver balance failed: %v", err)
	}

	assert.Equal(t, "100", receiverBalance.String())
	t.Logf("receiver balance after TransferFrom: %s", receiverBalance)
}
