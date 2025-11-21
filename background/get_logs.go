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
	"log"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

const RETRY = 10

// TODO: make it configurable
func GetLogs(address string) {
	client, err := internalEth.NewClient(context.Background(), "http://localhost:8545")
	if err != nil {
		panic(err)
	}

	defer client.Close()

	// TODO: get last block from db
	lastBlock := uint64(0)
	size := uint64(100)

	ticker := time.NewTicker(time.Second * 5)
	retry := 0
	for range ticker.C {

		now := time.Now()

		if retry > RETRY {
			log.Printf("get logs error, retry: %d", retry)
			break
		}

		blockSync, err := blocksync.GetBlockSyncStatus(context.Background(), db.GetMysql(db.EVENT_DB), address)
		if err != nil {
			log.Printf("get block sync status error: %v", err)
			retry++
			time.Sleep(time.Second * 5)
			continue
		}

		if blockSync == nil {
			blockSync = new(model.BlockSyncStatus)
		}

		lastBlock = blockSync.LastSyncNumber

		eventLogs, err := client.GetLogs(internalEth.GetLogsParams{
			FromBlock: lastBlock,
			ToBlock:   lastBlock + size,
			Addresses: []common.Address{common.HexToAddress(address)},
			Topics:    [][]common.Hash{},
		})
		if err != nil {
			log.Printf("get logs error: %v", err)
			retry++
			time.Sleep(time.Second * 5)
			continue
		}

		log.Printf("get logs - lastBlock: %d, size: %d, eventLogs count: %d", lastBlock, size, len(eventLogs))

		if len(eventLogs) == 0 {
			continue
		}

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
			}
		}

		eventDB := db.GetMysql(db.EVENT_DB)

		if err := utils.NewTx(eventDB).Exec(context.Background(),
			func(ctx context.Context, tx *sql.Tx) error {
				return eventlog.TxUpsertLog(ctx, tx, logs...)
			},
			func(ctx context.Context, tx *sql.Tx) error {
				return blocksync.TxUpsertBlockSync(ctx, tx, &model.BlockSyncStatus{
					Address:                address,
					LastSyncNumber:         lastBlock + uint64(len(eventLogs)),
					LastSyncTimestamp:      now,
					LastFinalizedNumber:    lastBlock + uint64(len(eventLogs)),
					LastFinalizedTimestamp: now,
					UpdatedAt:              now,
				})
			},
		); err != nil {
			log.Printf("upsert log error: %v", err)
			retry++
			time.Sleep(time.Second * 5)
			continue
		}

		if len(eventLogs) < int(size) {
			lastBlock += uint64(len(eventLogs))
		} else {
			lastBlock += size
		}

	}
}
