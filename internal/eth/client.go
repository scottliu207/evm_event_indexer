package eth

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

type (
	Client struct {
		Client  *ethclient.Client
		chainID *big.Int
		ctx     context.Context
	}

	GetLogsParams struct {
		FromBlock uint64
		ToBlock   uint64
		Addresses []common.Address
		Topics    [][]common.Hash
	}
)

func NewClient(ctx context.Context, rpcUrl string) (*Client, error) {
	// 1) Connect to an RPC endpoint (e.g., Anvil/Hardhat/Ganache or real node)
	client, err := ethclient.DialContext(ctx, rpcUrl)
	if err != nil {
		return nil, fmt.Errorf("dial rpc: %w", err)
	}

	chainID, err := client.NetworkID(ctx)
	if err != nil {
		return nil, fmt.Errorf("network id: %w", err)
	}

	return &Client{
		Client:  client,
		chainID: chainID,
		ctx:     ctx,
	}, nil
}

func (i *Client) Close() {
	i.Client.Close()
}

func (i Client) GetChainID() *big.Int {
	return i.chainID
}

func (i Client) GetBlockNumber() (uint64, error) {
	return i.Client.BlockNumber(i.ctx)
}

func (i Client) GetLogs(params GetLogsParams) ([]types.Log, error) {
	return i.Client.FilterLogs(i.ctx, ethereum.FilterQuery{
		FromBlock: big.NewInt(int64(params.FromBlock)),
		ToBlock:   big.NewInt(int64(params.ToBlock)),
		Addresses: params.Addresses,
		Topics:    params.Topics,
	})
}

func (i Client) Subscribe(headers chan<- *types.Header) (ethereum.Subscription, error) {
	return i.Client.SubscribeNewHead(i.ctx, headers)
}
