package service

import (
	"bytes"
	"github.com/drep-project/drep-chain/app"
	"github.com/drep-project/drep-chain/chain/types"
	"github.com/drep-project/drep-chain/crypto/sha3"
	accountTypes "github.com/drep-project/drep-chain/pkgs/accounts/types"
	"testing"
)

var (
	testConfig = &accountTypes.Config{
		Enable:      true,
		Type:        "memorystore",
		KeyStoreDir: "keystore",
	}
)

func Test_WalletOpend(t *testing.T) {
	password := "password"
	wallet, err := NewWallet(testConfig, app.ChainIdType{})
	if err != nil {
		t.Error(err)
	}
	wallet.Open(password)
	if wallet.IsOpen() == false {
		t.Error("expected wallet is open but got close")
	}
}

func Test_WalletClosed(t *testing.T) {
	password := "password"
	wallet, err := NewWallet(testConfig, app.ChainIdType{})
	if err != nil {
		t.Error(err)
	}
	wallet.Open(password)
	if wallet.IsOpen() == false {
		t.Error("expected wallet is open but got close")
	}
	wallet.Close()
	if wallet.IsOpen() == true {
		t.Error("expected wallet is close but got open")
	}

	_, err = wallet.ListAddress()
	if err == nil {
		t.Error("expect ListAddress fail but success")
	}

	_, err = wallet.DumpPrivateKey(nil)
	if err == nil {
		t.Error("expect DumpPrivateKey fail but success")
	}
}

func Test_WalletLock(t *testing.T) {
	password := "password"
	wallet, err := NewWallet(testConfig, app.ChainIdType{})
	if err != nil {
		t.Error(err)
	}
	wallet.Open(password)
	node, err := wallet.NewAccount()
	if err != nil {
		t.Error(err)
	}
	wallet.Lock()
	if wallet.IsLock() == false {
		t.Error("expected wallet is lock but got close")
	}
	_, err = wallet.ListAddress()
	if err != nil {
		t.Error("expect ListAddress success but fail")
	}

	_, err = wallet.DumpPrivateKey(node.Address)
	if err == nil {
		t.Error("expect ListAddress fail but success")
	}
}

func Test_WalletUnLock(t *testing.T) {
	password := "password"
	wallet, err := NewWallet(testConfig, app.ChainIdType{})
	if err != nil {
		t.Error(err)
	}
	wallet.Open(password)

	wallet.Lock()
	if wallet.IsLock() == false {
		t.Error("expected wallet is lock but got unlock")
	}

	wallet.UnLock(password)
	if wallet.IsLock() == true {
		t.Error("expected wallet is unlock but got lock")
	}
}

func Test_NewAccountAndListAddress(t *testing.T) {
	password := "password"
	wallet, err := NewWallet(testConfig, app.ChainIdType{})
	if err != nil {
		t.Error(err)
	}
	wallet.Open(password)

	count := 10
	nodes := make([]*types.Node, count)
	for i := 0; i < count; i++ {
		node, err := wallet.NewAccount()
		if err != nil {
			t.Error(err)
		}
		nodes[i] = node
	}

	addresses, err := wallet.ListAddress()
	if err != nil {
		t.Error(err)
	}
	for _, node := range nodes {
		isFind := false
		for _, addr := range addresses {
			if node.Address == addr {
				isFind = true
				break
			}
		}
		if !isFind {
			t.Error("cannot find address of new node :", node.Address, " ", addresses)
		}
	}
}

func Test_NewAccountAndDumpPrivateKey(t *testing.T) {
	password := "password"
	wallet, err := NewWallet(testConfig, app.ChainIdType{})
	if err != nil {
		t.Error(err)
	}
	wallet.Open(password)

	count := 10
	nodes := make([]*types.Node, count)
	for i := 0; i < count; i++ {
		node, err := wallet.NewAccount()
		if err != nil {
			t.Error(err)
		}
		nodes[i] = node
	}
	for _, node := range nodes {
		privkey, err := wallet.DumpPrivateKey(node.Address)
		if err != nil {
			t.Error(err)
		}
		if !bytes.Equal(privkey.Serialize(), node.PrivateKey.Serialize()) {
			t.Error("dumprivate key fail expect:", node.PrivateKey, " but got ", privkey)
		}
	}
}

func Test_Sign(t *testing.T) {
	password := "password"
	wallet, err := NewWallet(testConfig, app.ChainIdType{})
	if err != nil {
		t.Error(err)
	}
	wallet.Open(password)
	node, err := wallet.NewAccount()
	if err != nil {
		t.Error(err)
	}

	msg := sha3.Keccak256([]byte("helloworld"))
	_, err = wallet.Sign(node.Address, msg)
	if err != nil {
		t.Error(err)
	}
}
