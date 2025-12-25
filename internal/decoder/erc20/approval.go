package erc20

import (
	"evm_event_indexer/internal/decoder/provider"
	"evm_event_indexer/service/model"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

var _ provider.EventDecoder = (*ApprovalDecoder)(nil)

type ApprovalDecoder struct{}

func (d *ApprovalDecoder) EventName() string {
	return "Approval"
}

func (d *ApprovalDecoder) Decode(log *model.Log) (map[string]string, error) {

	// Approval(address indexed owner, address indexed spender, uint256 value)
	// topic[0] = signature
	// topic[1] = owner (indexed)
	// topic[2] = spender (indexed)
	// data = value
	if log == nil {
		return nil, fmt.Errorf("invalid log")
	}

	topics := []string{log.Topic0, log.Topic1, log.Topic2}

	if len(topics) != 3 {
		return nil, fmt.Errorf("event Approval: expected 3 topics, got %d", len(topics))
	}

	owner := common.HexToHash(topics[1])
	spender := common.HexToHash(topics[2])
	value := new(big.Int).SetBytes(log.Data)

	return map[string]string{
		"owner":   owner.String(),
		"spender": spender.String(),
		"value":   value.String(),
	}, nil
}
