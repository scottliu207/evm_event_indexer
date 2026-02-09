package tools

import (
	"evm_event_indexer/internal/metrics"
	"time"
)

// ObserveRPC observes the duration of an RPC request and the status of the request
func ObserveRPC(method string, start time.Time, err error) {
	status := statusFromErr(err)
	metrics.RpcRequestDuration.WithLabelValues(method, status).Observe(time.Since(start).Seconds())
}

// ObserveDBWrite observes the duration of a DB write operation and the status of the operation
func ObserveDBWrite(method string, start time.Time, err error) {
	status := statusFromErr(err)
	if err != nil {
		metrics.DBWriteErrors.WithLabelValues(method).Inc()
	}
	metrics.DBWriteDuration.WithLabelValues(method, status).Observe(time.Since(start).Seconds())
}

// ObserveScanBatch observes the duration of a scan batch and the status of each scan
func ObserveScanBatch(chainID string, address string, start time.Time, status string) {
	metrics.ScanBatchDuration.WithLabelValues(chainID, address, status).Observe(time.Since(start).Seconds())
}

func statusFromErr(err error) string {
	if err != nil {
		return "failure"
	}
	return "success"
}
