package database

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/drep-project/drep-chain/app"
	"math/big"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"strconv"
	"strings"
	"testing"

	chainType "github.com/drep-project/drep-chain/types"
	"github.com/drep-project/drep-chain/crypto"
)

func TestGetSetAlias(t *testing.T) {
	os.RemoveAll("./test/")
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

	idb.BeginTransaction(true)
	err = idb.AliasSet(&addr, alias)
	if err != nil {
		fmt.Println(err)
		return
	}
	idb.Commit()

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
	idb.BeginTransaction(true)
	err = idb.AliasSet(&addr, "")
	if err != nil {
		fmt.Println(err)
		return
	}
	idb.Commit()

	addr1 = idb.AliasGet(alias)
	if addr1 != nil {
		t.Fatal("aliase has deleted")
	}

	alias2 = idb.GetStorageAlias(&addr)
	if "" != alias2 {
		t.Fatal(alias2)
		return
	}

	os.RemoveAll("./test/")
}

func TestPutGetStorage(t *testing.T) {
	os.RemoveAll("./test/")

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
		Balance: *new(big.Int).SetInt64(111111),
		Nonce:   1,
	}

	err = idb.PutStorage(&addr, &st)
	if err != nil {
		t.Fatal(err)
	}

	store := idb.GetStorage(&addr)
	if store == nil || store.Nonce != 1 {
		t.Fatal("storage not exist")
	}

	os.RemoveAll("./test/")
}

func TestRollBack(t *testing.T) {
	defer os.RemoveAll("./test/")
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

	//idb.RecordBlockJournal(uint64(0))
	for i = 0; i < 5; i++ {
		idb.BeginTransaction(true)
		for j = 0; j < 10; j++ {
			pk, err := crypto.GenerateKey(rand.Reader)
			addr = crypto.Bytes2Address(pk.PubKey().Serialize())
			err = idb.AliasSet(&addr, alias+strconv.Itoa(int(i*10+j)))
			if err != nil {
				fmt.Println(err)
				return
			}
		}
		idb.Commit()
		//idb.RecordBlockJournal(uint64(i + 1))
	}

	seqVal, err := db.diskDb.Get([]byte(dbOperaterMaxSeqKey))
	seq := new(big.Int).SetBytes(seqVal)

	if seq.Uint64() != (i)*(j)*2+1 {
		t.Fatal("operate journal count err", seq)
	}

	err, n := idb.Rollback2Block(uint64(5), &crypto.Hash{})
	if err != nil {
		t.Fatal("roolback err", err)
	}
	if n != 0 {
		t.Fatal("2 roolback err", n)
	}

	err, n = idb.Rollback2Block(0, &crypto.Hash{})
	if err != nil {
		t.Fatal("roolback err", err)
	}

	if n != int64(i*j*2) {
		t.Fatal("2 roolback err", err)
	}
}

func TestDatabaseInit(t *testing.T) {
	dbs := DatabaseService{}
	dbs.config = &DatabaseConfig{}
	executeContext := app.ExecuteContext{}

	err := dbs.Init(&executeContext)
	if err == nil {
		t.Fatal("err, init must fail")
	}

	executeContext.AddService(&dbs)
	executeContext.CommonConfig = &app.CommonConfig{ConfigFile:"config.json"}

	//common.AppDataDir("testDatebase", false)
	executeContext.CommonConfig.HomeDir,_ = Home()

	executeContext.CommonConfig.HomeDir += "/testdb/data"

	os.RemoveAll(executeContext.CommonConfig.HomeDir)

	pc := make(map[string]json.RawMessage)
	dc := DatabaseConfig{}
	byteDC, _ := json.Marshal(&dc)

	rm := &json.RawMessage{}
	rm.UnmarshalJSON(byteDC)
	bc, _ := rm.MarshalJSON()
	pc["database"] = bc
	executeContext.PhaseConfig = pc

	err = dbs.Init(&executeContext)
	if err != nil {
		t.Fatal("init must fail")
	}

	os.RemoveAll(executeContext.CommonConfig.HomeDir)
}


// Home returns the home directory for the executing user.
//
// This uses an OS-specific method for discovering the home directory.
// An error is returned if a home directory cannot be detected.
func Home() (string, error) {
	user, err := user.Current()
	if nil == err {
		return user.HomeDir, nil
	}

	// cross compile support
	if "windows" == runtime.GOOS {
		return homeWindows()
	}

	// Unix-like system, so just assume Unix
	return homeUnix()
}

func homeUnix() (string, error) {
	// First prefer the HOME environmental variable
	if home := os.Getenv("HOME"); home != "" {
		return home, nil
	}

	// If that fails, try the shell
	var stdout bytes.Buffer
	cmd := exec.Command("sh", "-c", "eval echo ~$USER")
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return "", err
	}

	result := strings.TrimSpace(stdout.String())
	if result == "" {
		return "", errors.New("blank output when reading home directory")
	}

	return result, nil
}

func homeWindows() (string, error) {
	drive := os.Getenv("HOMEDRIVE")
	path := os.Getenv("HOMEPATH")
	home := drive + path
	if drive == "" || path == "" {
		home = os.Getenv("USERPROFILE")
	}
	if home == "" {
		return "", errors.New("HOMEDRIVE, HOMEPATH, and USERPROFILE are blank")
	}

	return home, nil
}