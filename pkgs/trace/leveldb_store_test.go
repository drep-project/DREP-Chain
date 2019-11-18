package trace

import (
	"bytes"
	"encoding/hex"
	"github.com/drep-project/DREP-Chain/app"
	"github.com/drep-project/DREP-Chain/common"
	"github.com/drep-project/DREP-Chain/common/math"
	"github.com/drep-project/DREP-Chain/crypto"
	"github.com/drep-project/DREP-Chain/crypto/secp256k1"
	"github.com/drep-project/DREP-Chain/types"
	"io/ioutil"
	"math/big"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"
)

var (
	fromAddr = "0x33e846059ebe404c4000902c864190456b642e9b"
	fromPriv = "cb5441fac80cf4e5438c6f873ecd5f3d0fa67d597f6c68bbc1106e8563e7f419"
	toAddr   = "0x2967f0629a5b84c981279ffbe330fe6154be7ad5"
	toPriv   = "987c5c49033044141f7a20fe411f27df041e326762d8f251c0323649d9006466"
)

func makeData(path string) (*LevelDbStore, []*types.Block) {
	levelDbStore, _ := NewLevelDbStore(path)
	testData := []*types.Block{}
	for i := 1; i < 10; i++ {
		block := randomBlock()
		testData = append(testData, block)
		levelDbStore.InsertRecord(block)
	}
	return levelDbStore, testData
}

func Test_LeveldbInsertAndExistRecord(t *testing.T) {
	path := "test_db1"
	levelDbStore, testData := makeData(path)
	for _, data := range testData {
		exist, err := levelDbStore.ExistRecord(data)
		if err != nil {
			t.Error(err)
		}
		if !exist {
			t.Errorf("expect exist in block but not found")
		}
	}
}

func Test_LeveldbInsertAndDelRecord(t *testing.T) {
	path := "test_db2"
	levelDbStore, testData := makeData(path)
	defer func() {
		levelDbStore.Close()
		deleteFolder(path)
	}()

	for _, data := range testData {
		levelDbStore.DelRecord(data)
	}
	for _, data := range testData {
		exist, err := levelDbStore.ExistRecord(data)
		if err != nil {
			t.Error(err)
		}
		if exist {
			t.Errorf("expect delete success but got a exist status")
		}
	}
}

func Test_LeveldbInsertAndGetRawTransaction(t *testing.T) {
	path := "test_db3"
	levelDbStore, testData := makeData(path)
	defer func() {
		levelDbStore.Close()
		deleteFolder(path)
	}()

	for _, data := range testData {
		for _, tx := range data.Data.TxList {
			txBytes, err := levelDbStore.GetRawTransaction(tx.TxHash())
			if err != nil {
				log.Error(err)
			}
			if !bytes.Equal(txBytes, tx.AsPersistentMessage()) {
				t.Errorf("tx raw in store not match real raw data")
			}
		}
	}
}

func Test_LeveldbInsertAndGetTransaction(t *testing.T) {
	path := "test_db4"
	levelDbStore, testData := makeData(path)
	defer func() {
		levelDbStore.Close()
		deleteFolder(path)
	}()

	for _, data := range testData {
		for _, tx := range data.Data.TxList {
			rpcTx, err := levelDbStore.GetTransaction(tx.TxHash())
			if err != nil {
				log.Error(err)
			}
			if *tx.To() != rpcTx.To {
				t.Errorf("tx message in store not match real tx")
			}
		}
	}
}

func Test_LeveldbInsertAndGetSendTransactionsByAddr(t *testing.T) {
	path := "test_db5"
	levelDbStore, testData := makeData(path)
	defer func() {
		levelDbStore.Close()
		deleteFolder(path)
	}()

	allCount := 0
	for _, data := range testData {
		allCount = allCount + int(data.Data.TxCount)
	}
	fromAddr := crypto.String2Address(fromAddr)
	all := levelDbStore.GetSendTransactionsByAddr(&fromAddr, 1, math.MaxInt32)
	if len(all) != allCount {
		t.Errorf("The total number of transactions does not match, real count %d but got %d", allCount, len(all))
	}

	for _, data := range testData {
		for _, tx := range data.Data.TxList {
			find := false
			for _, gotTx := range all {
				if bytes.Equal(tx.AsPersistentMessage(), gotTx.ToTx().AsPersistentMessage()) {
					find = true
				}
			}
			if !find {
				t.Error("transaction from store does not match the actual transaction")
			}
		}
	}
}

func Test_LeveldbGetSendTransactionsByAddrAndPagination(t *testing.T) {
	path := "test_db6"
	levelDbStore, testData := makeData(path)
	defer func() {
		levelDbStore.Close()
		deleteFolder(path)
	}()

	allCount := 0
	for _, data := range testData {
		allCount = allCount + int(data.Data.TxCount)
	}
	fromAddr := crypto.String2Address(fromAddr)
	all := levelDbStore.GetSendTransactionsByAddr(&fromAddr, 1, 3)
	if len(all) != 3 {
		t.Error("paging failure")
	}
	all = levelDbStore.GetSendTransactionsByAddr(&fromAddr, 2, 3)
	if len(all) != 3 {
		t.Error("paging failure")
	}
}

func Test_LeveldbInsertAndGetReceiveTransactionsByAddr(t *testing.T) {
	path := "test_db7"
	levelDbStore, testData := makeData(path)
	defer func() {
		levelDbStore.Close()
		deleteFolder(path)
	}()

	allCount := 0
	for _, data := range testData {
		allCount = allCount + int(data.Data.TxCount)
	}
	toAddr := crypto.String2Address(toAddr)
	all := levelDbStore.GetReceiveTransactionsByAddr(&toAddr, 1, math.MaxInt32)
	if len(all) != allCount {
		t.Errorf("The total number of receive transactions does not match, real count %d but got %d", allCount, len(all))
	}

	for _, data := range testData {
		for _, tx := range data.Data.TxList {
			find := false
			for _, gotTx := range all {
				if bytes.Equal(tx.AsPersistentMessage(), gotTx.ToTx().AsPersistentMessage()) {
					find = true
				}
			}
			if !find {
				t.Error("transaction from store does not match the actual transaction")
			}
		}
	}
}

func Test_LeveldbGetReceiveTransactionsByAddrAndPagination(t *testing.T) {
	path := "test_db8"
	levelDbStore, testData := makeData(path)
	defer func() {
		levelDbStore.Close()
		deleteFolder(path)
	}()

	allCount := 0
	for _, data := range testData {
		allCount = allCount + int(data.Data.TxCount)
	}
	toAddr := crypto.String2Address(toAddr)
	all := levelDbStore.GetReceiveTransactionsByAddr(&toAddr, 1, 3)
	if len(all) != 3 {
		t.Error("receive paging failure")
	}
	all = levelDbStore.GetReceiveTransactionsByAddr(&toAddr, 2, 3)
	if len(all) != 3 {
		t.Error("receive paging failure")
	}
}

func deleteFolder(ketStore string) {
	fileInfo, _ := ioutil.ReadDir(ketStore)
	for _, file := range fileInfo {
		path := filepath.Join(ketStore, file.Name())
		os.Remove(path)
	}
	os.Remove(ketStore)
}

func seuqenceBlock(n int) []*types.Block {
	blocks := make([]*types.Block, n)
	for i := 0; i < n; i++ {
		block := randomBlock()
		block.Header.Height = uint64(i)
		blocks[i] = block
	}
	return blocks
}
func randomBlock() *types.Block {
	txData := []*types.Transaction{
		randTransaction(), randTransaction(),
	}
	priBytes, _ := hex.DecodeString(fromPriv)
	priv, _ := secp256k1.PrivKeyFromScalar(priBytes)
	block := &types.Block{
		Header: &types.BlockHeader{
			ChainId:        app.ChainIdType{},
			Version:        rand.Int31(),
			PreviousHash:   crypto.RandomHash(),
			GasLimit:       *big.NewInt(rand.Int63()),
			GasUsed:        *big.NewInt(rand.Int63()),
			Height:         uint64(rand.Int63()),
			Timestamp:      uint64(time.Now().Nanosecond()),
			StateRoot:      crypto.RandomHash().Bytes(),
			TxRoot:         crypto.RandomHash().Bytes(),
			LeaderAddress:  *priv.PubKey(),
			MinorAddresses: []secp256k1.PublicKey{},
		},
		Data: &types.BlockData{
			TxList:  txData,
			TxCount: uint64(len(txData)),
		},
	}
	return block
}

func randTransaction() *types.Transaction {
	buf := make([]byte, 20)
	rand.Read(buf)
	priBytes, _ := hex.DecodeString(fromPriv)
	priv, _ := secp256k1.PrivKeyFromScalar(priBytes)
	transaction := types.Transaction{
		Data: types.TransactionData{
			Version:   rand.Int31(),
			Nonce:     uint64(rand.Int31()),
			Type:      types.TxType(0),
			To:        crypto.String2Address(toAddr),
			ChainId:   app.ChainIdType{},
			Amount:    common.Big(*big.NewInt(rand.Int63())),
			GasPrice:  common.Big(*big.NewInt(rand.Int63())),
			GasLimit:  common.Big(*big.NewInt(rand.Int63())),
			Timestamp: rand.Int63(),
			Data:      buf,
		},
	}
	sig, _ := secp256k1.SignCompact(priv, transaction.TxHash().Bytes(), true)
	transaction.Sig = sig
	return &transaction
}
