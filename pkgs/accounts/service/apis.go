package service

import (
	"math/big"
	"github.com/pkg/errors"
	"github.com/drep-project/drep-chain/crypto"
    "github.com/drep-project/drep-chain/crypto/secp256k1"
   chainService "github.com/drep-project/drep-chain/chain/service"
    "github.com/drep-project/drep-chain/database"
)


type AccountApi struct {
	Wallet *Wallet
	accountService *AccountService
	chainService *chainService.ChainService
	databaseService *database.DatabaseService
}

func (accountapi *AccountApi) AddressList() ([]*crypto.CommonAddress, error) {
	if !accountapi.Wallet.IsOpen() {
		return nil, errors.New("wallet is not open")
	}
	return accountapi.Wallet.ListAddress()
}

func (accountapi *AccountApi) Create(password string) error {
	err := accountapi.accountService.CreateWallet(password)
	if err != nil {
		return err
	}
	return accountapi.Open(password)
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
func (accountapi *AccountApi) DumpPrivkey(address *crypto.CommonAddress) (*secp256k1.PrivateKey, error) {
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

func (accountapi *AccountApi) SendTransaction(from *secp256k1.PublicKey, to crypto.CommonAddress, amount *big.Int) (string, error) {
	t := accountapi.chainService.GenerateBalanceTransaction(from, to, amount)
    accountapi.chainService.SendTransaction(t)
	hash, err := t.TxHash()
    return "0x" + string(hash), err
}

func (accountapi *AccountApi) Call(from *secp256k1.PublicKey, to crypto.CommonAddress, input []byte, amount *big.Int, readOnly bool)  (string, error){
    t := accountapi.chainService.GenerateCallContractTransaction(from, to, input, amount, readOnly)
    accountapi.chainService.SendTransaction(t)
    return t.TxId()
}

func (accountapi *AccountApi) CreateCode(from *secp256k1.PublicKey, to crypto.CommonAddress, byteCode []byte) (string, error){
    t := accountapi.chainService.GenerateCreateContractTransaction(from, to, byteCode)
    accountapi.chainService.SendTransaction(t)
    return t.TxId()
}

func (accountapi *AccountApi) Sign(prv *secp256k1.PrivateKey, msg string) ([]byte, error) {
    bytes := []byte(msg)
    return crypto.Sign(bytes, prv)
}

func (accountapi *AccountApi) GasPrice() *big.Int {
    return chainService.DefaultGasPrice
}

func (accountapi *AccountApi) GetCode(addr crypto.CommonAddress) []byte {
    return accountapi.databaseService.GetByteCode(addr, false)
}