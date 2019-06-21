package service

import (
	"fmt"
	"github.com/drep-project/drep-chain/app"
	"github.com/drep-project/drep-chain/common"
	"log"
	"testing"
	"github.com/drep-project/dlog"
	"github.com/drep-project/drep-chain/crypto/secp256k1"
	"github.com/drep-project/drep-chain/crypto/sha3"
	accountTypes "github.com/drep-project/drep-chain/pkgs/accounts/types"
)

func TestWallet(t *testing.T) {
	password := string(sha3.Keccak256([]byte("AAAAAAAAAAAAAAAA")))
	rootChain := 0
	newNode := accountTypes.NewNode(nil, rootChain)
	fmt.Println(newNode)

	accountConfig := &accountTypes.Config{
		KeyStoreDir: "TestWallet",
	}
	wallet, err := NewWallet(accountConfig, rootChain)
	wallet.chainId = rootChain
	if err != nil {
		log.Fatal("NewWallet error")
	}

	err = wallet.Open(password)
	if err != nil {
		log.Fatal("open wallet error")
	}

	nodes := []*accountTypes.Node{}
	for i := 0; i < 10; i++ {
		node, err := wallet.NewAccount()
		pk := node.PrivateKey.PubKey()
		isOnCurve := secp256k1.S256().IsOnCurve(pk.X, pk.Y)
		if !isOnCurve {
			log.Fatal("error public key")
		}
		if err != nil {
			log.Fatal("open wallet error")
		}
		nodes = append(nodes, node)
	}

	wallet.Lock()
	_, err = wallet.NewAccount()
	if err == nil {
		log.Fatal("Lock not effect")
	}

	wallet.UnLock(password)
	_, err = wallet.NewAccount()
	if err != nil {
		log.Fatal("UnLock not effect")
	}

	wallet.Close()

	wallet.Open(password)

	for _, node := range nodes {
		reloadNode, err := wallet.GetAccountByAddress(node.Address)
		if err != nil {
			log.Fatal("reload wallet error")
		}
		pk := reloadNode.PrivateKey.PubKey()
		isOnCurve := secp256k1.S256().IsOnCurve(pk.X, pk.Y)
		if !isOnCurve {
			log.Fatal("error public key")
		}
		if reloadNode.PrivateKey == nil {
			log.Fatal("privateKey wallet error")
		}
	}
}
