package home

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

	//ethereum "github.com/ethereum/go-ethereum"
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
	contract     *Home
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

	contractAddr, _, contract, err := DeployHome(txOpts, backend)
	if err != nil {
		return nil, err
	}

	return &testAccount{addr, contract, contractAddr, backend, txOpts}, nil
}

// deploys a new Ethereum contract, binding an instance of BridgeContract to it.
func DeployHome(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *Home, error) {
	bridgeabi, err := abi.JSON(strings.NewReader(HomeABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	fp, err := filepath.Abs("./build/Home.bin")
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
	fmt.Printf("Deployed contract to: %x", address)
	return address, tx, &Home{HomeCaller: HomeCaller{contract: contract}, HomeTransactor: HomeTransactor{contract: contract}, HomeFilterer: HomeFilterer{contract: contract}}, nil
}

func TestSetup(t *testing.T) {
	_, err := setup()
	if err != nil {
		t.Errorf("Can not deploy home contract: %v", err)
	}
}

func TestSetBridge(t *testing.T) {
	test, err := setup()
	if err != nil {
		t.Fatalf("Can not deploy home contract: %v", err)
	}

	bridgeAddr := common.HexToAddress("E4732e8D48810e49EEf741cc31e997513Fa999c5")

	// Set address
	tx, err := test.contract.SetBridge(test.txOpts, bridgeAddr)
	if err != nil {
		t.Error("could not set bridge address")
	}

	test.backend.Commit()

	receipt, err := test.backend.TransactionReceipt(context.Background(), tx.Hash())
	if err != nil {
		t.Fatalf("Unable to get receipt: %v", err)
	}

	if receipt.Status == 0 {
		t.Fatal("SetBridge failed")
	}

	// Check address
	addr, err := test.contract.HomeCaller.Bridge(&bind.CallOpts{From: bridgeAddr})
	if err != nil {
		t.Fatalf("could not get bridge address. %v", err)
	}

	if addr != bridgeAddr {
		t.Fatalf("Bridge address not correctly set. got=%x expected=%x", addr, bridgeAddr)
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
	tx, err := test.contract.Withdraw(test.txOpts, test.addr, big.NewInt(10), big.NewInt(1), txHash)
	if err != nil {
		t.Error("could not send tx")
	}

	test.backend.Commit()

	receipt, err := test.backend.TransactionReceipt(context.Background(), tx.Hash())
	if err != nil {
		t.Fatalf("Unable to get receipt: %v", err)
	}

	if receipt.Status == 0 {
		t.Fatal("No withdraw logs")
	}
}