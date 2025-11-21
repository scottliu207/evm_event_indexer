package eth

import (
	gen "evm_event_indexer/generated"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

type ERC20Contract struct {
	Address  common.Address
	Instance *gen.BasicErc20
}

func NewERC20(addr common.Address, backend bind.ContractBackend) (*ERC20Contract, error) {
	instance, err := gen.NewBasicErc20(addr, backend)
	if err != nil {
		return nil, err
	}
	return &ERC20Contract{Address: addr, Instance: instance}, nil
}
