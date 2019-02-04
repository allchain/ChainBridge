package client

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

type Chain struct {
	Name string 						`json:"name"`
	Url string 							`json:"url"`
	Id *big.Int 						`json:"id,omitempty"`
	Contract *common.Address 			`json:"contractAddr"`
	GasPrice *big.Int 					`json:"gasPrice"`
	From *common.Address 				`json:"from"`
	Password string 					`json:"password,omitempty"`
	Client *ethclient.Client 			`json:"client,omitempty"`
	Nonce uint64 						`json:"nonce,omitempty"`
	StartBlock *big.Int 				`json:"startBlock,omitempty"`
}

type Withdrawal struct {
	Recipient string
	Value *big.Int
	FromChain string
	TxHash string
	Data string
}

// events to listen for
type Events struct {
	DepositId string
  	CreationId string
 	WithdrawId string
	BridgeFundedId string
	PaidId string
	AuthorityAddedId string
	AuthorityRemovedId string
	ThresholdUpdated string
	SignedForWithdraw string
}