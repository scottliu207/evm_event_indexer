package background

import (
	"context"
	internalEth "evm_event_indexer/internal/eth"
	"evm_event_indexer/service/db"
	"evm_event_indexer/service/model"
	"evm_event_indexer/service/repo/blocksync"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"
)

const BACKOFF = time.Second
const MAX_BACKOFF = time.Second * 30

func Subscription() {
	backoff := BACKOFF

	for {

		err := func() error {
			ctx := context.Background()

			headers := make(chan *types.Header)
			client, err := internalEth.NewClient(ctx, os.Getenv("ETH_RPC_WS"))
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
			slog.Error("subscription error", slog.Any("err", err))
			time.Sleep(backoff)
			backoff = min(backoff*2, MAX_BACKOFF)
			continue
		}
		backoff = BACKOFF
	}

}

func subscription(ctx context.Context, sub ethereum.Subscription, headers chan *types.Header) error {

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
			bc, err := blocksync.GetBlockSync(ctx, db.GetMysql(db.EVENT_DB), os.Getenv("CONTRACT_ADDRESS"))
			if err != nil {
				slog.Error("get block sync error", slog.Any("err", err))
				return fmt.Errorf("get block sync error, %w", err)
			}

			if bc == nil {
				bc = new(model.BlockSync)
			}

			// if parent hash is the same as last sync hash, no reorg happened
			if bc.LastSyncHash == "" || header.ParentHash.String() == bc.LastSyncHash {
				continue
			}

			reorgWindowInt, err := strconv.Atoi(os.Getenv("REORG_WINDOW"))
			if err != nil {
				panic(err)
			}

			reorgWindow := uint64(reorgWindowInt)

			slog.Info("reorg happened", slog.Any("block number", header.Number), slog.Any("last sync number", bc.LastSyncNumber))
			rollbackHeight := uint64(0)
			if bc.LastSyncNumber > reorgWindow {
				rollbackHeight = bc.LastSyncNumber - reorgWindow
			}
			ReorgProducer(rollbackHeight)
		}

	}
}
