package service

import (
	"sync/atomic"

	"github.com/drep-project/drep-chain/app"
	chainTypes "github.com/drep-project/drep-chain/chain/types"
	"github.com/drep-project/drep-chain/crypto"
	"github.com/drep-project/drep-chain/crypto/secp256k1"
	"github.com/drep-project/drep-chain/crypto/sha3"
	accountsComponent "github.com/drep-project/drep-chain/pkgs/accounts/component"
	accountTypes "github.com/drep-project/drep-chain/pkgs/accounts/types"
)

const (
	RPERMISSION = iota //read
	WPERMISSION        //write
)

const (
	LOCKED   = iota //locked
	UNLOCKED        //unlocked
)
//Wallet is used to manage private keys, build simple transactions and other functions.
type Wallet struct {
	cacheStore *accountsComponent.CacheStore

	chainId app.ChainIdType
	config  *accountTypes.Config

	isLock   int32
	password string
}

// NewWallet based in config
func NewWallet(config *accountTypes.Config, chainId app.ChainIdType) (*Wallet, error) {
	wallet := &Wallet{
		config:  config,
		chainId: chainId,
	}
	return wallet, nil
}

// Open wallet to use wallet
func (wallet *Wallet) Open(password string) error {
	if wallet.cacheStore != nil {
		return ErrClosedWallet
	}
	cryptedPassword := wallet.cryptoPassword(password)

	var store accountsComponent.KeyStore
	if wallet.config.Type == "dbstore" {
		store = accountsComponent.NewDbStore(wallet.config.KeyStoreDir)
	} else if wallet.config.Type == "memorystore" {
		store = accountsComponent.NewMemoryStore()
	}else{
		store = accountsComponent.NewFileStore(wallet.config.KeyStoreDir)
	}


	accountCacheStore, err := accountsComponent.NewCacheStore(store, cryptedPassword)
	if err != nil {
		return err
	}
	wallet.cacheStore = accountCacheStore
	wallet.unLock(password)
	keys, err := wallet.cacheStore.ExportKey(cryptedPassword)
	if err != nil {
		return err
	}
	if len(keys) == 0 {
		wallet.NewAccount()
	}
	return nil
}

// Close wallet to disable wallet
func (wallet *Wallet) Close() {
	wallet.Lock()
	wallet.cacheStore = nil
	wallet.password = ""
}

// NewAccount create new address
func (wallet *Wallet) NewAccount() (*chainTypes.Node, error) {
	if err := wallet.checkWallet(WPERMISSION); err != nil {
		return nil, err
	}

	newNode := chainTypes.NewNode(nil, wallet.chainId)
	wallet.cacheStore.StoreKey(newNode, wallet.password)
	return newNode, nil
}

// GetAccountByAddress query account according to address
func (wallet *Wallet) GetAccountByAddress(addr *crypto.CommonAddress) (*chainTypes.Node, error) {
	if err := wallet.checkWallet(RPERMISSION); err != nil {
		return nil, ErrClosedWallet
	}
	return wallet.cacheStore.GetKey(addr, wallet.password)
}

// GetAccountByAddress query account according to public key
func (wallet *Wallet) GetAccountByPubkey(pubkey *secp256k1.PublicKey) (*chainTypes.Node, error) {
	if err := wallet.checkWallet(RPERMISSION); err != nil {
		return nil, ErrClosedWallet
	}
	addr := crypto.PubKey2Address(pubkey)
	return wallet.GetAccountByAddress(&addr)
}

// ListAddress get all address in wallet
func (wallet *Wallet) ListAddress() ([]*crypto.CommonAddress, error) {
	if err := wallet.checkWallet(RPERMISSION); err != nil {
		return nil, ErrClosedWallet
	}
	nodes, err := wallet.cacheStore.ExportKey(wallet.password)
	if err != nil {
		return nil, err
	}
	addreses := []*crypto.CommonAddress{}
	for _, node := range nodes {
		addreses = append(addreses, node.Address)
	}
	return addreses, nil
}

// DumpPrivateKey query private key by address
func (wallet *Wallet) DumpPrivateKey(addr *crypto.CommonAddress) (*secp256k1.PrivateKey, error) {
	if err := wallet.checkWallet(WPERMISSION); err != nil {
		return nil, err
	}

	node, err := wallet.cacheStore.GetKey(addr, wallet.password)
	if err != nil {
		return nil, err
	}
	return node.PrivateKey, nil
}

// Sign sign a message using key in wallet
func (wallet *Wallet) Sign(addr *crypto.CommonAddress, msg []byte) ([]byte, error) {
	if len(msg) != 32 {
		return nil, ErrNotAHash
	}
	if err := wallet.checkWallet(WPERMISSION); err != nil {
		return nil, err
	}

	node, err := wallet.cacheStore.GetKey(addr, wallet.password)
	if err != nil {
		return nil, err
	}
	sig, err := secp256k1.SignCompact(node.PrivateKey, msg, true)
	if err != nil {
		return nil, err
	}
	return sig, nil
}

// IsLock query current lock state  0 is locked  1 is unlock
func (wallet *Wallet) IsLock() bool {
	return atomic.LoadInt32(&wallet.isLock) == LOCKED
}

// IsOpen query current wallet open state
func (wallet *Wallet) IsOpen() bool {
	return wallet.cacheStore != nil
}

// Lock wallet to disable private key
func (wallet *Wallet) Lock() error {
	atomic.StoreInt32(&wallet.isLock, LOCKED)
	wallet.cacheStore.ClearKeys()
	return nil
}

// UnLock wallet to enable private key
func (wallet *Wallet) UnLock(password string) error {
	if wallet.cacheStore == nil {
		return wallet.Open(password)
	} else {
		return wallet.unLock(password)
	}
}

func (wallet *Wallet) unLock(password string) error {
	atomic.StoreInt32(&wallet.isLock, UNLOCKED)
	wallet.password = wallet.cryptoPassword(password)
	return wallet.cacheStore.ReloadKeys(wallet.password)
}

func (wallet *Wallet) checkWallet(op int) error {
	if wallet.cacheStore == nil {
		return ErrClosedWallet
	}
	if op == WPERMISSION {
		if wallet.IsLock() {
			return ErrLockedWallet
		}
	}
	return nil
}

func (wallet *Wallet) cryptoPassword(password string) string {
	return string(sha3.Keccak256([]byte(password)))
}
