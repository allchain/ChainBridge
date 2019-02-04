package client

import (
	"testing"
	"math/big"

	"github.com/ChainSafeSystems/leth/core"
)

func test_withdraw(t *testing.T) {
	chain := new(Chain)

	chain.Name = "testnet"

	client, err := core.NewConnection("http://localhost:8545")
	if err != nil {
		t.Error(err)
	}
	chain.Client = client

	contract, err := core.ContractAddress("Bridge", "testnet")
	if err != nil {
		t.Error(err)
	}
	*chain.Contract = contract

	w := new(Withdrawal)
	w.Recipient = "0xc84233646c0aa920c66a2e220ea790e548b72f9e"
	w.Value = big.NewInt(7777)
	w.FromChain = "4"
	err = Withdraw(chain, w)
	if err != nil {
		t.Error(err)
	}
}