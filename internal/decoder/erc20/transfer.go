package erc20

import (
	"evm_event_indexer/internal/decoder/provider"
	"evm_event_indexer/service/model"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

var _ provider.EventDecoder = (*TransferDecoder)(nil)

type TransferDecoder struct{}

func (d *TransferDecoder) EventName() string {
	return "Transfer"
}

func (d *TransferDecoder) Decode(log *model.Log) (map[string]string, error) {

	// Transfer(address indexed from, address indexed to, uint256 value)
	// topic[0] = signature
	// topic[1] = from (indexed)
	// topic[2] = to (indexed)
	// data = value
	if log == nil {
		return nil, fmt.Errorf("invalid log")
	}

	topics := []string{log.Topic0, log.Topic1, log.Topic2}
	if len(topics) != 3 {
		return nil, fmt.Errorf("event Transfer: expected 3 topics, got %d", len(topics))
	}

	from := common.HexToHash(topics[1])
	to := common.HexToHash(topics[2])
	value := new(big.Int).SetBytes(log.Data)

	return map[string]string{
		"from":  from.String(),
		"to":    to.String(),
		"value": value.String(),
	}, nil
}
