package decoder_test

import (
	"evm_event_indexer/internal/decoder/erc20"
	"evm_event_indexer/internal/decoder/provider"
	"evm_event_indexer/internal/enum"
	"evm_event_indexer/service/model"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

func Test_Decoder(t *testing.T) {
	decoder := provider.NewDecoderProvider()
	decoder.Register("Transfer(address,address,uint256)", &erc20.TransferDecoder{})
	// decoder.Register("Approval(address,address,uint256)", &erc20.ApprovalDecoder{})
	name, args, err := decoder.Decode(&model.Log{
		ChainType:      enum.CHOther,
		Address:        "0x5FbDB2315678afecb367f032d93F642f64180aa3",
		BlockHash:      "0x55e7e0a0bf7cca093b1d43eba319681b94e4439298a47e8e179616769095a9dd",
		BlockNumber:    2,
		TxHash:         "0x9a282b02307a238c439f61a70d81d239b5b22507e55b858468fe0750da5cc8b0",
		TxIndex:        0,
		LogIndex:       0,
		BlockTimestamp: time.Now(),
		Topics: &model.Topics{
			common.HexToHash("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"),
			common.HexToHash("0x000000000000000000000000f39fd6e51aad88f6f4ce6ab8827279cfffb92266"),
			common.HexToHash("0x0000000000000000000000005fbdb2315678afecb367f032d93f642f64180aa3"),
		},
		Data: []byte("\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x01"),
	})

	assert.NoError(t, err)
	assert.Equal(t, "Transfer", name)
	assert.Equal(t, "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266", args["from"])
	assert.Equal(t, "0x5FbDB2315678afecb367f032d93F642f64180aa3", args["to"])
	assert.Equal(t, big.NewInt(1).String(), args["value"])
}
