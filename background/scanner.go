package background

import (
	"context"
	"database/sql"
	internalCnf "evm_event_indexer/internal/config"
	internalEth "evm_event_indexer/internal/eth"
	internalStorage "evm_event_indexer/internal/storage"

	"evm_event_indexer/service/model"
	"evm_event_indexer/service/repo/blocksync"
	"evm_event_indexer/service/repo/eventlog"
	"evm_event_indexer/utils"
	"log/slog"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

const ERC20_TRANSFER_EVENT = "Transfer(address,address,uint256)"
const ERC20_APPROVAL_EVENT = "Approval(address,address,uint256)"

func LogScanner(address string) {

	ctx := context.Background()

	client, err := internalEth.NewClient(ctx, internalCnf.Get().EthRpcHTTP)
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

	syncBlock := uint64(0)
	size := uint64(100)

	ticker := time.NewTicker(internalCnf.Get().LogScannerInterval)
	defer ticker.Stop()

	retry := 0
	for range ticker.C {

		now := time.Now()

		if retry > internalCnf.Get().Retry {
			slog.Error("get logs error, exceed retry limit")
			break
		}

		bc, err := blocksync.GetBlockSync(context.Background(), internalStorage.GetMysql(internalCnf.Get().EventDB), address)
		if err != nil {
			slog.Error("get block sync status error",
				slog.Any("err", err),
				slog.Any("retry", retry),
			)
			retry++
			time.Sleep(time.Second * 5)
			continue
		}

		if bc == nil {
			bc = new(model.BlockSync)
		}

		// if there is no sync block, start from 0
		if bc.LastSyncNumber > 0 {
			syncBlock = bc.LastSyncNumber + 1
		}

		latestBlock, err := client.GetBlockNumber()
		if err != nil {
			slog.Error("get current block number error",
				slog.Any("err", err),
				slog.Any("retry", retry),
			)
			retry++
			time.Sleep(time.Second * 5)
			continue
		}

		toBlock := min(syncBlock+size, latestBlock)
		if syncBlock > toBlock {
			slog.Info("no new blocks to scan",
				slog.Any("lastSyncNumber", bc.LastSyncNumber),
				slog.Any("latestBlock", latestBlock),
			)
			retry = 0
			continue
		}

		eventLogs, err := client.GetLogs(internalEth.GetLogsParams{
			FromBlock: syncBlock,
			ToBlock:   toBlock,
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

		header, err := client.GetHeaderByNumber(toBlock)
		if err != nil {
			slog.Error("get block header error",
				slog.Any("blockNumber", toBlock),
				slog.Any("err", err),
			)
			retry++
			time.Sleep(time.Second * 5)
			continue
		}

		newSyncNumber := toBlock
		newSyncHash := header.Hash().Hex()

		slog.Info("event logs info",
			slog.Any("lastSyncNumber", syncBlock),
			slog.Any("toBlock", toBlock),
			slog.Any("total logs count", len(eventLogs)),
		)

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

			slog.Info("event log info",
				slog.Any("log", v),
			)
		}

		if err := utils.NewTx(internalStorage.GetMysql(internalCnf.Get().EventDB)).Exec(ctx,
			func(ctx context.Context, tx *sql.Tx) error {
				if len(logs) == 0 {
					return nil
				}
				return eventlog.TxUpsertLog(ctx, tx, logs...)
			},
			func(ctx context.Context, tx *sql.Tx) error {
				return blocksync.TxUpsertBlock(ctx, tx, &model.BlockSync{
					Address:        address,
					LastSyncNumber: newSyncNumber,
					LastSyncHash:   newSyncHash,
					UpdatedAt:      now,
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

		// reset retry
		retry = 0
	}

}
