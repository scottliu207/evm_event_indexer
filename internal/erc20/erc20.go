package erc20

import (
	"context"
	erc20 "evm_event_indexer/generated"
	"evm_event_indexer/internal/eth"

	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type (
	ERC20Service struct {
		c          *eth.Client
		privateKey string
	}

	DeployResult struct {
		Address common.Address
		Tx      *types.Transaction
	}
)

func NewERC20Service(client *eth.Client, privateKey string) *ERC20Service {
	return &ERC20Service{c: client, privateKey: privateKey}
}

func (i *ERC20Service) Deploy(initialSupply *big.Int) (*DeployResult, error) {

	transactor, err := eth.NewTransactor(i.c.GetChainID(), i.privateKey)
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

func (i *ERC20Service) Address() (common.Address, error) {
	auth, err := eth.NewTransactor(i.c.GetChainID(), i.privateKey)
	if err != nil {
		return common.Address{}, fmt.Errorf("new transactor failed: %w", err)
	}
	return auth.From, nil
}

func (i *ERC20Service) GetBalance(ctx context.Context, contractAddr common.Address) (*big.Int, error) {
	ownerAddr, err := i.Address()
	if err != nil {
		return nil, err
	}
	return i.GetBalanceOf(ctx, contractAddr, ownerAddr)
}

func (i *ERC20Service) GetBalanceOf(ctx context.Context, contractAddr common.Address, ownerAddr common.Address) (*big.Int, error) {
	contract, err := eth.NewERC20(contractAddr, i.c.Client)
	if err != nil {
		return nil, fmt.Errorf("get balance failed: %w", err)
	}

	return contract.Instance.BalanceOf(&bind.CallOpts{Context: ctx}, ownerAddr)
}

func (i *ERC20Service) Transfer(ctx context.Context, contractAddr common.Address, to common.Address, amount *big.Int) (*types.Transaction, error) {
	auth, err := eth.NewTransactor(i.c.GetChainID(), i.privateKey)
	if err != nil {
		return nil, fmt.Errorf("new transactor failed: %w", err)
	}
	auth.Context = ctx

	contract, err := eth.NewERC20(contractAddr, i.c.Client)
	if err != nil {
		return nil, fmt.Errorf("new contract failed: %w", err)
	}

	return contract.Instance.Transfer(auth, to, amount)
}

func (i *ERC20Service) Approve(ctx context.Context, contractAddr common.Address, spender common.Address, amount *big.Int) (*types.Transaction, error) {
	auth, err := eth.NewTransactor(i.c.GetChainID(), i.privateKey)
	if err != nil {
		return nil, fmt.Errorf("new transactor failed: %w", err)
	}
	auth.Context = ctx

	contract, err := eth.NewERC20(contractAddr, i.c.Client)
	if err != nil {
		return nil, fmt.Errorf("new contract failed: %w", err)
	}

	return contract.Instance.Approve(auth, spender, amount)
}

func (i *ERC20Service) TransferFrom(ctx context.Context, contractAddr common.Address, from common.Address, to common.Address, amount *big.Int) (*types.Transaction, error) {
	auth, err := eth.NewTransactor(i.c.GetChainID(), i.privateKey)
	if err != nil {
		return nil, fmt.Errorf("new transactor failed: %w", err)
	}
	auth.Context = ctx

	contract, err := eth.NewERC20(contractAddr, i.c.Client)
	if err != nil {
		return nil, fmt.Errorf("new contract failed: %w", err)
	}

	return contract.Instance.TransferFrom(auth, from, to, amount)
}

func (i *ERC20Service) GetAllowance(ctx context.Context, contractAddr common.Address, spender common.Address) (*big.Int, error) {
	owner, err := i.Address()
	if err != nil {
		return nil, err
	}
	return i.GetAllowanceOf(ctx, contractAddr, owner, spender)
}

func (i *ERC20Service) GetAllowanceOf(ctx context.Context, contractAddr common.Address, owner common.Address, spender common.Address) (*big.Int, error) {
	contract, err := eth.NewERC20(contractAddr, i.c.Client)
	if err != nil {
		return nil, fmt.Errorf("new contract failed: %w", err)
	}

	return contract.Instance.Allowance(&bind.CallOpts{Context: ctx}, owner, spender)
}
