package store

import (
	"bytes"
	"fmt"
	"math/big"

	"github.com/drep-project/binary"
	"github.com/drep-project/drep-chain/crypto"
	"github.com/drep-project/drep-chain/crypto/sha3"
	"github.com/drep-project/drep-chain/types"
)

const (
	candidateAddrs = "candidateAddrs" //参与竞选出块节点的地址集合
	stakeStorage   = "stakeStorage"   //以地址作为KEY,存储stake相关内容
)

type stakeStoreInterface interface {
	Get(key []byte) ([]byte, error)
	Put(key []byte, value []byte) error
	Delete(key []byte) error
}

type trieStakeStore struct {
	store stakeStoreInterface
}

func NewStakeStorage(store *StoreDB) *trieStakeStore {
	return &trieStakeStore{
		store: store,
	}
}

func (trieStore *trieStakeStore) GetStakeStorage(addr *crypto.CommonAddress) (*types.StakeStorage, error) {
	storage := &types.StakeStorage{}
	key := sha3.Keccak256([]byte(stakeStorage + addr.Hex()))

	value, err := trieStore.store.Get(key)
	if err != nil {
		log.Errorf("get storage err:%v", err)
		return nil, err
	}
	if value == nil {
		return nil, nil
	} else {
		err = binary.Unmarshal(value, storage)
		if err != nil {
			return nil, err
		}
	}
	return storage, nil
}

func (trieStore *trieStakeStore) PutStakeStorage(addr *crypto.CommonAddress, storage *types.StakeStorage) error {
	key := sha3.Keccak256([]byte(stakeStorage + addr.Hex()))
	value, err := binary.Marshal(storage)
	if err != nil {
		return err
	}

	return trieStore.store.Put(key, value)
}

func (trieStore *trieStakeStore) DelStakeStorage(addr *crypto.CommonAddress) error {
	key := sha3.Keccak256([]byte(stakeStorage + addr.Hex()))
	return trieStore.store.Delete(key)
}

func (trieStore *trieStakeStore) UpdateCandidateAddr(addr *crypto.CommonAddress, add bool) error {
	addrs, err := trieStore.GetCandidateAddrs()
	if err != nil {
		return err
	}

	if add {
		if len(addrs) > 0 {
			addrs[*addr] = struct{}{}
		} else {
			addrs = make(map[crypto.CommonAddress]struct{})
			addrs[*addr] = struct{}{}
		}
	} else { //del
		if len(addrs) == 0 {
			return nil
		} else {
			if _, ok := addrs[*addr]; ok {
				delete(addrs, *addr)
			}
		}
	}

	addrsBuf, err := binary.Marshal(addrs)
	if err == nil {
		trieStore.store.Put([]byte(candidateAddrs), addrsBuf)
	}
	return err
}

func (trieStore *trieStakeStore) AddCandidateAddr(addr *crypto.CommonAddress) error {
	return trieStore.UpdateCandidateAddr(addr, true)
}

func (trieStore *trieStakeStore) DelCandidateAddr(addr *crypto.CommonAddress) error {
	return trieStore.UpdateCandidateAddr(addr, false)
}

func (trieStore *trieStakeStore) GetCandidateAddrs() (map[crypto.CommonAddress]struct{}, error) {
	var addrsBuf []byte
	var err error
	key := []byte(candidateAddrs)
	addrs := make(map[crypto.CommonAddress]struct{})

	addrsBuf, err = trieStore.store.Get(key)
	if err != nil {
		log.Errorf("GetCandidateAddrs:%v", err)
		return nil, err
	}

	if addrsBuf == nil {
		return nil, nil
	}

	err = binary.Unmarshal(addrsBuf, &addrs)
	if err != nil {
		log.Errorf("GetCandidateAddrs, Unmarshal:%v", err)
		return nil, err
	}
	return addrs, nil
}

func (trieStore *trieStakeStore) VoteCredit(fromAddr, toAddr *crypto.CommonAddress, addBalance *big.Int) error {
	if toAddr == nil {
		toAddr = fromAddr
	}

	storage, _ := trieStore.GetStakeStorage(toAddr)
	if storage == nil {
		storage = &types.StakeStorage{}
	}

	if len(storage.ReceivedVoteCredit) == 0 {
		storage.ReceivedVoteCredit = make(map[crypto.CommonAddress]big.Int)
		storage.ReceivedVoteCredit[*fromAddr] = *addBalance
		trieStore.AddCandidateAddr(toAddr)
	} else {
		var totalBalance big.Int
		if v, ok := storage.ReceivedVoteCredit[*fromAddr]; ok {
			totalBalance = *addBalance.Add(addBalance, &v)
			storage.ReceivedVoteCredit[*fromAddr] = totalBalance
			//todo
		} else {
			storage.ReceivedVoteCredit[*fromAddr] = *addBalance
			trieStore.AddCandidateAddr(toAddr)
		}
	}

	return trieStore.PutStakeStorage(toAddr, storage)
}

func (trieStore *trieStakeStore) CancelVoteCredit(fromAddr, toAddr *crypto.CommonAddress, cancelBalance *big.Int, height uint64) error {
	if toAddr == nil {
		toAddr = fromAddr
	}

	//找到币被抵押到的stakeStorage;减去取消的值
	storage, _ := trieStore.GetStakeStorage(toAddr)
	if storage == nil {
		storage = &types.StakeStorage{}
	}
	if len(storage.ReceivedVoteCredit) == 0 {
		return fmt.Errorf("not exist vote credit")
	} else {
		if v, ok := storage.ReceivedVoteCredit[*fromAddr]; ok {
			resultBalance := new(big.Int)
			retCmp := v.Cmp(cancelBalance)
			if retCmp > 0 {
				storage.ReceivedVoteCredit[*fromAddr] = *resultBalance.Sub(&v, cancelBalance)
			} else if retCmp == 0 {
				delete(storage.ReceivedVoteCredit, *fromAddr)
				trieStore.DelCandidateAddr(fromAddr)
			} else {
				return fmt.Errorf("vote credit not enough")
			}
		} else {
			return fmt.Errorf("not exist vote credit")
		}
	}
	err :=  trieStore.PutStakeStorage(toAddr, storage)
	if err != nil {
		return err
	}

	//目的stakeStorage；存储临时被退回的币
	if bytes.Equal(toAddr.Bytes(), fromAddr.Bytes()) {
		storage.CancelVoteCredit[height] = *cancelBalance
	} else{
		storage, _ := trieStore.GetStakeStorage(fromAddr)
		if storage == nil {
			storage = &types.StakeStorage{}
		}
		storage.CancelVoteCredit[height] = *cancelBalance
	}
	return  trieStore.PutStakeStorage(fromAddr, storage)

}

func (trieStore *trieStakeStore) GetVoteCredit(addr *crypto.CommonAddress) map[crypto.CommonAddress]big.Int {
	storage, _ := trieStore.GetStakeStorage(addr)
	if storage == nil {
		return nil
	}

	return storage.ReceivedVoteCredit
}
