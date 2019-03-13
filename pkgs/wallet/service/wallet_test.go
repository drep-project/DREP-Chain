package service

import (
	"fmt"
	"github.com/drep-project/drep-chain/app"
	"testing"
	"github.com/drep-project/drep-chain/crypto/secp256k1"
	"github.com/drep-project/drep-chain/crypto/sha3"
	accountTypes "github.com/drep-project/drep-chain/pkgs/wallet/types"
)

func TestWallet(t *testing.T) {
	password := string(sha3.Hash256([]byte("AAAAAAAAAAAAAAAA")))
	rootChain := app.ChainIdType{}
	newNode := accountTypes.NewNode(nil, rootChain)
	fmt.Println(newNode)

	accountConfig := &accountTypes.Config{
		KeyStoreDir: "TestWallet",
	}
	wallet, err := NewWallet(accountConfig, rootChain)
	wallet.chainId = rootChain
	if err != nil {
		dlog.Fatal("NewWallet error")
	}

	err = wallet.Open(password)
	if err != nil {
		dlog.Fatal("open wallet error")
	}

	nodes := []*accountTypes.Node{}
	for i := 0; i < 10; i++ {
		node, err := wallet.NewAccount()
		pk := node.PrivateKey.PubKey()
		isOnCurve := secp256k1.S256().IsOnCurve(pk.X, pk.Y)
		if !isOnCurve {
			dlog.Fatal("error public key")
		}
		if err != nil {
			dlog.Fatal("open wallet error")
		}
		nodes = append(nodes, node)
	}

	wallet.Lock()
	_, err = wallet.NewAccount()
	if err == nil {
		dlog.Fatal("Lock not effect")
	}

	wallet.UnLock(password)
	_, err = wallet.NewAccount()
	if err != nil {
		dlog.Fatal("UnLock not effect")
	}

	wallet.Close()

	wallet.Open(password)

	for _, node := range nodes {
		reloadNode, err := wallet.GetAccountByAddress(node.Address)
		if err != nil {
			dlog.Fatal("reload wallet error")
		}
		pk := reloadNode.PrivateKey.PubKey()
		isOnCurve := secp256k1.S256().IsOnCurve(pk.X, pk.Y)
		if !isOnCurve {
			dlog.Fatal("error public key")
		}
		if reloadNode.PrivateKey == nil {
			dlog.Fatal("privateKey wallet error")
		}
	}
}
