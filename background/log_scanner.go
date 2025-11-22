package background

import (
	"context"
	"database/sql"
	internalEth "evm_event_indexer/internal/eth"
	"evm_event_indexer/service/db"
	"evm_event_indexer/service/model"
	"evm_event_indexer/service/repo/blocksync"
	"evm_event_indexer/service/repo/eventlog"
	"evm_event_indexer/utils"
	"log/slog"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

const RETRY = 10
const ERC20_TRANSFER_EVENT = "Transfer(address,address,uint256)"
const ERC20_APPROVAL_EVENT = "Approval(address,address,uint256)"

// TODO: make it configurable
// TODO: reorg
func LogScanner(address string) {

	scanInterval, err := time.ParseDuration(os.Getenv("LOG_SCANNER_INTERVAL"))
	if err != nil {
		panic(err)
	}

	ctx := context.Background()

	client, err := internalEth.NewClient(ctx, os.Getenv("ETH_RPC_HTTP"))
	if err != nil {
		panic(err)
	}

	defer client.Close()

	topics := [][]common.Hash{
		// topics, filetering Transfer and Approval events
		{
			common.Hash(crypto.Keccak256([]byte(ERC20_TRANSFER_EVENT))),
			common.Hash(crypto.Keccak256([]byte(ERC20_APPROVAL_EVENT))),
		},
		// from
		// to
	}

	addresses := []common.Address{common.HexToAddress(address)}

	// TODO: get last block from db
	lastBlock := uint64(0)
	size := uint64(100)

	ticker := time.NewTicker(scanInterval)
	retry := 0
	for range ticker.C {

		now := time.Now()

		if retry > RETRY {
			slog.Error("get logs error, exceed retry limit")
			break
		}

		blockSync, err := blocksync.GetBlockSync(context.Background(), db.GetMysql(db.EVENT_DB), address)
		if err != nil {
			slog.Error("get block sync status error",
				slog.Any("err", err),
				slog.Any("retry", retry),
			)
			retry++
			time.Sleep(time.Second * 5)
			continue
		}

		if blockSync == nil {
			blockSync = new(model.BlockSync)
		}

		lastBlock = blockSync.LastSyncNumber + 1

		eventLogs, err := client.GetLogs(internalEth.GetLogsParams{
			FromBlock: lastBlock,
			ToBlock:   lastBlock + size,
			Addresses: addresses,
			Topics:    topics,
		})
		if err != nil {
			slog.Error("get logs error",
				slog.Any("err", err),
				slog.Any("retry", retry),
			)
			retry++
			time.Sleep(time.Second * 5)
			continue
		}

		slog.Info("event logs info",
			slog.Any("lastSyncNumber", lastBlock),
			slog.Any("fromBlock", lastBlock),
			slog.Any("toBlock", lastBlock+size),
			slog.Any("total logs count", len(eventLogs)),
		)

		newLastBlock := lastBlock
		logs := make([]*model.Log, len(eventLogs))
		for i, v := range eventLogs {

			topics := make(model.Topics, len(v.Topics))
			for j, topic := range v.Topics {
				topics[j] = topic.Hex()
			}

			logs[i] = &model.Log{
				Address:        v.Address.Hex(),
				BlockHash:      v.BlockHash.Hex(),
				BlockNumber:    v.BlockNumber,
				Topics:         &topics,
				TxIndex:        v.TxIndex,
				LogIndex:       v.Index,
				TxHash:         v.TxHash.Hex(),
				Data:           v.Data,
				BlockTimestamp: time.Unix(int64(v.BlockTimestamp), 0),
				CreatedAt:      now,
			}

			newLastBlock = v.BlockNumber
		}

		if err := utils.NewTx(db.GetMysql(db.EVENT_DB)).Exec(ctx,
			func(ctx context.Context, tx *sql.Tx) error {
				if len(logs) == 0 {
					return nil
				}
				return eventlog.TxUpsertLog(ctx, tx, logs...)
			},
			func(ctx context.Context, tx *sql.Tx) error {
				return blocksync.TxUpsertBlock(ctx, tx, &model.BlockSync{
					Address:                address,
					LastSyncNumber:         lastBlock + uint64(len(eventLogs)-1),
					LastSyncTimestamp:      now,
					LastFinalizedNumber:    lastBlock + uint64(len(eventLogs)-1),
					LastFinalizedTimestamp: now,
					UpdatedAt:              now,
				})
			},
		); err != nil {
			slog.Error("upsert log error",
				slog.Any("err", err),
				slog.Any("retry", retry),
			)
			retry++
			time.Sleep(time.Second * 5)
			continue
		}

		lastBlock = newLastBlock

	}
}
