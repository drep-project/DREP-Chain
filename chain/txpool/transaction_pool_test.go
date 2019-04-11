package txpool

import (
	"fmt"
	"github.com/drep-project/drep-chain/common"
	"github.com/drep-project/drep-chain/common/event"
	"github.com/drep-project/drep-chain/crypto"
	"github.com/drep-project/drep-chain/crypto/secp256k1"
	"github.com/drep-project/drep-chain/database"
	chainTypes "github.com/drep-project/drep-chain/chain/types"
	"math/big"
	"testing"
	"time"
)

var txNum1 uint64 = 1000000
var txNum2 int = 100000
var txPool *TransactionPool
var feed event.Feed

func TestNewTransactions(t *testing.T) {
	diskDb, err := database.NewDatabase("./data")
	if err != nil {
		t.Error("db init err")
	}
	db := database.NewDatabaseService(diskDb)
	txPool = NewTransactionPool(db)
	if txPool == nil {
		t.Error("init database service err")
	}

	txPool.Start(&feed)
}

func addTx(t *testing.T, num uint64) error {
	b := common.Bytes("0x0373654ccdb250f2cfcfe64c783a44b9ea85bc47f2f00c480d05082428d277d6d0")
	b.UnmarshalText(b)
	pubkey, _ := secp256k1.ParsePubKey(b)
	addr := crypto.PubKey2Address(pubkey)
	fmt.Println(string(addr.Hex()))
	txPool.databaseApi.BeginTransaction()

	var amount uint64 = 0xefffffffffffffff
	txPool.databaseApi.PutBalance(&addr, new(big.Int).SetUint64(amount), true)
	txPool.databaseApi.Commit()

	nonce := txPool.databaseApi.GetNonce(&addr, false)
	for i := 0; uint64(i) < num; i++ {
		tx := chainTypes.NewTransaction(addr, crypto.CommonAddress{}, new(big.Int).SetInt64(100), nonce+uint64(i))
		err := txPool.AddTransaction(tx)
		if err != nil {
			return err
		}
		if i%10000 == 0 {
			time.Sleep(time.Second * 1)
			fmt.Println("nonce id:", i)
		}
	}

	return nil
}

func TestAddTX(t *testing.T) {
	TestNewTransactions(t)
	err := addTx(t, txNum1)
	if err != nil {
		t.Fatal(err)
	}
}

func TestAddIntevalTX(t *testing.T) {

	b := common.Bytes("0x03177b8e4ef31f4f801ce00260db1b04cc501287e828692a404fdbc46c7ad6ff26")
	b.UnmarshalText(b)
	pubkey, _ := secp256k1.ParsePubKey(b)
	addr := crypto.PubKey2Address(pubkey)
	for i := 0; i < txNum2; i++ {
		if i != 0 && i%100 == 0 {
			continue
		}

		tx := chainTypes.NewTransaction(addr, crypto.CommonAddress{}, new(big.Int).SetUint64(100000000), uint64(i))
		txPool.AddTransaction(tx)
	}
}

func TestSyncTx(t *testing.T) {
	feed.Send(struct{}{})
}

//池子里面的都是未处理的交易
func TestGetPendingTxs(t *testing.T) {
	TestNewTransactions(t)
	ch := make(chan uint64)
	go func() {
		err := addTx(t, txNum1)
		if err != nil {
			return
		}

		time.Sleep(time.Second * 10)
		ch <- txNum1
	}()

	func() {
		var nonce uint64
		for {
			select {
			case num := <-ch:
				if nonce != num {
					t.Fatalf("recv nonce:%d sendTxNum:%d", nonce, num)
				}
				break
			default:
				gasLimit := new(big.Int).SetInt64(10000000)
				mapTxs := txPool.GetPending(gasLimit)
				fmt.Println("pending tx len:", len(mapTxs))
				if len(mapTxs) == 0 {
					time.Sleep(time.Second * 1)
					continue
				}

				var addrs []*crypto.CommonAddress
				mapAddr := make(map[crypto.CommonAddress]struct{})
				for _, tx := range mapTxs {
					from := tx.From()
					if _, ok := mapAddr[*from]; !ok {
						mapAddr[*from] = struct{}{}
						addrs = append(addrs, from)
					}
					nonce = tx.Nonce()
				}
				fmt.Println("recv nonce:", nonce)
				txPool.databaseApi.BeginTransaction()
				txPool.databaseApi.PutNonce(addrs[0], nonce+1, false)
				txPool.databaseApi.Commit()

				feed.Send(addrs)
				time.Sleep(time.Second * 1)
			}
		}
	}()
}
