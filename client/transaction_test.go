package client

import (
	"errors"
	"io/ioutil"
	"path/filepath"
	"testing"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/accounts/keystore"
)

// deploys the bridge contract and setups Chain type with name, url, client, from address, password, 
// chain ID, gas price and contract address
func setup() (*Chain, error) {
	chain := new(Chain)
	chain.Name = "testnet"
	chain.Url = "https://rinkeby.infura.io"

	client, err := ethclient.Dial(chain.Url)
	if err != nil {
		return nil, err
	}

	chain.Client = client

	fp, err := filepath.Abs("../keystore")
	if err != nil {
		return nil, err
	}

	ks := keystore.NewKeyStore(fp, keystore.StandardScryptN, keystore.StandardScryptP)
	ksaccounts := ks.Accounts()
	if len(ksaccounts) == 0 {
		return nil, errors.New("no accounts in keystore")
	}

	a := common.HexToAddress("0x8f9b540b19520f8259115a90e4b4ffaeac642a30")
	chain.From = &a

	fp, err = filepath.Abs("../solidity/build/Bridge.bin")
	if err != nil {
		return nil, err
	}

	bin, err := ioutil.ReadFile(fp)
	if err != nil {
		return nil, err
	}

	chain.Password = "password"
	chain.Id = big.NewInt(0)
	chain.GasPrice = big.NewInt(10000000)

	address, err := Deploy(*chain, ks, bin, "Bridge")
	if err != nil {
		return nil, err
	}

	chain.Contract = &address

	return chain, nil
}

func TestSendTx(t *testing.T) {
	chain, err := setup()
	if err != nil {
		t.Fatal(err)
	}

	_, err = SendTx(chain, big.NewInt(0), []byte{})	
	if err != nil {
		t.Fatal(err)
	}
}

// func TestWithdraw(t *testing.T) {
// 	chain := new(Chain)
// 	chain.Name = "testnet"

// 	client, err := core.NewConnection()
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	chain.Client = client

// 	contract, err := core.ContractAddress("Bridge", "testnet")
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	*chain.Contract = contract

// 	w := new(Withdrawal)
// 	w.Recipient = "0xc84233646c0aa920c66a2e220ea790e548b72f9e"
// 	w.Value = big.NewInt(7777)
// 	w.FromChain = "4"
// 	err = Withdraw(chain, w)
// 	if err != nil {
// 		t.Error(err)
// 	}
// }

// func test_AddAuthority(t *testing.T) {
// 	chain := new(Chain)
// 	chain.Name = "testnet"

// 	client, err := core.NewConnection("http://localhost:8545")
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	chain.Client = client

// 	contract, err := core.ContractAddress("Bridge", "testnet")
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	*chain.Contract = contract

// 	err = AddAuthority(chain, "0xc84233646c0aa920c66a2e220ea790e548b72f9e")
// 	if err != nil {
// 		t.Error(err)
// 	}
// }

// func test_Deposit(t *testing.T) {
// 	chain := new(Chain)
// 	chain.Name = "testnet"

// 	client, err := core.NewConnection("http://localhost:8545")
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	chain.Client = client

// 	contract, err := core.ContractAddress("Bridge", "testnet")
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	*chain.Contract = contract

// 	err = Deposit(chain, big.NewInt(77), "5")
// 	if err != nil {
// 		t.Error(err)
// 	}
// }