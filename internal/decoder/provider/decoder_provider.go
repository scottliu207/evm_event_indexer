package provider

import (
	"evm_event_indexer/service/model"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

type EventDecoder interface {
	EventName() string
	Decode(data *model.Log) (map[string]string, error)
}

type DecoderProvider struct {
	decoders map[common.Hash]EventDecoder
}

func NewDecoderProvider() *DecoderProvider {
	return &DecoderProvider{
		decoders: make(map[common.Hash]EventDecoder),
	}
}

func (p *DecoderProvider) Register(signature string, decoder EventDecoder) {
	hash := crypto.Keccak256Hash([]byte(signature))
	p.decoders[hash] = decoder
}

func (p *DecoderProvider) Decode(log *model.Log) (name string, args map[string]string, err error) {
	if log == nil {
		return "", nil, fmt.Errorf("invalid log")
	}

	decoder, ok := p.decoders[common.HexToHash(log.Topic0)]
	if !ok {
		return "", nil, fmt.Errorf("decoder not found")
	}

	args, err = decoder.Decode(log)
	if err != nil {
		return "", nil, fmt.Errorf("decode event failed, signature: %s, error: %w", log.Topic0, err)
	}

	return decoder.EventName(), args, nil
}
