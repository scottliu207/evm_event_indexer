package eth

import (
	"context"
	"evm_event_indexer/internal/tools"
	"fmt"
	"math/big"
	"time"

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

	start := time.Now()

	logs, err := i.Client.FilterLogs(i.ctx, ethereum.FilterQuery{
		FromBlock: big.NewInt(int64(params.FromBlock)),
		ToBlock:   big.NewInt(int64(params.ToBlock)),
		Addresses: params.Addresses,
		Topics:    params.Topics,
	})
	tools.ObserveRPC("FilterLogs", start, err)
	if err != nil {
		return nil, fmt.Errorf("filter logs: %w", err)
	}

	return logs, nil
}

func (i Client) Subscribe(headers chan<- *types.Header) (ethereum.Subscription, error) {

	sub, err := i.Client.SubscribeNewHead(i.ctx, headers)
	if err != nil {
		return nil, fmt.Errorf("subscribe new head: %w", err)
	}

	return sub, nil
}

func (i Client) GetHeaderByNumber(number uint64) (*types.Header, error) {

	header, err := i.Client.HeaderByNumber(i.ctx, big.NewInt(int64(number)))
	if err != nil {
		return nil, fmt.Errorf("header by number: %w", err)
	}

	return header, nil
}

func (i Client) SubscribeFilterLogs(log chan<- types.Log, filter ethereum.FilterQuery) (ethereum.Subscription, error) {

	sub, err := i.Client.SubscribeFilterLogs(i.ctx, filter, log)
	if err != nil {
		return nil, fmt.Errorf("subscribe filter logs: %w", err)
	}

	return sub, nil
}
