package decoder

import (
	"evm_event_indexer/internal/decoder/erc20"
	"evm_event_indexer/internal/decoder/provider"
)

var Provider = provider.NewDecoderProvider()

func init() {
	Provider.Register("Transfer(address,address,uint256)", &erc20.TransferDecoder{})
	Provider.Register("Approval(address,address,uint256)", &erc20.ApprovalDecoder{})
}
