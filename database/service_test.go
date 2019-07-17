package database

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"math/big"
	"os"
	"strconv"
	"testing"

	chainType "github.com/drep-project/drep-chain/chain/types"
	"github.com/drep-project/drep-chain/crypto"
)

func TestGetSetAlias(t *testing.T) {
	db, err := NewDatabase("./test/")
	if err != nil {
		fmt.Println(err)
		return
	}
	idb := NewDatabaseService(db)
	addrStr := "0xc4ac59f52b3052e5c14566ed397453ea913c6fbc"
	addr := crypto.CommonAddress{}
	addr.SetBytes([]byte(addrStr))
	alias := "115108924-test"

	idb.BeginTransaction()
	err = idb.AliasSet(&addr, alias)
	if err != nil {
		fmt.Println(err)
		return
	}
	idb.Commit(false)

	addr1 := idb.AliasGet(alias)
	if addr1 == nil || !bytes.Equal(addr1.Bytes(), addr.Bytes()) {
		t.Fatal("alias get set err,", addr1)
	}

	alias2 := idb.GetStorageAlias(&addr)
	if alias != alias2 {
		t.Fatal(alias, alias)
		return
	}

	//测试2
	idb.BeginTransaction()
	err = idb.AliasSet(&addr, "")
	if err != nil {
		fmt.Println(err)
		return
	}
	idb.Commit(true)

	addr1 = idb.AliasGet(alias)
	if addr1 != nil {
		t.Fatal("aliase has deleted")
	}

	alias2 = idb.GetStorageAlias(&addr)
	if "" != alias2 {
		t.Fatal(alias2)
		return
	}

	os.Remove("./test/")
}

func TestPutStorage(t *testing.T) {
	db, err := NewDatabase("./test/")
	if err != nil {
		fmt.Println(err)
		return
	}
	idb := NewDatabaseService(db)
	addrStr := "0xc4ac59f52b3052e5c14566ed397453ea913c6fbc"
	addr := crypto.CommonAddress{}
	addr.SetBytes([]byte(addrStr))

	st := chainType.Storage{

	}

	err = idb.PutStorage(&addr, &st)
	if err != nil {
		t.Fatal(err)
	}

}

func TestGetStorage(t *testing.T) {
	db, err := NewDatabase("./test/")
	if err != nil {
		fmt.Println(err)
		return
	}
	idb := NewDatabaseService(db)
	addrStr := "0xc4ac59f52b3052e5c14566ed397453ea913c6fbc"
	addr := crypto.CommonAddress{}
	addr.SetBytes([]byte(addrStr))

	store := idb.GetStorage(&addr)
	if store == nil {
		t.Fatal("storage not exist")
	}
}

func TestRollBack(t *testing.T) {
	db, err := NewDatabase("./test/")
	if err != nil {
		fmt.Println(err)
		return
	}

	idb := NewDatabaseService(db)
	addrStr := "0xc4ac59f52b3052e5c14566ed397453ea913c6fbc"
	addr := crypto.CommonAddress{}
	addr.SetBytes([]byte(addrStr))
	alias := "115108924-test"
	var i, j uint64

	idb.RecordBlockJournal(uint64(0))
	for i = 0; i < 5; i++ {
		idb.BeginTransaction()
		for j = 0; j < 10; j++ {
			pk, err := crypto.GenerateKey(rand.Reader)
			addr = crypto.Bytes2Address(pk.PubKey().Serialize())
			err = idb.AliasSet(&addr, alias+strconv.Itoa(int(i*10+j)))
			if err != nil {
				fmt.Println(err)
				return
			}
		}
		idb.Commit(true)
		idb.RecordBlockJournal(uint64(i + 1))
	}

	seqVal, err := db.diskDb.Get([]byte(dbOperaterMaxSeqKey))
	seq := new(big.Int).SetBytes(seqVal)

	if seq.Uint64() != (i)*(j)*2+1 {
		t.Fatal("operate journal count err", seq)
	}

	err, n := idb.Rollback2Block(5)
	if err != nil {
		t.Fatal("roolback err", err)
	}
	if n != 0 {
		t.Fatal("2 roolback err", n)
	}

	err, n = idb.Rollback2Block(0)
	if err != nil {
		t.Fatal("roolback err", err)
	}

	if n != int64(i*j*2) {
		t.Fatal("2 roolback err", err)
	}
}

//测试操作日志 && 混合普通交易与alias交易
