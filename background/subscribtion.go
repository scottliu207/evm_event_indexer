package background

import (
	"context"
	internalCnf "evm_event_indexer/internal/config"
	internalEth "evm_event_indexer/internal/eth"
	internalStorage "evm_event_indexer/internal/storage"

	"evm_event_indexer/service/repo/blocksync"
	"fmt"
	"log/slog"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"
)

func Subscription() {
	backoff := internalCnf.Get().Backoff

	for {

		err := func() error {
			ctx := context.Background()

			headers := make(chan *types.Header)
			client, err := internalEth.NewClient(ctx, internalCnf.Get().EthRpcWS)
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
			backoff = min(backoff*2, internalCnf.Get().MaxBackoff)
			continue
		}
		backoff = internalCnf.Get().Backoff
	}

}

func subscription(ctx context.Context, sub ethereum.Subscription, headers chan *types.Header) error {

	if len(internalCnf.Get().ContractsAddress) == 0 {
		slog.Warn("no contract addresses configured, skip subscription reorg check")
		return nil
	}

	client, err := internalEth.NewClient(ctx, internalCnf.Get().EthRpcHTTP)
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

			// get last sync block number
			bcMap, err := blocksync.GetBlockSyncMap(ctx, internalStorage.GetMysql(internalCnf.Get().MySQL.EventDBS.DBName), internalCnf.Get().ContractsAddress)
			if err != nil {
				slog.Error("get block sync error", slog.Any("err", err))
				return fmt.Errorf("get block sync error, %w", err)
			}

			for _, addr := range internalCnf.Get().ContractsAddress {
				bc, ok := bcMap[addr]
				if !ok || bc == nil || bc.LastSyncNumber == 0 || bc.LastSyncHash == "" {
					slog.Warn("no block sync found, skipping reorg process", slog.Any("address", addr))
					continue
				}

				// get the latest sync block header
				lbHeader, err := client.GetHeaderByNumber(bc.LastSyncNumber)
				if err != nil {
					slog.Error("get header error", slog.Any("height", bc.LastSyncNumber), slog.Any("err", err))
					continue
				}

				// check if reorg happened
				if lbHeader.Hash().String() == bc.LastSyncHash {
					continue
				}

				reorgWindow := uint64(internalCnf.Get().ReorgWindow)

				rollbackNum := uint64(0)
				if bc.LastSyncNumber > reorgWindow {
					rollbackNum = bc.LastSyncNumber - reorgWindow
				}

				ReorgProducer(&reorgMsg{
					RollbackNumber:  rollbackNum,
					Backoff:         internalCnf.Get().Backoff,
					ContractAddress: addr,
					Retry:           0,
				})

				slog.Info("reorg happened",
					slog.Any("block number", header.Number),
					slog.Any("last sync number", bc.LastSyncNumber),
					slog.Any("rollback number", rollbackNum),
					slog.Any("rollback hash", lbHeader.Hash().String()),
				)
			}
		}
	}
}
