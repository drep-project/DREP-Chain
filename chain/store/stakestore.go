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
	StakeStorage   = "StakeStorage"   //以地址作为KEY,存储stake相关内容

	pledgeLimit uint64 = 1000000    //候选节点需要抵押币的总数
	drepUnit    uint64 = 1000000000 //drep单位

	ChangeCycle = 100 //出块节点Change cycle
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
	key := sha3.Keccak256([]byte(StakeStorage + addr.Hex()))

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
	key := sha3.Keccak256([]byte(StakeStorage + addr.Hex()))
	value, err := binary.Marshal(storage)
	if err != nil {
		return err
	}

	return trieStore.store.Put(key, value)
}

func (trieStore *trieStakeStore) DelStakeStorage(addr *crypto.CommonAddress) error {
	key := sha3.Keccak256([]byte(StakeStorage + addr.Hex()))
	return trieStore.store.Delete(key)
}

func (trieStore *trieStakeStore) UpdateCandidateAddr(addr *crypto.CommonAddress, add bool) error {
	addrs, err := trieStore.GetCandidateAddrs()
	if err != nil {
		return err
	}

	if add {
		if len(addrs) > 0 {
			if _, ok := addrs[*addr]; ok {
				return nil
			}
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

	totalBalance := *addBalance
	if len(storage.ReceivedVoteCreditValue) == 0 {
		storage.ReceivedVoteCreditValue = make([]big.Int, 0)
		storage.ReceivedVoteCreditAddr = make([]crypto.CommonAddress, 0)

		storage.ReceivedVoteCreditAddr = append(storage.ReceivedVoteCreditAddr, *toAddr)
		storage.ReceivedVoteCreditValue = append(storage.ReceivedVoteCreditValue, totalBalance)
	} else {
		found := false
		for index, addr := range storage.ReceivedVoteCreditAddr {
			if addr.String() == toAddr.String() {
				totalBalance.Add(&storage.ReceivedVoteCreditValue[index], &totalBalance)
				storage.ReceivedVoteCreditValue[index] = totalBalance
				found = true
				break
			}
		}

		if !found {
			storage.ReceivedVoteCreditAddr = append(storage.ReceivedVoteCreditAddr, *toAddr)
			storage.ReceivedVoteCreditValue = append(storage.ReceivedVoteCreditValue, totalBalance)
		}
	}

	//投给自己，而且数量足够大
	if bytes.Equal(toAddr.Bytes(), fromAddr.Bytes()) && totalBalance.Cmp(new(big.Int).Mul(new(big.Int).SetUint64(pledgeLimit), new(big.Int).SetUint64(drepUnit))) >= 0 {
		trieStore.AddCandidateAddr(toAddr)
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
	if len(storage.ReceivedVoteCreditValue) == 0 {
		return fmt.Errorf("not exist vote credit")
	} else {
		found := false
		for index, addr := range storage.ReceivedVoteCreditAddr {
			if addr.String() == fromAddr.String() {
				found = true
				voteCredit := storage.ReceivedVoteCreditValue[index]
				resultBalance := new(big.Int)
				retCmp := voteCredit.Cmp(cancelBalance)
				if retCmp > 0 {
					storage.ReceivedVoteCreditValue[index] = *resultBalance.Sub(&voteCredit, cancelBalance)
				} else if retCmp == 0 {
					storage.ReceivedVoteCreditAddr = append(storage.ReceivedVoteCreditAddr[0:index], storage.ReceivedVoteCreditAddr[index+1:]...)
					storage.ReceivedVoteCreditValue = append(storage.ReceivedVoteCreditValue[0:index], storage.ReceivedVoteCreditValue[index+1:]...)
				} else {
					return fmt.Errorf("vote credit not enough")
				}

				if bytes.Equal(toAddr.Bytes(), fromAddr.Bytes()) && resultBalance.Cmp(new(big.Int).Mul(new(big.Int).SetUint64(pledgeLimit), new(big.Int).SetUint64(drepUnit))) < 0 {
					trieStore.DelCandidateAddr(toAddr)
				}
			}
		}

		if !found {
			return fmt.Errorf("not exist vote credit")
		}
	}

	err := trieStore.PutStakeStorage(toAddr, storage)
	if err != nil {
		return err
	}

	//目的stakeStorage；存储临时被退回的币
	if !bytes.Equal(toAddr.Bytes(), fromAddr.Bytes()) {
		storage, _ = trieStore.GetStakeStorage(fromAddr)
		if storage == nil {
			storage = &types.StakeStorage{}
		}
	}

	if len(storage.CancelVoteCreditValue) == 0 {
		storage.CancelVoteCreditValue = make([]big.Int, 0)
		storage.CancelVoteCreditValue[height] = *cancelBalance
	} else {
		found := false
		for index, vh := range storage.CancelVoteCreditHeight {
			if vh == height {
				found = true
				storage.CancelVoteCreditValue[index].Add(&storage.CancelVoteCreditValue[index], cancelBalance)
			}
		}
		if !found {
			storage.CancelVoteCreditHeight = append(storage.CancelVoteCreditHeight, height)
			storage.CancelVoteCreditValue = append(storage.CancelVoteCreditValue, *cancelBalance)
		}
	}

	return trieStore.PutStakeStorage(fromAddr, storage)
}

//取消抵押周期已经到，取消的币可以加入到account的balance中了
func (trieStore *trieStakeStore) GetCancelVoteCreditForBalance(addr *crypto.CommonAddress, height uint64) *big.Int {
	storage, _ := trieStore.GetStakeStorage(addr)
	if storage == nil {
		return &big.Int{}
	}

	total := new(big.Int)
	for index, cancelHeight := range storage.CancelVoteCreditHeight {
		if height >= cancelHeight+ChangeCycle {
			total.Add(total, &storage.CancelVoteCreditValue[index])
		}
	}

	return total
}

//取消抵押周期已经到，取消的币可以加入到account的balance中了
func (trieStore *trieStakeStore) CancelVoteCreditToBalance(addr *crypto.CommonAddress, height uint64) (*big.Int, error) {
	storage, _ := trieStore.GetStakeStorage(addr)
	if storage == nil {
		return &big.Int{}, nil
	}

	total := new(big.Int)
	for index, cancelHeight := range storage.CancelVoteCreditHeight {
		if height >= cancelHeight+ChangeCycle {
			total.Add(total, &storage.CancelVoteCreditValue[index])
			storage.CancelVoteCreditHeight = append(storage.CancelVoteCreditHeight[0:index], storage.CancelVoteCreditHeight[index+1:]...)
			storage.CancelVoteCreditValue = append(storage.CancelVoteCreditValue[0:index], storage.CancelVoteCreditValue[index+1:]...)
		}
	}

	err := trieStore.PutStakeStorage(addr, storage)
	if err != nil {
		return &big.Int{}, nil
	}
	return total, nil
}

//获取到候选人所有的质押金
func (trieStore *trieStakeStore) GetVoteCreditCount(addr *crypto.CommonAddress) *big.Int {
	storage, _ := trieStore.GetStakeStorage(addr)
	if storage == nil {
		return &big.Int{}
	}

	total := new(big.Int)
	for _, value := range storage.ReceivedVoteCreditValue {
		total.Add(total, &value)
	}

	return total
}

func (trieStore *trieStakeStore) GetVoteCreditDetails(addr *crypto.CommonAddress) map[crypto.CommonAddress]big.Int {
	m := make(map[crypto.CommonAddress]big.Int)
	storage, _ := trieStore.GetStakeStorage(addr)
	if storage == nil {
		return nil
	}

	for index,addr := range storage.ReceivedVoteCreditAddr{
		m[addr] = storage.ReceivedVoteCreditValue[index]
	}

	return m
}
