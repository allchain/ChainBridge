package client

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"io/ioutil"
	"math/big"
	"path/filepath"
	"strings"
	"testing"

	bridgecontract "github.com/ChainSafeSystems/ChainBridge/solidity/Bridge"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

type testAccount struct {
	addr         common.Address
	contract     *bridgecontract.Bridge
	contractAddr common.Address
	backend      *backends.SimulatedBackend
	txOpts       *bind.TransactOpts
}

func setup() (*testAccount, error) {
	genesis := make(core.GenesisAlloc)
	privKey, _ := crypto.GenerateKey()
	pubKeyECDSA, ok := privKey.Public().(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("error casting public key to ECDSA")
	}

	// strip off the 0x and the first 2 characters 04 which is always the EC prefix and is not required.
	publicKeyBytes := crypto.FromECDSAPub(pubKeyECDSA)[4:]
	var pubKey = make([]byte, 48)
	copy(pubKey[:], []byte(publicKeyBytes))

	addr := crypto.PubkeyToAddress(privKey.PublicKey)
	txOpts := bind.NewKeyedTransactor(privKey)
	txOpts.GasLimit = 6700000
	startingBalance, _ := new(big.Int).SetString("100000000000000000000000000000000000000", 10)
	genesis[addr] = core.GenesisAccount{Balance: startingBalance}
	backend := backends.NewSimulatedBackend(genesis, 210000000000)

	contractAddr, _, contract, err := DeployBridge(txOpts, backend)
	if err != nil {
		return nil, err
	}

	return &testAccount{addr, contract, contractAddr, backend, txOpts}, nil
}

func logs(test *testAccount) ([]types.Log, error) {
	query := ethereum.FilterQuery{
		Addresses: []common.Address{
			test.contractAddr,
		},
		FromBlock: big.NewInt(0),
	}

	logs, err := test.backend.FilterLogs(context.Background(), query)
	if err != nil {
		return nil, err
	}

	fmt.Println(logs)
	return logs, err
}

// deploys a new Ethereum contract, binding an instance of BridgeContract to it.
func DeployBridge(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *bridgecontract.Bridge, error) {
	bridgeabi, err := abi.JSON(strings.NewReader(bridgecontract.BridgeABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	fp, err := filepath.Abs("../build/Bridge.bin")
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	bin, err := ioutil.ReadFile(fp)
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	address, tx, contract, err := bind.DeployContract(auth, bridgeabi, bin, backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &bridgecontract.Bridge{BridgeCaller: bridgecontract.BridgeCaller{contract: contract}, BridgeTransactor: bridgecontract.BridgeTransactor{contract: contract}, BridgeFilterer: bridgecontract.BridgeFilterer{contract: contract}}, nil
}

func TestSetup(t *testing.T) {
	_, err := setup()
	if err != nil {
		t.Errorf("Can not deploy bridge contract: %v", err)
	}
}

// deploys the bridge contract and setups Chain type with name, url, client, from address, password, 
// chain ID, gas price and contract address
// func setup() (*Chain, error) {
// 	chain := new(Chain)
// 	chain.Name = "testnet"
// 	chain.Url = "https://rinkeby.infura.io"

// 	client, err := ethclient.Dial(chain.Url)
// 	if err != nil {
// 		return nil, err
// 	}

// 	chain.Client = client

// 	fp, err := filepath.Abs("../keystore")
// 	if err != nil {
// 		return nil, err
// 	}

// 	ks := keystore.NewKeyStore(fp, keystore.StandardScryptN, keystore.StandardScryptP)
// 	ksaccounts := ks.Accounts()
// 	if len(ksaccounts) == 0 {
// 		return nil, errors.New("no accounts in keystore")
// 	}

// 	a := common.HexToAddress("0x8f9b540b19520f8259115a90e4b4ffaeac642a30")
// 	chain.From = &a

// 	fp, err = filepath.Abs("../solidity/build/Bridge.bin")
// 	if err != nil {
// 		return nil, err
// 	}

// 	bin, err := ioutil.ReadFile(fp)
// 	if err != nil {
// 		return nil, err
// 	}

// 	chain.Password = "password"
// 	chain.Id = big.NewInt(0)
// 	chain.GasPrice = big.NewInt(10000000)

// 	address, err := Deploy(*chain, ks, bin, "Bridge")
// 	if err != nil {
// 		return nil, err
// 	}

// 	chain.Contract = &address

// 	return chain, nil
// }