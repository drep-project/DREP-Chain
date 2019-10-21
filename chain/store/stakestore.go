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
	CandidateAddrs = "CandidateAddrs" //参与竞选出块节点的地址集合
	StakeStorage   = "StakeStorage"   //以地址作为KEY,存储stake相关内容

	registerPledgeLimit uint64 = 1000000    //候选节点需要抵押币的总数
	drepUnit            uint64 = 1000000000 //drep币最小单位

	ChangeCycle = 100 //出块节点Change cycle

	interestRate = 1000000000 //每个存储高度，奖励的利率
)

type stakeStoreInterface interface {
	Get(key []byte) ([]byte, error)
	Put(key []byte, value []byte) error
	Delete(key []byte) error
}

type trieStakeStore struct {
	store *StoreDB
}

func NewStakeStorage(store *StoreDB) *trieStakeStore {
	return &trieStakeStore{
		store: store,
	}
}

func (trieStore *trieStakeStore) getStakeStorage(addr *crypto.CommonAddress) (*types.StakeStorage, error) {
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

func (trieStore *trieStakeStore) putStakeStorage(addr *crypto.CommonAddress, storage *types.StakeStorage) error {
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
		addrs = append(addrs, *addr)

	} else { //del
		if len(addrs) == 0 {
			return nil
		} else {
			for index,temAddr := range addrs{
				if temAddr.String() == addr.String() {
					addrs = append(addrs[0:index], addrs[index+1:]...)
				}
			}
		}
	}

	addrsBuf, err := binary.Marshal(addrs)
	if err == nil {
		trieStore.store.Put([]byte(CandidateAddrs), addrsBuf)
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
	storage, _ := trieStore.getStakeStorage(addr)
	if storage == nil {
		return []byte{}, errors.New("addr stake no exist")
	}

	return storage.CandidateData, nil
}

func (trieStore *trieStakeStore) GetCandidateAddrs() ([]crypto.CommonAddress, error) {
	var addrsBuf []byte
	var err error
	key := []byte(CandidateAddrs)
	addrs := []crypto.CommonAddress{}
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

func (trieStore *trieStakeStore) VoteCredit(fromAddr, toAddr *crypto.CommonAddress, addBalance *big.Int, height uint64) error {
	if toAddr == nil || fromAddr == nil {
		return errors.New("addr cannot equal nil")
	}

	if fromAddr.String() == toAddr.String() {
		return errors.New("from euqal to addr")
	}

	if addBalance == nil {
		return errors.New("vote credit value == 0")
	}

	storage, _ := trieStore.getStakeStorage(toAddr)
	if storage == nil {
		storage = &types.StakeStorage{}
	}

	totalBalance := *addBalance
	if len(storage.RC) == 0 {
		storage.RC = make([]types.ReceivedCredit, 0, 1)
	}

	hv := types.HeightValue{height, totalBalance}

	found := false
	for index, rc := range storage.RC {
		if rc.Addr.String() == fromAddr.String() {
			found = true

			storage.RC[index].Hv = append(storage.RC[index].Hv, hv)
			break
		}
	}

	if !found {
		rc := types.ReceivedCredit{Addr: *fromAddr, Hv: make([]types.HeightValue, 0, 1)}
		rc.Hv = append(rc.Hv, hv)
		storage.RC = append(storage.RC, rc)
	}

	return trieStore.putStakeStorage(toAddr, storage)
}

func (trieStore *trieStakeStore) cancelCredit(fromAddr, toAddr *crypto.CommonAddress, cancelBalance *big.Int, height uint64) error {
	getInterst := func(startHeight, endHeight uint64, value *big.Int) *big.Int {
		diff := new(big.Int).SetUint64(height - startHeight + ChangeCycle)
		diff.Mul(diff, value)
		return diff.Sub(diff, new(big.Int).SetUint64(interestRate))
	}

	//找到币被抵押到的stakeStorage;减去取消的值
	storage, _ := trieStore.getStakeStorage(toAddr)
	if storage == nil {
		storage = &types.StakeStorage{}
	}
	if len(storage.RC) == 0 {
		return fmt.Errorf("not exist vote credit")
	} else {
		found := false
		for index, rc := range storage.RC {
			if rc.Addr.String() == fromAddr.String() {
				found = true
				voteCredit := new(big.Int)
				for _, vc := range rc.Hv {
					voteCredit.Add(voteCredit, &vc.CreditValue)
				}

				if voteCredit.Cmp(cancelBalance) >= 0 {
					for hvIndex, vc := range rc.Hv {
						if voteCredit.Cmp(&vc.CreditValue) >= 0 {
							interest := getInterst(vc.CreditHeight, height+ChangeCycle, &vc.CreditValue)
							cancelBalance.Add(cancelBalance, interest)
							voteCredit.Sub(voteCredit, &vc.CreditValue)
							rc.Hv = rc.Hv[1:]

							if voteCredit.Cmp(new(big.Int).SetUint64(0)) == 0 {
								break
							}
						} else {
							interest := getInterst(vc.CreditHeight, height+ChangeCycle, &vc.CreditValue)
							cancelBalance.Add(cancelBalance, interest)
							rc.Hv[hvIndex].CreditValue = *vc.CreditValue.Sub(voteCredit, &vc.CreditValue)
							break
						}
					}
					if len(rc.Hv) == 0 {
						storage.RC = append(storage.RC[0:index], storage.RC[index+1:]...)
					} else {
						storage.RC[index] = rc
					}
				} else {
					return fmt.Errorf("vote credit not enough")
				}
			}
		}

		if !found {
			return fmt.Errorf("not exist vote credit")
		}
	}

	err := trieStore.putStakeStorage(toAddr, storage)
	if err != nil {
		return err
	}

	if fromAddr.String() != toAddr.String() {
		//目的stakeStorage；存储临时被退回的币,给币所属地址storage
		storage, _ = trieStore.getStakeStorage(fromAddr)
		if storage == nil {
			storage = &types.StakeStorage{}
		}
	}

	if len(storage.CC) == 0 {
		storage.CC = make([]types.CancelCredit, 0, 1)
		cc := types.CancelCredit{CancelCreditHeight: height, CancelCreditValue: make([]big.Int, 0, 1)}
		cc.CancelCreditValue = append(cc.CancelCreditValue, *cancelBalance)
		storage.CC = append(storage.CC, cc)

	} else {
		found := false
		for index, cc := range storage.CC {
			if cc.CancelCreditHeight == height {
				found = true
				storage.CC[index].CancelCreditValue = append(storage.CC[index].CancelCreditValue, *cancelBalance)
				break
			}
		}

		if !found {
			cc := types.CancelCredit{CancelCreditHeight: height, CancelCreditValue: make([]big.Int, 0, 1)}
			cc.CancelCreditValue = append(cc.CancelCreditValue, *cancelBalance)
			storage.CC = append(storage.CC, cc)
		}
	}

	return trieStore.putStakeStorage(fromAddr, storage)
}

func (trieStore *trieStakeStore) CancelVoteCredit(fromAddr, toAddr *crypto.CommonAddress, cancelBalance *big.Int, height uint64) error {
	if toAddr == nil || fromAddr == nil {
		return errors.New("addr cannot equal nil")
	}

	if fromAddr.String() == toAddr.String() {
		return errors.New("from euqal to addr")
	}

	if cancelBalance == nil {
		return errors.New("cancel credit value == 0")
	}

	return trieStore.cancelCredit(fromAddr, toAddr, cancelBalance, height)
}

//取消抵押周期已经到，取消的币可以加入到account的balance中了
func (trieStore *trieStakeStore) GetCancelCreditForBalance(addr *crypto.CommonAddress, height uint64) *big.Int {
	storage, _ := trieStore.getStakeStorage(addr)
	if storage == nil {
		return &big.Int{}
	}

	total := new(big.Int)
	for _, cc := range storage.CC {
		if height >= cc.CancelCreditHeight+ChangeCycle {
			for _, value := range cc.CancelCreditValue {
				total.Add(total, &value)
			}
			//storage.CC = append(storage.CC[0:index], storage.CC[index+1:]...)
		}
	}

	return total
}

//取消抵押周期已经到，取消的币可以加入到account的balance中了
func (trieStore *trieStakeStore) CancelCreditToBalance(addr *crypto.CommonAddress, height uint64) (*big.Int, error) {
	storage, _ := trieStore.getStakeStorage(addr)
	if storage == nil {
		return &big.Int{}, nil
	}

	total := new(big.Int)
	for index, cc := range storage.CC {
		if height >= cc.CancelCreditHeight+ChangeCycle {
			for _, value := range cc.CancelCreditValue {
				total.Add(total, &value)
			}
			storage.CC = append(storage.CC[0:index], storage.CC[index+1:]...)
		}
	}

	err := trieStore.putStakeStorage(addr, storage)
	if err != nil {
		return &big.Int{}, nil
	}
	return total, nil
}

//获取到候选人所有的质押金
func (trieStore *trieStakeStore) GetCreditCount(addr *crypto.CommonAddress) *big.Int {
	storage, _ := trieStore.getStakeStorage(addr)
	if storage == nil {
		return &big.Int{}
	}

	total := new(big.Int)
	for _, cc := range storage.CC {
		for _, value := range cc.CancelCreditValue {
			total.Add(total, &value)
		}
	}

	return total
}

func (trieStore *trieStakeStore) GetCreditDetails(addr *crypto.CommonAddress) map[crypto.CommonAddress]big.Int {
	m := make(map[crypto.CommonAddress]big.Int)
	storage, _ := trieStore.getStakeStorage(addr)
	if storage == nil {
		return nil
	}

	for _, rc := range storage.RC {
		total := new(big.Int)
		for _, value := range rc.Hv {
			total.Add(total, &value.CreditValue)
		}
		m[rc.Addr] = *total //storage.ReceivedCreditValue[index]
	}

	return m
}

func (trieStore *trieStakeStore) CandidateCredit(addresses *crypto.CommonAddress, addBalance *big.Int, data []byte, height uint64) error {
	storage, _ := trieStore.getStakeStorage(addresses)
	if storage == nil {
		storage = &types.StakeStorage{}
	}

	update := false

	if addBalance != nil {
		hv := types.HeightValue{height, *addBalance}
		update = true

		totalBalance := new(big.Int).Set(addBalance)

		if len(storage.RC) == 0 {
			storage.RC = make([]types.ReceivedCredit, 0, 1)
		}

		found := false
		for index, rc := range storage.RC {
			if rc.Addr.String() == addresses.String() {
				for _, hv := range rc.Hv {
					totalBalance.Add(totalBalance, &hv.CreditValue)
				}

				storage.RC[index].Hv = append(storage.RC[index].Hv, hv)
				found = true
				break
			}
		}

		if !found {
			rc := types.ReceivedCredit{Addr: *addresses, Hv: make([]types.HeightValue, 0, 1)}
			rc.Hv = append(rc.Hv, hv)
			storage.RC = append(storage.RC, rc)
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
		return trieStore.putStakeStorage(addresses, storage)
	} else {
		return nil
	}
}

//可以全部取消质押的币；也可以只取消一部分质押的币，当质押的币不满足最低候选要求，则会被撤销候选人地址列表
func (trieStore *trieStakeStore) CancelCandidateCredit(fromAddr *crypto.CommonAddress, cancelBalance *big.Int, height uint64) error {
	return trieStore.cancelCredit(fromAddr, fromAddr, cancelBalance, height)
}
