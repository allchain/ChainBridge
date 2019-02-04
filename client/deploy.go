package client

import (
	"fmt"
	//"encoding/json"
	"context"
	"errors"
	"math/big"

	"github.com/ChainSafeSystems/ChainBridge/logger"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/common"
	//"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/core/types"
)

func Deploy(chain Chain, ks *keystore.KeyStore, bin []byte, contract string) (common.Address, error) {
	nonce, err := chain.Client.PendingNonceAt(context.Background(), *chain.From)
	if err != nil {
		return *new(common.Address), err
	}

	tx := types.NewContractCreation(nonce, big.NewInt(0), uint64(4600000), chain.GasPrice, bin)
	fmt.Println(ks.Accounts()[0])
	txSigned, err := ks.SignTxWithPassphrase(ks.Accounts()[0], chain.Password, tx, chain.Id)
	if err != nil {
		return *new(common.Address), errors.New(fmt.Sprintf("could not sign tx: %s", err))
	}

	txHash := txSigned.Hash()
	logger.Info(fmt.Sprintf("attempting to send tx %s to from account %s to deploy contract %s.sol", txHash.Hex(), chain.From.Hex(), contract))

	err = chain.Client.SendTransaction(context.Background(), txSigned)
	if err != nil {
		return *new(common.Address), errors.New(fmt.Sprintf("could not send tx %s: %s", txHash.Hex(), err))
	}

	WaitOnPending(chain.Client, txHash)
	receipt, err := chain.Client.TransactionReceipt(context.Background(), txHash)
	if err != nil {
		receipt = waitOnPendingKovan(chain.Client, txHash)
	}

	if receipt.Status == 0 {
		// todo: sometimes status == 0 but the contract is successfully deployed
		logger.Warn(fmt.Sprintf("tx receipt status = 0 for deployment %s.sol", contract))
	}

	contractAddr := receipt.ContractAddress
	logger.Info(fmt.Sprintf("contract deployed at address %s", contractAddr.Hex()))
	logger.Info(fmt.Sprintf("gas used to deploy contract %s.sol: %d", contract, receipt.GasUsed))
	return contractAddr, nil
}

func WaitOnPending(client *ethclient.Client, txHash common.Hash) (*types.Transaction) {
	for {
		tx, pending, _ := client.TransactionByHash(context.Background(), txHash)
		if !pending { 
			return tx 
		}
	}
}

func waitOnPendingKovan(client *ethclient.Client, txHash common.Hash) (*types.Receipt) {
	for {
		receipt, err := client.TransactionReceipt(context.Background(), txHash)
		if err == nil {
			return receipt
		}
	}
}