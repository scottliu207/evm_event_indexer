package decoder

import (
	"evm_event_indexer/internal/config"
	"evm_event_indexer/internal/decoder/erc20"
	"evm_event_indexer/internal/decoder/provider"
	"fmt"
)

var Provider = provider.NewDecoderProvider()
var decoderMap = map[string]provider.EventDecoder{
	"Transfer": &erc20.TransferDecoder{},
	"Approval": &erc20.ApprovalDecoder{},
}

func InitDecoder() {
	for _, decoder := range config.Get().Decoders {
		d, ok := decoderMap[decoder.Name]
		if !ok {
			panic(fmt.Sprintf("unknown decoder %s", decoder.Name))
		}
		Provider.Register(decoder.Signature, d)
	}
}
