package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// tracking the latest block number synced by the indexer
	LatestSyncedBlock = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "indexer_latest_synced_block",
		Help: "The latest block number synced by the indexer",
	}, []string{"chain_id", "address"})

	// tracking the total number of logs indexed by the indexer
	TotalLogsIndexed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "indexer_total_logs_indexed",
		Help: "The total number of logs indexed",
	}, []string{"chain_id", "address"})

	// tracking the duration and status of RPC requests
	RpcRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "indexer_rpc_duration_seconds",
		Help:    "Duration of RPC requests",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "status"}) // status: success/failure/noop

	// tracking the duration and status of each scan batch
	ScanBatchDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "indexer_scan_batch_duration_seconds",
		Help:    "Duration of each scan batch",
		Buckets: prometheus.DefBuckets,
	}, []string{"chain_id", "address", "status"}) // status: success/failure/noop

	// tracking the duration and status of DB write operations
	DBWriteDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "indexer_db_write_duration_seconds",
		Help:    "Duration of DB write operations",
		Buckets: prometheus.DefBuckets,
	}, []string{"operation", "status"}) // status: success/failure

	// tracking the number of failed DB write operations
	DBWriteErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "indexer_db_write_errors_total",
		Help: "Total number of failed DB write operations",
	}, []string{"operation"})
)
