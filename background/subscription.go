package background

import (
	"context"
	"evm_event_indexer/internal/config"

	"evm_event_indexer/internal/eth"
	"evm_event_indexer/internal/storage"

	"evm_event_indexer/service/repo/blocksync"
	"fmt"
	"log/slog"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"
)

func Subscription() {
	backoff := config.Get().Backoff

	for {

		err := func() error {
			ctx := context.Background()

			headers := make(chan *types.Header)
			client, err := eth.NewClient(ctx, config.Get().EthRpcWS)
			if err != nil {
				return err
			}

			defer client.Close()

			sub, err := client.Subscribe(headers)
			if err != nil {
				return err
			}

			defer sub.Unsubscribe()

			if err := subscription(ctx, sub, headers); err != nil {
				return err
			}

			return nil
		}()
		if err != nil {
			slog.Error("subscription error occurred, waiting to retry", slog.Any("err", err))
			time.Sleep(backoff)
			backoff = min(backoff*2, config.Get().MaxBackoff)
			continue
		}
		backoff = config.Get().Backoff
	}

}

func subscription(ctx context.Context, sub ethereum.Subscription, headers chan *types.Header) error {

	if len(config.Get().Scanner) == 0 {
		slog.Warn("no contract addresses configured, skip subscription reorg check")
		return nil
	}

	client, err := eth.NewClient(ctx, config.Get().EthRpcHTTP)
	if err != nil {
		return err
	}

	defer client.Close()

	for {
		select {
		case err := <-sub.Err():
			slog.Error("subscription error", slog.Any("err", err))
			return fmt.Errorf("subscription error, %w", err)
		case header, ok := <-headers:
			if !ok {
				slog.Error("headers channel closed")
				return fmt.Errorf("headers channel closed")
			}

			slog.Info("new block", slog.Any("block number", header.Number), slog.Any("new", header))

			addresses := []string{}
			for _, scanner := range config.Get().Scanner {
				addresses = append(addresses, scanner.Address)
			}

			// get last sync block number
			bcMap, err := blocksync.GetBlockSyncMap(ctx, storage.GetMysql(config.Get().MySQL.EventDBS.DBName), addresses)
			if err != nil {
				slog.Error("get block sync error", slog.Any("err", err))
				return fmt.Errorf("get block sync error, %w", err)
			}

			for _, addr := range config.Get().Scanner {
				bc, ok := bcMap[addr.Address]
				if !ok || bc == nil || bc.LastSyncNumber == 0 || bc.LastSyncHash == "" {
					slog.Warn("no block sync found, skipping reorg process", slog.Any("address", addr))
					continue
				}

				// get the latest sync block header on chain
				lbHeader, err := client.GetHeaderByNumber(bc.LastSyncNumber)
				if err != nil {
					slog.Error("get header error", slog.Any("height", bc.LastSyncNumber), slog.Any("err", err))
					continue
				}

				// check if the latest sync block hash is the same as the block hash on chain
				if lbHeader.Hash().String() == bc.LastSyncHash {
					continue
				}

				ReorgProducer(&reorgMsg{
					LastSyncNumber:  bc.LastSyncNumber,
					Backoff:         config.Get().Backoff,
					ContractAddress: addr.Address,
					Retry:           0,
				})

				slog.Info("reorg happened",
					slog.Any("block number", header.Number),
					slog.Any("last sync number", bc.LastSyncNumber),
					slog.Any("rollback hash", lbHeader.Hash().String()),
				)
			}
		}
	}
}
