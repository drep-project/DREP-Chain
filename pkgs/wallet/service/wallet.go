package service

import (
	"errors"
	"github.com/drep-project/drep-chain/app"
	"github.com/drep-project/drep-chain/crypto/secp256k1"
	"github.com/drep-project/drep-chain/crypto/sha3"
	accountsComponent "github.com/drep-project/drep-chain/pkgs/wallet/component"
	accountTypes "github.com/drep-project/drep-chain/pkgs/wallet/types"
	"sync/atomic"
)

const (
	RPERMISSION = iota //read
	WPERMISSION        //write
)

const (
	LOCKED   = iota //locked
	UNLOCKED        //unlocked
)

type Wallet struct {
	cacheStore *accountsComponent.CacheStore

	chainId app.ChainIdType
	config  *accountTypes.Config

	isLock   int32
	password string
}

func NewWallet(config *accountTypes.Config, chainId app.ChainIdType) (*Wallet, error) {
	wallet := &Wallet{
		config:  config,
		chainId: chainId,
	}
	return wallet, nil
}

func CreateWallet(config *accountTypes.Config, chainId app.ChainIdType, password string) (*Wallet, error) {
	wallet := &Wallet{
		config:  config,
		chainId: chainId,
	}
	err := wallet.Open(password)
	if err != nil {
		return nil, err
	}
	return wallet, nil
}

func (wallet *Wallet) Open(password string) error {
	if wallet.cacheStore != nil {
		return errors.New("wallet is already open")
	}
	cryptedPassword := wallet.cryptoPassword(password)
	accountCacheStore, err := accountsComponent.NewCacheStore(wallet.config.KeyStoreDir, cryptedPassword)
	if err != nil {
		return err
	}
	wallet.cacheStore = accountCacheStore
	wallet.unLock(password)
	return nil
}

func (wallet *Wallet) Close() {
	wallet.Lock()
	wallet.cacheStore = nil
	wallet.password = ""
}

func (wallet *Wallet) ListKeys() ([]*secp256k1.PublicKey, error) {
	if err := wallet.checkWallet(RPERMISSION); err != nil {
		return nil, errors.New("wallet is not open")
	}
	keys, err := wallet.cacheStore.ExportKey(wallet.password)
	if err != nil {
		return nil, err
	}
	pKeys := []*secp256k1.PublicKey{}
	for _, key := range keys {
		pKeys = append(pKeys, key.Pubkey)
	}
	return pKeys, nil
}

func (wallet *Wallet) DumpPrivateKey(pubkey *secp256k1.PublicKey) (*secp256k1.PrivateKey, error) {
	if err := wallet.checkWallet(WPERMISSION); err != nil {
		return nil, err
	}

	key, err := wallet.cacheStore.GetKey(pubkey, wallet.password)
	if err != nil {
		return nil, err
	}
	return key.PrivKey, nil
}

// 0 is locked  1 is unlock
func (wallet *Wallet) IsLock() bool {
	return atomic.LoadInt32(&wallet.isLock) == LOCKED
}

func (wallet *Wallet) IsOpen() bool {
	return wallet.cacheStore != nil
}

func (wallet *Wallet) Lock() error {
	atomic.StoreInt32(&wallet.isLock, LOCKED)
	wallet.cacheStore.ClearKeys()
	return nil
}

func (wallet *Wallet) UnLock(password string) error {
	if wallet.cacheStore == nil {
		wallet.Open(password)
	} else {
		wallet.unLock(password)
	}
	return nil
}

func (wallet *Wallet) unLock(password string) error {
	atomic.StoreInt32(&wallet.isLock, UNLOCKED)
	wallet.password = wallet.cryptoPassword(password)
	wallet.cacheStore.ReloadKeys(wallet.password)
	return nil
}

func (wallet *Wallet) checkWallet(op int) error {
	if wallet.cacheStore == nil {
		return errors.New("wallet is not open")
	}
	if op == WPERMISSION {
		if wallet.IsLock() {
			return errors.New("wallet is locked")
		}
	}
	return nil
}

func (wallet *Wallet) cryptoPassword(password string) string {
	return string(sha3.Hash256([]byte(password + "drep")))
}
