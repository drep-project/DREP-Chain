package store

import (
	"errors"
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

	registerPledgeLimit uint64 = 1000000    //候选节点需要抵押币的总数
	drepUnit            uint64 = 1000000000 //drep币最小单位

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

func (trieStore *trieStakeStore) GetCandidateData(addr *crypto.CommonAddress) ([]byte, error) {
	storage, _ := trieStore.GetStakeStorage(addr)
	if storage == nil {
		return []byte{}, errors.New("addr stake no exist")
	}

	return storage.CandidateData, nil
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
	if toAddr == nil || fromAddr == nil {
		return errors.New("addr cannot equal nil")
	}

	if fromAddr.String() == toAddr.String() {
		return errors.New("from euqal to addr")
	}

	storage, _ := trieStore.GetStakeStorage(toAddr)
	if storage == nil {
		storage = &types.StakeStorage{}
	}

	totalBalance := *addBalance
	if len(storage.ReceivedCreditValue) == 0 {
		storage.ReceivedCreditValue = make([]big.Int, 0)
		storage.ReceivedCreditAddr = make([]crypto.CommonAddress, 0)

		storage.ReceivedCreditAddr = append(storage.ReceivedCreditAddr, *fromAddr)
		storage.ReceivedCreditValue = append(storage.ReceivedCreditValue, totalBalance)
	} else {
		found := false
		for index, addr := range storage.ReceivedCreditAddr {
			if addr.String() == fromAddr.String() {
				totalBalance.Add(&storage.ReceivedCreditValue[index], &totalBalance)
				storage.ReceivedCreditValue[index] = totalBalance
				found = true
				break
			}
		}

		if !found {
			storage.ReceivedCreditAddr = append(storage.ReceivedCreditAddr, *fromAddr)
			storage.ReceivedCreditValue = append(storage.ReceivedCreditValue, totalBalance)
		}
	}

	return trieStore.PutStakeStorage(toAddr, storage)
}

func (trieStore *trieStakeStore) CancelVoteCredit(fromAddr, toAddr *crypto.CommonAddress, cancelBalance *big.Int, height uint64) error {
	if toAddr == nil || fromAddr == nil {
		return errors.New("addr cannot equal nil")
	}

	if fromAddr.String() == toAddr.String() {
		return errors.New("from euqal to addr")
	}

	//找到币被抵押到的stakeStorage;减去取消的值
	storage, _ := trieStore.GetStakeStorage(toAddr)
	if storage == nil {
		storage = &types.StakeStorage{}
	}
	if len(storage.ReceivedCreditValue) == 0 {
		return fmt.Errorf("not exist vote credit")
	} else {
		found := false
		for index, addr := range storage.ReceivedCreditAddr {
			if addr.String() == fromAddr.String() {
				found = true
				voteCredit := storage.ReceivedCreditValue[index]
				resultBalance := new(big.Int)
				retCmp := voteCredit.Cmp(cancelBalance)
				if retCmp > 0 {
					storage.ReceivedCreditValue[index] = *resultBalance.Sub(&voteCredit, cancelBalance)
				} else if retCmp == 0 {
					storage.ReceivedCreditAddr = append(storage.ReceivedCreditAddr[0:index], storage.ReceivedCreditAddr[index+1:]...)
					storage.ReceivedCreditValue = append(storage.ReceivedCreditValue[0:index], storage.ReceivedCreditValue[index+1:]...)
				} else {
					return fmt.Errorf("vote credit not enough")
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
	storage, _ = trieStore.GetStakeStorage(fromAddr)
	if storage == nil {
		storage = &types.StakeStorage{}
	}

	if len(storage.CancelCreditValue) == 0 {
		storage.CancelCreditValue = make([]big.Int, 0, 1)
		storage.CancelCreditHeight = make([]uint64, 0, 1)
		storage.CancelCreditHeight = append(storage.CancelCreditHeight, height)
		storage.CancelCreditValue = append(storage.CancelCreditValue, *cancelBalance)
	} else {
		found := false
		for index, vh := range storage.CancelCreditHeight {
			if vh == height {
				found = true
				storage.CancelCreditValue[index].Add(&storage.CancelCreditValue[index], cancelBalance)
			}
		}
		if !found {
			storage.CancelCreditHeight = append(storage.CancelCreditHeight, height)
			storage.CancelCreditValue = append(storage.CancelCreditValue, *cancelBalance)
		}
	}

	return trieStore.PutStakeStorage(fromAddr, storage)
}

//取消抵押周期已经到，取消的币可以加入到account的balance中了
func (trieStore *trieStakeStore) GetCancelCreditForBalance(addr *crypto.CommonAddress, height uint64) *big.Int {
	storage, _ := trieStore.GetStakeStorage(addr)
	if storage == nil {
		return &big.Int{}
	}

	total := new(big.Int)
	for index, cancelHeight := range storage.CancelCreditHeight {
		if height >= cancelHeight+ChangeCycle {
			total.Add(total, &storage.CancelCreditValue[index])
		}
	}

	return total
}

//取消抵押周期已经到，取消的币可以加入到account的balance中了
func (trieStore *trieStakeStore) CancelCreditToBalance(addr *crypto.CommonAddress, height uint64) (*big.Int, error) {
	storage, _ := trieStore.GetStakeStorage(addr)
	if storage == nil {
		return &big.Int{}, nil
	}

	total := new(big.Int)
	for index, cancelHeight := range storage.CancelCreditHeight {
		if height >= cancelHeight+ChangeCycle {
			total.Add(total, &storage.CancelCreditValue[index])
			storage.CancelCreditHeight = append(storage.CancelCreditHeight[0:index], storage.CancelCreditHeight[index+1:]...)
			storage.CancelCreditValue = append(storage.CancelCreditValue[0:index], storage.CancelCreditValue[index+1:]...)
		}
	}

	err := trieStore.PutStakeStorage(addr, storage)
	if err != nil {
		return &big.Int{}, nil
	}
	return total, nil
}

//获取到候选人所有的质押金
func (trieStore *trieStakeStore) GetCreditCount(addr *crypto.CommonAddress) *big.Int {
	storage, _ := trieStore.GetStakeStorage(addr)
	if storage == nil {
		return &big.Int{}
	}

	total := new(big.Int)
	for _, value := range storage.ReceivedCreditValue {
		total.Add(total, &value)
	}

	return total
}

func (trieStore *trieStakeStore) GetCreditDetails(addr *crypto.CommonAddress) map[crypto.CommonAddress]big.Int {
	m := make(map[crypto.CommonAddress]big.Int)
	storage, _ := trieStore.GetStakeStorage(addr)
	if storage == nil {
		return nil
	}

	for index, addr := range storage.ReceivedCreditAddr {
		m[addr] = storage.ReceivedCreditValue[index]
	}

	return m
}

func (trieStore *trieStakeStore) CandidateCredit(addresses *crypto.CommonAddress, addBalance *big.Int, data []byte) error {
	storage, _ := trieStore.GetStakeStorage(addresses)
	if storage == nil {
		storage = &types.StakeStorage{}
	}

	update := false

	if addBalance != nil {
		update = true
		totalBalance := *addBalance
		if len(storage.ReceivedCreditValue) == 0 {
			storage.ReceivedCreditValue = make([]big.Int, 0)
			storage.ReceivedCreditAddr = make([]crypto.CommonAddress, 0)

			storage.ReceivedCreditAddr = append(storage.ReceivedCreditAddr, *addresses)
			storage.ReceivedCreditValue = append(storage.ReceivedCreditValue, totalBalance)
		} else {
			found := false
			for index, addr := range storage.ReceivedCreditAddr {
				if addr.String() == addresses.String() {
					totalBalance.Add(&storage.ReceivedCreditValue[index], &totalBalance)
					storage.ReceivedCreditValue[index] = totalBalance
					found = true
					break
				}
			}

			if !found {
				storage.ReceivedCreditAddr = append(storage.ReceivedCreditAddr, *addresses)
				storage.ReceivedCreditValue = append(storage.ReceivedCreditValue, totalBalance)
			}
		}

		//投给自己，而且数量足够大
		if totalBalance.Cmp(new(big.Int).Mul(new(big.Int).SetUint64(registerPledgeLimit), new(big.Int).SetUint64(drepUnit))) >= 0 {
			trieStore.AddCandidateAddr(addresses)
		}
	}

	if len(data) > 0 {
		update = true
		storage.CandidateData = data
	}

	if update {
		return trieStore.PutStakeStorage(addresses, storage)
	} else {
		return nil
	}
}

//可以全部取消质押的币；也可以只取消一部分质押的币，当质押的币不满足最低候选要求，则会被撤销候选人地址列表
func (trieStore *trieStakeStore) CancelCandidateCredit(fromAddr *crypto.CommonAddress, cancelBalance *big.Int, height uint64) error {
	//找到币被抵押到的stakeStorage;减去取消的值
	storage, _ := trieStore.GetStakeStorage(fromAddr)
	if storage == nil {
		storage = &types.StakeStorage{}
	}

	if cancelBalance == nil {
		return nil
	}

	if len(storage.ReceivedCreditValue) == 0 {
		return fmt.Errorf("not exist vote credit")
	} else {
		found := false
		for index, addr := range storage.ReceivedCreditAddr {
			if addr.String() == fromAddr.String() {
				found = true
				voteCredit := storage.ReceivedCreditValue[index]
				resultBalance := new(big.Int)
				retCmp := voteCredit.Cmp(cancelBalance)
				if retCmp > 0 {
					storage.ReceivedCreditValue[index] = *resultBalance.Sub(&voteCredit, cancelBalance)
				} else if retCmp == 0 {
					storage.ReceivedCreditAddr = append(storage.ReceivedCreditAddr[0:index], storage.ReceivedCreditAddr[index+1:]...)
					storage.ReceivedCreditValue = append(storage.ReceivedCreditValue[0:index], storage.ReceivedCreditValue[index+1:]...)
				} else {
					return fmt.Errorf("vote credit not enough")
				}

				if resultBalance.Cmp(new(big.Int).Mul(new(big.Int).SetUint64(registerPledgeLimit), new(big.Int).SetUint64(drepUnit))) < 0 {
					trieStore.DelCandidateAddr(fromAddr)
					storage.CandidateData = []byte{}
				}
			}
		}

		if !found {
			return fmt.Errorf("not exist vote credit")
		}
	}

	//存储临时被退回的币
	if len(storage.CancelCreditValue) == 0 {
		storage.CancelCreditValue = make([]big.Int, 0)
		storage.CancelCreditValue[height] = *cancelBalance
	} else {
		found := false
		for index, vh := range storage.CancelCreditHeight {
			if vh == height { //一个块中多笔退款交易
				found = true
				storage.CancelCreditValue[index].Add(&storage.CancelCreditValue[index], cancelBalance)
			}
		}
		if !found {
			storage.CancelCreditHeight = append(storage.CancelCreditHeight, height)
			storage.CancelCreditValue = append(storage.CancelCreditValue, *cancelBalance)
		}
	}

	return trieStore.PutStakeStorage(fromAddr, storage)
}
