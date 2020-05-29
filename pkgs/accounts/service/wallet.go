package service

import (
	"github.com/drep-project/DREP-Chain/common/fileutil"
	"github.com/drep-project/DREP-Chain/crypto"
	"github.com/drep-project/DREP-Chain/crypto/secp256k1"
	accountsComponent "github.com/drep-project/DREP-Chain/pkgs/accounts/component"
	accountTypes "github.com/drep-project/DREP-Chain/pkgs/accounts/types"
	"github.com/drep-project/DREP-Chain/types"
	"github.com/pkg/errors"
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

	chainId types.ChainIdType
	config  *accountTypes.Config

	isLock   int32
	password string
}

// NewWallet based in config
func NewWallet(config *accountTypes.Config, chainId types.ChainIdType) (*Wallet, error) {
	wallet := &Wallet{
		config:  config,
		chainId: chainId,
	}
	wallet.password = config.Password
	return wallet, nil
}

// OpenWallet wallet to use wallet
func (wallet *Wallet) OpenWallet(password string) error {
	if wallet.cacheStore != nil {
		return ErrClosedWallet
	}

	//cryptedPassword := wallet.cryptoPassword(wallet.password)
	wallet.password = password

	var store accountsComponent.KeyStore
	if wallet.config.Type == "dbstore" {
		store = accountsComponent.NewDbStore(wallet.config.KeyStoreDir)
	} else if wallet.config.Type == "memorystore" {
		store = accountsComponent.NewMemoryStore()
	} else {
		store = accountsComponent.NewFileStore(wallet.config.KeyStoreDir)
	}

	accountCacheStore, err := accountsComponent.NewCacheStore(store, password)
	if err != nil {
		log.WithField("err", err).Info("cache account err")
		return nil
	}
	wallet.cacheStore = accountCacheStore
	//wallet.unLock(wallet.config.Password)
	//keys, err := wallet.cacheStore.ExportKey(cryptedPassword)
	//if err != nil {
	//	log.WithField("err", err).Info("export key err")
	//}
	//if len(keys) == 0 {
	//	wallet.NewAccount()
	//}

	return nil
}

func (wallet *Wallet) UnlockAccount(addr *crypto.CommonAddress) error {
	return wallet.cacheStore.LoadKeys(addr, wallet.password)
}

// Close wallet to disable wallet
func (wallet *Wallet) Close() {
	//wallet.Lock()
	wallet.cacheStore = nil
	wallet.password = ""
}

// NewAccount create new address
func (wallet *Wallet) NewAccount() (*types.Node, error) {
	if err := wallet.checkWallet(WPERMISSION); err != nil {
		return nil, err
	}

	newNode := types.NewNode(nil, wallet.chainId)
	wallet.cacheStore.StoreKey(newNode, wallet.password)
	return newNode, nil
}

// GetAccountByAddress query account according to address
func (wallet *Wallet) GetAccountByAddress(addr *crypto.CommonAddress) (*types.Node, error) {
	if err := wallet.checkWallet(RPERMISSION); err != nil {
		return nil, ErrClosedWallet
	}

	return wallet.cacheStore.GetKey(addr, wallet.password)
}

// GetAccountByAddress query account according to public key
func (wallet *Wallet) GetAccountByPubkey(pubkey *secp256k1.PublicKey) (*types.Node, error) {
	if err := wallet.checkWallet(RPERMISSION); err != nil {
		return nil, ErrClosedWallet
	}
	addr := crypto.PubkeyToAddress(pubkey)
	if err := wallet.unLock(&addr); err != nil {
		return nil, ErrAccountExist
	}

	return wallet.GetAccountByAddress(&addr)
}

// ListAddress get all address in wallet
func (wallet *Wallet) ListAddress() ([]string, error) {
	if err := wallet.checkWallet(RPERMISSION); err != nil {
		return nil, ErrClosedWallet
	}

	addrs, err := wallet.cacheStore.ListAddr(wallet.password)
	if err != nil {
		return nil, err
	}

	return addrs, nil
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
	//return atomic.LoadInt32(&wallet.isLock) == LOCKED
	return false
}

// IsOpen query current wallet open state
func (wallet *Wallet) IsOpen() bool {
	return wallet.cacheStore != nil
}

// Lock wallet to disable private key
func (wallet *Wallet) Lock(addr *crypto.CommonAddress) error {
	//atomic.StoreInt32(&wallet.isLock, LOCKED)
	wallet.cacheStore.ClearKey(addr)
	return nil
}

// UnLock wallet to enable private key
func (wallet *Wallet) UnLock(addr *crypto.CommonAddress) error {
	if wallet.cacheStore == nil {
		err := wallet.OpenWallet(wallet.password)
		if err != nil {
			return err
		}
	}

	return wallet.unLock(addr)
}

func (wallet *Wallet) unLock(addr *crypto.CommonAddress) error {
	//atomic.StoreInt32(&wallet.isLock, UNLOCKED)
	return wallet.cacheStore.LoadKeys(addr, wallet.password)
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

//func (wallet *Wallet) cryptoPassword(password string) string {
//	return string(sha3.Keccak256([]byte(password)))
//}

func (wallet *Wallet) ImportPrivKey(key *secp256k1.PrivateKey) (*types.Node, error) {
	if err := wallet.checkWallet(WPERMISSION); err != nil {
		return nil, err
	}
	addr := crypto.PubkeyToAddress(key.PubKey())
	node := &types.Node{
		Address:    &addr,
		PrivateKey: key,
		ChainId:    wallet.chainId,
	}
	_, err := wallet.cacheStore.GetKey(&addr, wallet.password)
	if err == nil {
		return nil, errors.Wrap(ErrExistKey, addr.String())
	}

	err = wallet.cacheStore.LoadKeys(&addr, wallet.password)
	if err == nil {
		return nil, errors.Wrap(ErrExistKey, addr.String())
	}

	err = wallet.cacheStore.StoreKey(node, wallet.password)
	if err != nil {
		return nil, err
	}
	return node, nil
}

func (wallet *Wallet) ImportKeyStore(path, password string) ([]*crypto.CommonAddress, error) {
	if err := wallet.checkWallet(WPERMISSION); err != nil {
		return nil, err
	}
	if !fileutil.IsDirExists(path) {
		return nil, errors.Wrap(ErrMissingKeystore, path)
	}

	newWallet, err := NewWallet(&accountTypes.Config{
		Enable:      true,
		Type:        "keystore",
		KeyStoreDir: path,
	}, wallet.chainId)
	if err != nil {
		return nil, err
	}
	err = newWallet.OpenWallet(password)
	if err != nil {
		return nil, err
	}
	nodes, err := newWallet.cacheStore.ExportKey(password)
	if err != nil {
		return nil, err
	}
	addrs := []*crypto.CommonAddress{}
	for _, node := range nodes {
		_, err := wallet.cacheStore.GetKey(node.Address, wallet.password)
		if err == nil {
			log.WithField("addr", node.Address.String()).Info("privkey exist")
			continue
		}
		err = wallet.cacheStore.StoreKey(node, wallet.password)
		if err != nil {
			return addrs, err
		}
		addrs = append(addrs, node.Address)
	}
	return addrs, nil
}
