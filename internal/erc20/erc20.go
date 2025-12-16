package erc20

import (
	"context"
	erc20 "evm_event_indexer/generated"
	internalEth "evm_event_indexer/internal/eth"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type (
	ERC20Service struct {
		c          *internalEth.Client
		privateKey string
	}

	DeployResult struct {
		Address common.Address
		Tx      *types.Transaction
	}
)

func NewERC20Service(client *internalEth.Client, privateKey string) *ERC20Service {
	return &ERC20Service{c: client, privateKey: privateKey}
}

func (i *ERC20Service) Deploy(initialSupply *big.Int) (*DeployResult, error) {

	transactor, err := internalEth.NewTransactor(i.c.GetChainID(), i.privateKey)
	if err != nil {
		return nil, fmt.Errorf("new transactor failed: %w", err)
	}

	// Optional gas config; leave nil to let the client suggest
	transactor.Value = big.NewInt(0)

	// 3) Deploy the contract
	addr, tx, _, err := erc20.DeployBasicErc20(transactor, i.c.Client, "MyToken", "MTK", initialSupply)
	if err != nil {
		return nil, fmt.Errorf("deploy failed: %w", err)
	}

	return &DeployResult{
		Address: addr,
		Tx:      tx,
	}, nil
}

func (i *ERC20Service) GetBalance(ctx context.Context, contractAddr common.Address, ownerAddr common.Address) (*big.Int, error) {
	contract, err := internalEth.NewERC20(contractAddr, i.c.Client)
	if err != nil {
		return nil, fmt.Errorf("get balance failed: %w", err)
	}

	return contract.Instance.BalanceOf(&bind.CallOpts{Context: ctx}, ownerAddr)
}

func (i *ERC20Service) Transfer(ctx context.Context, addr common.Address, amount *big.Int) (*types.Transaction, error) {
	auth, err := internalEth.NewTransactor(i.c.GetChainID(), i.privateKey)
	if err != nil {
		return nil, fmt.Errorf("new transactor failed: %w", err)
	}

	contract, err := internalEth.NewERC20(addr, i.c.Client)
	if err != nil {
		return nil, fmt.Errorf("new contract failed: %w", err)
	}

	return contract.Instance.Transfer(auth, addr, amount)
}

func (i *ERC20Service) Approve(ctx context.Context, addr common.Address, amount *big.Int) (*types.Transaction, error) {
	auth, err := internalEth.NewTransactor(i.c.GetChainID(), i.privateKey)
	if err != nil {
		return nil, fmt.Errorf("new transactor failed: %w", err)
	}

	contract, err := internalEth.NewERC20(addr, i.c.Client)
	if err != nil {
		return nil, fmt.Errorf("new contract failed: %w", err)
	}

	return contract.Instance.Approve(auth, addr, amount)
}

func (i *ERC20Service) TransferFrom(ctx context.Context, addr common.Address, from common.Address, to common.Address, amount *big.Int) (*types.Transaction, error) {
	auth, err := internalEth.NewTransactor(i.c.GetChainID(), i.privateKey)
	if err != nil {
		return nil, fmt.Errorf("new transactor failed: %w", err)
	}

	contract, err := internalEth.NewERC20(addr, i.c.Client)
	if err != nil {
		return nil, fmt.Errorf("new contract failed: %w", err)
	}

	return contract.Instance.TransferFrom(auth, from, to, amount)
}
