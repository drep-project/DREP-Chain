package service

import (
	"bytes"
	"encoding/hex"
	"github.com/drep-project/DREP-Chain/app"
	"github.com/drep-project/DREP-Chain/crypto"
	"github.com/drep-project/DREP-Chain/crypto/secp256k1"
	"github.com/drep-project/DREP-Chain/crypto/sha3"
	accountTypes "github.com/drep-project/DREP-Chain/pkgs/accounts/types"
	"github.com/drep-project/DREP-Chain/types"
	"os"
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

func TestWallet_ImportPrivKey(t *testing.T) {
	testData := map[string]string{
		"0xc9a53673960f8dea4aecadca77bf363c2b21b4ce": "e2eee3e0791242f1eabeb33e6c2f474a6932b04153ef9f14c0ffad205c88be2e",
		"0x8faf3fa18d0a53d3ef02d3bdcd35d43c71439aec": "f8eb7b53e9eb00465cde6309d7aa2942dc528847deab7dc251b1b10d420fa994",
		"0x2670e46875fab57293aaeb442f3a901be06a5998": "52fc9a3eb26b3b3d4545d8b1e85c166226deaf69664d07b15e3f0254a7723bed",
	}
	password := "password"
	wallet, err := NewWallet(testConfig, app.ChainIdType{})
	if err != nil {
		t.Error(err)
	}
	wallet.Open(password)

	for _, pri := range testData {
		privBytes, _ := hex.DecodeString(pri)
		priv, _ := secp256k1.PrivKeyFromScalar(privBytes)
		wallet.ImportPrivKey(priv)
	}

	for addrStr, priStr := range testData {
		addr := crypto.String2Address(addrStr)
		pri, err := wallet.DumpPrivateKey(&addr)
		if err != nil {
			t.Error(err)
		}
		loadPriStr := hex.EncodeToString(pri.Serialize())
		if loadPriStr != priStr {
			t.Errorf("privkey is wrong while reload from wallet, savekey:%s, loadkey: %s", priStr, loadPriStr)
		}
	}
}

func Test_ImportKeyStore(t *testing.T) {
	testData := map[string]string{
		"0xc9a53673960f8dea4aecadca77bf363c2b21b4ce": "e2eee3e0791242f1eabeb33e6c2f474a6932b04153ef9f14c0ffad205c88be2e",
		"0x8faf3fa18d0a53d3ef02d3bdcd35d43c71439aec": "f8eb7b53e9eb00465cde6309d7aa2942dc528847deab7dc251b1b10d420fa994",
		"0x2670e46875fab57293aaeb442f3a901be06a5998": "52fc9a3eb26b3b3d4545d8b1e85c166226deaf69664d07b15e3f0254a7723bed",
	}
	oldWalletPath := "test_import_key_store_oldwallet"
	newWalletPath := "test_import_key_store_newwallet"
	oldPassword := "123"
	newPassword := "ffg"
	defer func() {
		os.RemoveAll(oldWalletPath)
		os.RemoveAll(newWalletPath)
	}()

	makeTestWallet := func() {
		wallet, err := getWallet(oldWalletPath, oldPassword)
		if err != nil {
			t.Error(err)
		}
		for _, pri := range testData {
			privBytes, _ := hex.DecodeString(pri)
			priv, _ := secp256k1.PrivKeyFromScalar(privBytes)
			wallet.ImportPrivKey(priv)
		}
	}

	makeTestWallet()
	newWallet, err := getWallet(newWalletPath, newPassword)
	if err != nil {
		t.Error(err)
	}
	newWallet.ImportKeyStore(oldWalletPath, oldPassword)

	for addrStr, priStr := range testData {
		addr := crypto.String2Address(addrStr)
		pri, err := newWallet.DumpPrivateKey(&addr)
		if err != nil {
			t.Error(err)
		}
		loadPriStr := hex.EncodeToString(pri.Serialize())
		if loadPriStr != priStr {
			t.Errorf("privkey is wrong while reload from wallet, savekey:%s, loadkey: %s", priStr, loadPriStr)
		}
	}
}

func getWallet(path, password string) (*Wallet, error) {
	testConfig := &accountTypes.Config{
		Enable:      true,
		Type:        "keystore",
		KeyStoreDir: path,
	}
	wallet, err := NewWallet(testConfig, app.ChainIdType{})
	if err != nil {
		return nil, err
	}
	wallet.Open(password)
	return wallet, nil
}
