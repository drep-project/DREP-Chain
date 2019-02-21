package service

import (
	accountCommponent "github.com/drep-project/drep-chain/accounts/component"
	"github.com/drep-project/drep-chain/crypto"
    "github.com/drep-project/drep-chain/common"
    "github.com/drep-project/drep-chain/chain/service"
    "github.com/drep-project/drep-chain/crypto/secp256k1"
    "github.com/pkg/errors"
    "math/big"
    "strconv"
    "encoding/json"
)

var chainService = service.ChainService{}

type AccountApi struct {
	Wallet *accountCommponent.Wallet
}

func (accountapi *AccountApi) AddressList() ([]*crypto.CommonAddress, error) {
	if !accountapi.Wallet.IsOpen() {
		return nil, errors.New("wallet is not open")
	}
	return accountapi.Wallet.ListAddress()
}

// CreateAccount create a new account and return address
func (accountapi *AccountApi) CreateAccount() (*crypto.CommonAddress, error) {
	if !accountapi.Wallet.IsOpen() {
		return nil, errors.New("wallet is not open")
	}
	newAaccount, err := accountapi.Wallet.NewAccount()
	if err != nil {
		return nil, err
	}
	return newAaccount.Address, nil
}

// DumpPrikey dumpPrivate
func (accountapi *AccountApi) DumpPrikey(address *crypto.CommonAddress) (*secp256k1.PrivateKey, error) {
	if !accountapi.Wallet.IsOpen() {
		return nil, errors.New("wallet is not open")
	}
	if accountapi.Wallet.IsLock() {
		return nil, errors.New("wallet has locked")
	}

	node, err := accountapi.Wallet.GetAccountByAddress(address)
	if err != nil {
		return nil, err
	}
	return node.PrivateKey, nil
}

// Lock lock the wallet to protect private key
func (accountapi *AccountApi) Lock() error {
	if !accountapi.Wallet.IsOpen() {
		return errors.New("wallet is not open")
	}
	if !accountapi.Wallet.IsLock() {
		return accountapi.Wallet.Lock()
	}
	return errors.New("wallet is already locked")
}

// UnLock unlock the wallet
func (accountapi *AccountApi) UnLock(password string) error {
	if !accountapi.Wallet.IsOpen() {
		return errors.New("wallet is not open")
	}
	if accountapi.Wallet.IsLock() {
		return accountapi.Wallet.UnLock(password)
	}
	return errors.New("wallet is already unlock")
}

func (accountapi *AccountApi) Open(password string) error {
	return accountapi.Wallet.Open(password)
}

func (accountapi *AccountApi) Close() {
	accountapi.Wallet.Close()
}

func (accountapi *AccountApi) SendTransaction(from, to, chainId, amount string) (string, error) {
    if from == "" || to == "" || chainId == "" || amount == "" {
        return "", errors.New("param can not be nil")
    }

    fromBytes := []byte(from)
    fromPubK := &secp256k1.PublicKey{}
    json.Unmarshal(fromBytes, fromPubK)

    destAddr := crypto.String2Address(to)
    destChain := common.String2ChainId(chainId)
    x, err := strconv.ParseInt(amount, 10, 64)
    if err != nil {
        return "", err
    }
    a := big.NewInt(x)
	t := chainService.GenerateBalanceTransaction(fromPubK, destAddr, destChain, a)
    if chainService.SendTransaction(t) != nil {
        return "", errors.New("Offline")
    } else {
        return t.TxId()
    }
}

func (accountapi *AccountApi) Call(from, to, chainId, input, amount string, readOnly bool)  (string, error){
    if from == "" || to == "" || chainId == "" || input == "" || amount == "" {
        return "", errors.New("param can not be nil")
    }
    fromBytes := []byte(from)
    fromPubK := &secp256k1.PublicKey{}
    json.Unmarshal(fromBytes, fromPubK)

    destAddr := crypto.String2Address(to)
    destChain := common.String2ChainId(chainId)

    inputBytes := []byte(input)

    x, err := strconv.ParseInt(amount, 10, 64)
    if err != nil {
        return "", err
    }
    value := big.NewInt(x)

    t := chainService.GenerateCallContractTransaction(fromPubK, destAddr, destChain, inputBytes, value, readOnly)
    if chainService.SendTransaction(t) != nil {
        return "", errors.New("Offline")
    } else {
        return t.TxId()
    }
}
