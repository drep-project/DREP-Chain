package trace

import (
	"bytes"
	"fmt"
	"github.com/drep-project/drep-chain/chain/types"
	"github.com/drep-project/drep-chain/crypto"
	"github.com/drep-project/drep-chain/common/math"
	"testing"
)

var (
	testMongoUrl = "mongodb://localhost:27017"
)

func makeMongoData(dbName string) (*MongogDbStore,[]*types.Block) {
	mongoStore, err := NewMongogDbStore(testMongoUrl, dbName)
	if err != nil {
		fmt.Println(err)
	}
	testData := []*types.Block{}
	for i:=1;i<10;i++ {
		block := randomBlock()
		testData = append(testData, block)
		mongoStore.InsertRecord(block)
	}
	return mongoStore, testData
}


func Test_MongoInsertAndExistRecord(t *testing.T) {
	db := "drep_test_1"
	mongoStore, testData := makeMongoData(db)
	defer func() {
		deleteDb(mongoStore)
		mongoStore.Close()
	}()
	for _, data := range testData {
		exist, err := mongoStore.ExistRecord(data)
		if err != nil {
			t.Error(err)
		}
		if !exist {
			t.Errorf("expect exist in block but not found")
		}
	}
}

func Test_MongoInsertAndDelRecord(t *testing.T) {
	db := "drep_test_2"
	mongoStore, testData := makeMongoData(db)
	defer func() {
		deleteDb(mongoStore)
		mongoStore.Close()
	}()

	for _, data := range testData {
		mongoStore.DelRecord(data)
	}
	for _, data := range testData {
		exist, err := mongoStore.ExistRecord(data)
		if err != nil {
			t.Error(err)
		}
		if exist {
			t.Errorf("expect delete success but got a exist status")
		}
	}
}

func Test_MongoInsertAndGetRawTransaction(t *testing.T) {
	db := "drep_test_3"
	mongoStore, testData := makeMongoData(db)
	defer func() {
		deleteDb(mongoStore)
		mongoStore.Close()
	}()

	for _, data := range testData {
		for _, tx := range data.Data.TxList {
			txBytes, err := mongoStore.GetRawTransaction(tx.TxHash())
			if err != nil {
				log.Error(err)
			}
			if !bytes.Equal(txBytes, tx.AsPersistentMessage()) {
				t.Errorf("tx raw in store not match real raw data")
			}
		}
	}
}

func Test_MongoInsertAndGetTransaction(t *testing.T) {
	db := "drep_test_4"
	mongoStore, testData := makeMongoData(db)
	defer func() {
		deleteDb(mongoStore)
		mongoStore.Close()
	}()

	for _, data := range testData {
		for _, tx := range data.Data.TxList {
			rpcTx, err := mongoStore.GetTransaction(tx.TxHash())
			if err != nil {
				log.Error(err)
			}
			if *tx.To() != rpcTx.To {
				t.Errorf("tx message in store not match real tx")
			}
		}
	}
}

func Test_MongoInsertAndGetSendTransactionsByAddr(t *testing.T) {
	db := "drep_test_6"
	mongoStore, testData := makeMongoData(db)
	defer func() {
		deleteDb(mongoStore)
		mongoStore.Close()
	}()

	allCount := 0
	for _, data := range testData {
		allCount = allCount + int(data.Data.TxCount)
	}
	fromAddr := crypto.String2Address(fromAddr)
	all := mongoStore.GetSendTransactionsByAddr(&fromAddr,1 ,math.MaxInt32)
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

func Test_MongoGetSendTransactionsByAddrAndPagination (t *testing.T) {
	db := "drep_test_7"
	mongoStore, testData := makeMongoData(db)
	defer func() {
		deleteDb(mongoStore)
		mongoStore.Close()
	}()

	allCount := 0
	for _, data := range testData {
		allCount = allCount + int(data.Data.TxCount)
	}
	fromAddr := crypto.String2Address(fromAddr)
	all := mongoStore.GetSendTransactionsByAddr(&fromAddr,1 ,3)
	if len(all) != 3 {
		t.Error("paging failure")
	}
	all = mongoStore.GetSendTransactionsByAddr(&fromAddr,2 ,3)
	if len(all) != 3 {
		t.Error("paging failure")
	}
}


func Test_MongoInsertAndGetReceiveTransactionsByAddr(t *testing.T) {
	db := "drep_test_8"
	mongoStore, testData := makeMongoData(db)
	defer func() {
		deleteDb(mongoStore)
		mongoStore.Close()
	}()

	allCount := 0
	for _, data := range testData {
		allCount = allCount + int(data.Data.TxCount)
	}
	toAddr := crypto.String2Address(toAddr)
	all := mongoStore.GetReceiveTransactionsByAddr(&toAddr,1 ,math.MaxInt32)
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

func Test_MongoGetReceiveTransactionsByAddrAndPagination (t *testing.T) {
	db := "drep_test_9"
	mongoStore, testData := makeMongoData(db)
	defer func() {
		deleteDb(mongoStore)
		mongoStore.Close()
	}()

	allCount := 0
	for _, data := range testData {
		allCount = allCount + int(data.Data.TxCount)
	}
	toAddr := crypto.String2Address(toAddr)
	all := mongoStore.GetReceiveTransactionsByAddr(&toAddr,1 ,3)
	if len(all) != 3 {
		t.Error("receive paging failure")
	}
	all = mongoStore.GetReceiveTransactionsByAddr(&toAddr,2 ,3)
	if len(all) != 3 {
		t.Error("receive paging failure")
	}
}


func deleteDb(store *MongogDbStore) {
	err := store.db.Drop(nil)
	if err != nil {
		fmt.Println(err)
	}
}



