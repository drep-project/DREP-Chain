package service

import (
	"math/big"

	blockmgr "github.com/drep-project/drep-chain/chain/service/blockmgr"

	chainTypes "github.com/drep-project/drep-chain/chain/types"
	"github.com/drep-project/drep-chain/common"
	"github.com/drep-project/drep-chain/crypto"
	"github.com/drep-project/drep-chain/crypto/secp256k1"
	"github.com/drep-project/drep-chain/database"
)

type AccountApi struct {
	Wallet          *Wallet
	accountService  *AccountService
	blockmgr    *blockmgr.BlockMgr
	databaseService *database.DatabaseService
}

func (accountapi *AccountApi) ListAddress() ([]*crypto.CommonAddress, error) {
	if !accountapi.Wallet.IsOpen() {
		return nil, ErrClosedWallet
	}
	return accountapi.Wallet.ListAddress()
}

// CreateAccount create a new account and return address
func (accountapi *AccountApi) CreateAccount() (*crypto.CommonAddress, error) {
	if !accountapi.Wallet.IsOpen() {
		return nil, ErrClosedWallet
	}
	newAaccount, err := accountapi.Wallet.NewAccount()
	if err != nil {
		return nil, err
	}
	return newAaccount.Address, nil
}

func (accountapi *AccountApi) CreateWallet(password string) error {
	err := accountapi.accountService.CreateWallet(password)
	if err != nil {
		return err
	}
	return accountapi.OpenWallet(password)
}

// Lock lock the wallet to protect private key
func (accountapi *AccountApi) LockWallet() error {
	if !accountapi.Wallet.IsOpen() {
		return ErrClosedWallet
	}
	if !accountapi.Wallet.IsLock() {
		return accountapi.Wallet.Lock()
	}
	return ErrLockedWallet
}

// UnLock unlock the wallet
func (accountapi *AccountApi) UnLockWallet(password string) error {
	if !accountapi.Wallet.IsOpen() {
		return ErrClosedWallet
	}
	if accountapi.Wallet.IsLock() {
		return accountapi.Wallet.UnLock(password)
	}
	return ErrAlreadyUnLocked
}

func (accountapi *AccountApi) OpenWallet(password string) error {
	return accountapi.Wallet.Open(password)
}

func (accountapi *AccountApi) CloseWallet() {
	accountapi.Wallet.Close()
}

func (accountapi *AccountApi) Transfer(from crypto.CommonAddress, to crypto.CommonAddress, amount, gasprice, gaslimit *common.Big, data common.Bytes) (string, error) {
	nonce := accountapi.blockmgr.GetTransactionCount(&from)
	tx := chainTypes.NewTransaction(to, (*big.Int)(amount), (*big.Int)(gasprice), (*big.Int)(gaslimit), nonce)
	sig, err := accountapi.Wallet.Sign(&from, tx.TxHash().Bytes())
	if err != nil {
		return "", err
	}
	tx.Sig = sig
	err = accountapi.blockmgr.SendTransaction(tx, true)
	if err != nil {
		return "", err
	}
	return tx.TxHash().String(), nil
}

func (accountapi *AccountApi) SetAlias(srcAddr crypto.CommonAddress, alias string, gasprice, gaslimit *common.Big) (string, error) {
	nonce := accountapi.blockmgr.GetTransactionCount(&srcAddr)
	t := chainTypes.NewAliasTransaction(alias, (*big.Int)(gasprice), (*big.Int)(gasprice), nonce)
	sig, err := accountapi.Wallet.Sign(&srcAddr, t.TxHash().Bytes())
	if err != nil {
		return "", err
	}
	t.Sig = sig
	err = accountapi.blockmgr.SendTransaction(t, true)
	if err != nil {
		return "", err
	}
	return t.TxHash().String(), nil
}

func (accountapi *AccountApi) Call(from crypto.CommonAddress, to crypto.CommonAddress, input common.Bytes, amount, gasprice, gaslimit *common.Big) (string, error) {
	nonce := accountapi.blockmgr.GetTransactionCount(&from)
	t := chainTypes.NewCallContractTransaction(to, input, (*big.Int)(amount), (*big.Int)(gasprice), (*big.Int)(gaslimit), nonce)
	sig, err := accountapi.Wallet.Sign(&from, t.TxHash().Bytes())
	if err != nil {
		return "", err
	}
	t.Sig = sig
	accountapi.blockmgr.SendTransaction(t, true)
	return t.TxHash().String(), nil
}

func (accountapi *AccountApi) CreateCode(from crypto.CommonAddress, byteCode common.Bytes, amount, gasprice, gaslimit *common.Big) (string, error) {
	nonce := accountapi.blockmgr.GetTransactionCount(&from)
	t := chainTypes.NewContractTransaction(byteCode, (*big.Int)(gasprice), (*big.Int)(gaslimit), nonce)
	sig, err := accountapi.Wallet.Sign(&from, t.TxHash().Bytes())
	if err != nil {
		return "", err
	}
	t.Sig = sig
	accountapi.blockmgr.SendTransaction(t, true)
	return t.TxHash().String(), nil
}

// DumpPrikey dumpPrivate
func (accountapi *AccountApi) DumpPrivkey(address *crypto.CommonAddress) (*secp256k1.PrivateKey, error) {
	if !accountapi.Wallet.IsOpen() {
		return nil, ErrClosedWallet
	}
	if accountapi.Wallet.IsLock() {
		return nil, ErrLockedWallet
	}

	node, err := accountapi.Wallet.GetAccountByAddress(address)
	if err != nil {
		return nil, err
	}
	return node.PrivateKey, nil
}

func (accountapi *AccountApi) Sign(address crypto.CommonAddress, hash []byte) ([]byte, error) {
	sig, err := accountapi.Wallet.Sign(&address, hash)
	if err != nil {
		return nil, err
	}
	return sig, nil
}

func (accountapi *AccountApi) GasPrice() *big.Int {
	return chainTypes.DefaultGasPrice
}

func (accountapi *AccountApi) GetCode(addr crypto.CommonAddress) []byte {
	return accountapi.databaseService.GetByteCode(&addr)
}
