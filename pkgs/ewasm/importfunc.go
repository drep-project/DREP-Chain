package dwasm

import (
	"errors"
	"fmt"
	"github.com/drep-project/drep-chain/crypto"
	"github.com/drep-project/drep-chain/crypto/sha3"
	"github.com/go-interpreter/wagon/exec"
	"math"
)

type ImportFunc struct {
	State       	*State
	Time        	uint64
	TxHash			crypto.Hash
	ContractName    string
}

func (importFunc *ImportFunc) Debug(proc *exec.Process, ptr uint32, len uint32){
	buf := make([]byte, len)
	_, err := proc.ReadAt(buf, int64(ptr))
	if err != nil {
		panic(err)
	}
	fmt.Println(string(buf))
}

// ------------------------------Block Message--------------------------

func (importFunc *ImportFunc) TimeStamp(proc *exec.Process) uint64{
	//self.checkGas(TIME_STAMP_GAS)
	return importFunc.Time
}

func (importFunc *ImportFunc) GetTxHash(proc *exec.Process, ptr uint32) uint32{
	length, err := proc.WriteAt(importFunc.TxHash[:], int64(ptr))
	if err != nil {
		panic(err)
	}

	return uint32(length)
}

// ------------------------------Database Message--------------------------
func (importFunc *ImportFunc) StorageRead(proc *exec.Process, keyPtr uint32, klen uint32, val uint32, vlen uint32, offset uint32) uint32 {
	//self.checkGas(STORAGE_GET_GAS)
	keybytes := make([]byte, klen)
	_, err := proc.ReadAt(keybytes, int64(keyPtr))
	if err != nil {
		panic(err)
	}


	item := importFunc.State.Load(sha3.HashS256([]byte(importFunc.ContractName), keybytes))
	if item == nil {
		return math.MaxUint32
	}
	length := vlen
	itemlen := uint32(len(item))
	if itemlen < vlen {
		length = itemlen
	}

	if uint32(len(item)) < offset {
		panic(errors.New("offset is invalid"))
	}
	_, err = proc.WriteAt(item[offset:offset+length], int64(val))

	if err != nil {
		panic(err)
	}
	return uint32(len(item))
}

func (importFunc *ImportFunc) StorageWrite(proc *exec.Process, keyPtr uint32, keylen uint32, valPtr uint32, valLen uint32) {
	keybytes := make([]byte, keylen)
	_, err := proc.ReadAt(keybytes, int64(keyPtr))
	if err != nil {
		panic(err)
	}

	valbytes := make([]byte, valLen)
	_, err = proc.ReadAt(valbytes, int64(valPtr))
	if err != nil {
		panic(err)
	}

	modifiedLoc := sha3.HashS256([]byte(importFunc.ContractName), keybytes)
	importFunc.State.Store(modifiedLoc, sha3.HashS256(valbytes))
}

func (importFunc *ImportFunc) StorageDelete(proc *exec.Process, keyPtr uint32, keylen uint32) {
	keybytes := make([]byte, keylen)
	_, err := proc.ReadAt(keybytes, int64(keyPtr))
	if err != nil {
		panic(err)
	}
	importFunc.State.Delete(keybytes)
}

func (importFunc *ImportFunc) GetBalance(proc *exec.Process, keyPtr uint32, klen uint32, val uint32, vlen uint32, offset uint32) uint32{
	//self.checkGas(STORAGE_GET_GAS)
	keybytes := make([]byte, klen)
	_, err := proc.ReadAt(keybytes, int64(keyPtr))
	if err != nil {
		panic(err)
	}

	balance := importFunc.State.GetBalance(string(keybytes)).Bytes()
	if balance == nil {
		return math.MaxUint32
	}
	length := vlen
	itemlen := uint32(len(balance))
	if itemlen < vlen {
		length = itemlen
	}

	if uint32(len(balance)) < offset {
		panic(errors.New("offset is invalid"))
	}
	_, err = proc.WriteAt(balance[offset:offset+length], int64(val))

	if err != nil {
		panic(err)
	}
	return uint32(len(balance))
}

func (importFunc *ImportFunc) GetReputation(proc *exec.Process, keyPtr uint32, klen uint32, val uint32, vlen uint32, offset uint32) uint32 {
	//self.checkGas(STORAGE_GET_GAS)
	keybytes := make([]byte, klen)
	_, err := proc.ReadAt(keybytes, int64(keyPtr))
	if err != nil {
		panic(err)
	}

	reputation := importFunc.State.GetReputation(string(keybytes)).Bytes()
	if reputation == nil {
		return math.MaxUint32
	}
	length := vlen
	itemlen := uint32(len(reputation))
	if itemlen < vlen {
		length = itemlen
	}

	if uint32(len(reputation)) < offset {
		panic(errors.New("offset is invalid"))
	}
	_, err = proc.WriteAt(reputation[offset:offset+length], int64(val))

	if err != nil {
		panic(err)
	}
	return uint32(len(reputation))
}

func (importFunc *ImportFunc) Notify(proc *exec.Process, ptr uint32, len uint32) {
	bs := make([]byte, len)
	_, err := proc.ReadAt(bs, int64(ptr))
	if err != nil {
		panic(err)
	}

	//	notify := &event.NotifyEventInfo{self.Service.ContextRef.CurrentContext().ContractAddress, string(bs)}
	//	notifys := make([]*event.NotifyEventInfo, 1)
	//	notifys[0] = notify
	//	importFunc.Service.ContextRef.PushNotifications(notifys)
	fmt.Println(string(bs))
}

func (importFunc *ImportFunc) ValidateAccount(proc *exec.Process, ptr uint32, len uint32) uint32 {
	bs := make([]byte, len)
	_, err := proc.ReadAt(bs, int64(ptr))
	if err != nil {
		panic(err)
	}

	val, _ := importFunc.State.databaseApi.GetStorage(string(bs), true)
	if val != nil {
		return 1
	}
	return 0

	//	notify := &event.NotifyEventInfo{self.Service.ContextRef.CurrentContext().ContractAddress, string(bs)}
	//	notifys := make([]*event.NotifyEventInfo, 1)
	//	notifys[0] = notify
	//	importFunc.Service.ContextRef.PushNotifications(notifys)
}