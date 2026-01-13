package background

import (
	"context"
	"errors"
	"evm_event_indexer/internal/config"

	"evm_event_indexer/internal/eth"

	"fmt"
	"log/slog"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

var _ Worker = (*Subscription)(nil)

type Subscription struct {
	rpcWS   string
	rpcHTTP string
	address []string
}

// for now, only handle removed log, new log will be handled by scanner
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
			ch := make(chan types.Log)
			client, err := eth.NewClient(ctx, s.rpcWS)
			if err != nil {
				return err
			}

			defer client.Close()

			addresses := make([]common.Address, len(s.address))
			for i := range s.address {
				addresses[i] = common.HexToAddress(s.address[i])
			}

			sub, err := client.SubscribeFilterLogs(ch, ethereum.FilterQuery{
				Addresses: addresses,
			})
			if err != nil {
				return err
			}

			defer sub.Unsubscribe()

			if err := s.subscription(ctx, sub, ch); err != nil {
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

func (s *Subscription) subscription(ctx context.Context, sub ethereum.Subscription, ch chan types.Log) error {

	if len(s.address) == 0 {
		return errors.New("no contract addresses configured, skip subscription reorg check")
	}

	for {
		select {
		case <-ctx.Done():
			return nil // context done, exit the loop
		case err := <-sub.Err():
			slog.Error("subscription error", slog.Any("error", err))
			return fmt.Errorf("subscription error, %w", err)
		case log, ok := <-ch:

			if !ok {
				slog.Error("headers channel closed")
				return fmt.Errorf("headers channel closed")
			}

			// skip new log, it will be handled by scanner
			if !log.Removed {
				slog.Info("new log, skipping reorg process", slog.Any("address", log.Address.Hex()), slog.Any("block number", log.BlockNumber))
				continue
			}

			ReorgProducer(&reorgMsg{
				Log:             log,
				Backoff:         config.Get().Backoff,
				ContractAddress: log.Address.Hex(),
				RpcHttp:         s.rpcHTTP,
				Retry:           0,
			})

			slog.Info("reorg happened",
				slog.Any("block number", log.BlockNumber),
				slog.Any("txhash", log.TxHash.Hex()),
			)
		}
	}
}
