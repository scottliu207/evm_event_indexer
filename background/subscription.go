package background

import (
	"context"
	"evm_event_indexer/internal/config"
	"evm_event_indexer/service"

	"evm_event_indexer/internal/eth"

	"fmt"
	"log/slog"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"
)

var _ Worker = (*Subscription)(nil)

type Subscription struct {
	rpcWS   string
	rpcHTTP string
	address []string
}

func NewSubscription(rpcHTTP string, rpcWS string, address []string) *Subscription {
	return &Subscription{
		rpcHTTP: rpcHTTP,
		rpcWS:   rpcWS,
		address: address,
	}
}

func (s *Subscription) Run(ctx context.Context) error {
	backoff := config.Get().Backoff
	for {

		err := func() error {
			headers := make(chan *types.Header)
			client, err := eth.NewClient(ctx, s.rpcWS)
			if err != nil {
				return err
			}

			defer client.Close()

			sub, err := client.Subscribe(headers)
			if err != nil {
				return err
			}

			defer sub.Unsubscribe()

			if err := s.subscription(ctx, sub, headers); err != nil {
				return err
			}

			return nil
		}()

		if ctx.Err() != nil {
			return ctx.Err()
		}

		if err != nil {
			slog.Error("subscription error occurred, waiting to retry", slog.Any("eror", err))
			time.Sleep(backoff)
			backoff = min(backoff*2, config.Get().MaxBackoff)
			continue
		}

		// reset backoff
		backoff = config.Get().Backoff

	}
}

func (s *Subscription) subscription(ctx context.Context, sub ethereum.Subscription, headers chan *types.Header) error {

	if len(config.Get().Scanners) == 0 {
		slog.Warn("no contract addresses configured, skip subscription reorg check")
		return nil
	}

	client, err := eth.NewClient(ctx, s.rpcWS)
	if err != nil {
		return err
	}

	defer client.Close()

	for {
		select {
		case <-ctx.Done():
			return nil // context done, exit the loop
		case err := <-sub.Err():
			slog.Error("subscription error", slog.Any("error", err))
			return fmt.Errorf("subscription error, %w", err)
		case header, ok := <-headers:
			if !ok {
				slog.Error("headers channel closed")
				return fmt.Errorf("headers channel closed")
			}

			slog.Info("new block", slog.Any("block number", header.Number), slog.Any("new", header))

			// get last sync block number
			bcMap, err := service.GetBlockSyncMap(ctx, client.GetChainID().Int64(), s.address)
			if err != nil {
				slog.Error("get block sync error", slog.Any("error", err))
				return fmt.Errorf("get block sync error, %w", err)
			}

			for _, addr := range s.address {
				bc, ok := bcMap[addr]
				if !ok || bc == nil || bc.LastSyncNumber == 0 || bc.LastSyncHash == "" {
					slog.Warn("no block sync found, skipping reorg process", slog.Any("address", addr))
					continue
				}

				// get the latest sync block header on chain
				lbHeader, err := client.GetHeaderByNumber(bc.LastSyncNumber)
				if err != nil {
					slog.Error("get header error", slog.Any("height", bc.LastSyncNumber), slog.Any("error", err))
					continue
				}

				// check if the latest sync block hash is the same as the block hash on chain
				if lbHeader.Hash().String() == bc.LastSyncHash {
					continue
				}

				ReorgProducer(&reorgMsg{
					LastSyncNumber:  bc.LastSyncNumber,
					Backoff:         config.Get().Backoff,
					ContractAddress: addr,
					RpcHttp:         s.rpcHTTP,
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
