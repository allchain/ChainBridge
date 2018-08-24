package client 

import (
	"math/big"
	"github.com/ChainSafeSystems/geth/common"
)

type Root struct {
	Hash common.Hash
	Contract *common.Address
	Start *big.Int
	End *big.Int
}

