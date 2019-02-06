package bridgecontract

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"math/big"
	"path/filepath"
	"strings"
	"testing"

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
	contract     *Bridge
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
func DeployBridge(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *Bridge, error) {
	bridgeabi, err := abi.JSON(strings.NewReader(BridgeABI))
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
	return address, tx, &Bridge{BridgeCaller: BridgeCaller{contract: contract}, BridgeTransactor: BridgeTransactor{contract: contract}, BridgeFilterer: BridgeFilterer{contract: contract}}, nil
}

func TestSetup(t *testing.T) {
	_, err := setup()
	if err != nil {
		t.Errorf("Can not deploy bridge contract: %v", err)
	}
}

func TestDeposit(t *testing.T) {
	test, err := setup()
	if err != nil {
		t.Errorf("Can not deploy bridge contract: %v", err)
	}	

	test.txOpts.Value = big.NewInt(1000000000000000000)
	_, err = test.contract.Deposit(test.txOpts, test.addr, big.NewInt(1))
	if err != nil {
		t.Error("could not deposit into bridge")
	}

	_, err = logs(test)
	if err != nil {
		t.Fatalf("Unable to get logs of bridge contract: %v", err)
	}
}

func TestFundBridge(t *testing.T) {
	test, err := setup()
	if err != nil {
		t.Errorf("Can not deploy bridge contract: %v", err)
	}	

	test.txOpts.Value = big.NewInt(1000000000000000000)
	_, err = test.contract.FundBridge(test.txOpts)
	if err != nil {
		t.Error("could not fund bridge")
	}
}

// should try to withdraw more than deposited
func TestWithdrawToFailing(t *testing.T) {
	test, err := setup()
	if err != nil {
		t.Errorf("Can not deploy bridge contract: %v", err)
	}	

	test.txOpts.Value = big.NewInt(1000000000000000000)
	_, err = test.contract.WithdrawTo(test.txOpts, test.addr, big.NewInt(1), big.NewInt(2000000000000000000))
	if err == nil {
		t.Error("could withdraw more than balance")
	}
}

// try to call withdraw from an address that is not the authority
func TestWithdrawWrongAddress(t *testing.T) {
	test, err := setup()
	if err != nil {
		t.Errorf("Can not deploy bridge contract: %v", err)
	}	

	test.txOpts.Value = big.NewInt(1000000000000000000)
	txHashBytes, err := hex.DecodeString("09ed879028eb8a0b28763584639d0d609a42d4263b90ed3635323502aa9efde5")
	if err != nil {
		t.Error(err)
	}

	txHash := [32]byte{}
	copy(txHash[:], txHashBytes)
	_, err = test.contract.Withdraw(test.txOpts, common.HexToAddress("0x01"), big.NewInt(10), big.NewInt(1), txHash)
	if err == nil {
		t.Error("could withdraw from non authority account")
	}
}

func TestWithdraw(t *testing.T) {
	test, err := setup()
	if err != nil {
		t.Errorf("Can not deploy bridge contract: %v", err)
	}	

	test.txOpts.Value = big.NewInt(1000000000000000000)
	txHashBytes, err := hex.DecodeString("09ed879028eb8a0b28763584639d0d609a42d4263b90ed3635323502aa9efde5")
	if err != nil {
		t.Error(err)
	}

	txHash := [32]byte{}
	copy(txHash[:], txHashBytes)
	_, err = test.contract.Withdraw(test.txOpts, test.addr, big.NewInt(10), big.NewInt(1), txHash)
	if err != nil {
		t.Error("could not withdraw")
	}
}