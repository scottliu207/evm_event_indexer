package eth

import (
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/crypto"
)

func NewTransactor(chainID *big.Int, privHex string) (*bind.TransactOpts, error) {
	pk, err := crypto.HexToECDSA(strings.TrimPrefix(privHex, "0x"))
	if err != nil {
		return nil, err
	}
	return bind.NewKeyedTransactorWithChainID(pk, chainID)

}
