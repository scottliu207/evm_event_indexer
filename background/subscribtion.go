package background

import (
	"context"
	internalEth "evm_event_indexer/internal/eth"
	"log/slog"
	"os"

	"github.com/ethereum/go-ethereum/core/types"
)

func Subscription() {
	ctx := context.Background()

	client, err := internalEth.NewClient(ctx, os.Getenv("ETH_RPC_WS"))
	if err != nil {
		panic(err)
	}
	defer client.Close()

	headers := make(chan *types.Header)
	sub, err := client.Subscribe(headers)
	if err != nil {
		panic(err)
	}

	for {
		select {
		case <-sub.Err():
			slog.Error("subscription error", slog.Any("err", sub.Err()))
			return
		case header := <-headers:
			slog.Info("new block", slog.Any("block number", header.Number), slog.Any("new", header))
		}

	}
}
