package store

import (
	"errors"
	"fmt"
	"github.com/drep-project/DREP-Chain/common"
	"github.com/drep-project/DREP-Chain/params"
	"math/big"

	"github.com/drep-project/DREP-Chain/crypto"
	"github.com/drep-project/DREP-Chain/crypto/sha3"
	"github.com/drep-project/DREP-Chain/types"
	"github.com/drep-project/binary"
)

const (
	//CandidateAddrs Participates in the selection of the address set of block nodes
	CandidateAddrs = "CandidateAddrs"
	//StakeStorage With the address as the KEY, the relevant content is stored
	StakeStorage = "StakeStorage"

	RegisterPledgeLimit uint64 = 50000000     //The candidate node needs the total number of collateral COINS, unit 1drep
	interestRate               = 1000000 * 12 //Each storage height rewards the interest rate

	threeMonthHeight = 1555200 //Less than 3 months out of the block height
	sixMonthHeight   = 3110400 //6 months out of the block height
	oneYearHeight    = 6220800 //12 months out of the block height
)

type trieStakeStore struct {
	store *StoreDB
}

func newStakeStorage(store *StoreDB) *trieStakeStore {
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
	}

	err = binary.Unmarshal(value, storage)
	if err != nil {
		return nil, err
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
		for _, temAddr := range addrs {
			if temAddr.String() == addr.String() {
				return nil
			}
		}
		addrs = append(addrs, *addr)

	} else { //del
		if len(addrs) == 0 {
			return nil
		}

		for index, temAddr := range addrs {
			if temAddr.String() == addr.String() {
				addrs = append(addrs[0:index], addrs[index+1:]...)
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
	if addr == nil {
		return errors.New("add candidate daddr param err")
	}
	return trieStore.UpdateCandidateAddr(addr, true)
}

func (trieStore *trieStakeStore) DelCandidateAddr(addr *crypto.CommonAddress) error {
	return trieStore.UpdateCandidateAddr(addr, false)
}

func (trieStore *trieStakeStore) GetCandidateData(addr *crypto.CommonAddress) ([]byte, error) {
	if addr == nil {
		return nil, errors.New("get candidate daddr param err")
	}
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

	hv := types.HeightValue{height, common.Big(totalBalance)}

	found := false
	for index, rc := range storage.RC {
		if rc.Addr.String() == fromAddr.String() {
			found = true

			storage.RC[index].HeghtValues = append(storage.RC[index].HeghtValues, hv)
			break
		}
	}

	if !found {
		rc := types.ReceivedCredit{Addr: *fromAddr, HeghtValues: make([]types.HeightValue, 0, 1)}
		rc.HeghtValues = append(rc.HeghtValues, hv)
		storage.RC = append(storage.RC, rc)
	}

	return trieStore.putStakeStorage(toAddr, storage)
}

func (trieStore *trieStakeStore) cancelCredit(fromAddr, toAddr *crypto.CommonAddress, cancelBalance *big.Int, height uint64, changeInterval uint64,
	f func(leftCredit *big.Int, storage *types.StakeStorage) (*types.StakeStorage, error)) (*types.CancelCreditDetail, error) {

	interestData := types.CancelCreditDetail{PrincipalData: make([]types.HeightValue, 0, 1)}

	//Find the stakeStorage to which the coin is pledged; Minus the canceled value
	storage, _ := trieStore.getStakeStorage(toAddr)
	if storage == nil {
		return nil, fmt.Errorf("not exist vote credit")
	}

	leftCredit := new(big.Int)

	if len(storage.RC) == 0 {
		return nil, fmt.Errorf("not exist vote credit")
	}

	found := false
	for index, rc := range storage.RC {
		if rc.Addr.String() == fromAddr.String() {
			found = true

			for _, vc := range rc.HeghtValues {
				leftCredit.Add(leftCredit, vc.CreditValue.ToInt())
			}

			cancelBalanceTmp := new(big.Int).Set(cancelBalance)
			if leftCredit.Cmp(cancelBalance) >= 0 {
				leftCredit.Sub(leftCredit, cancelBalance)
				left := 0
				leftHeightValues := make([]types.HeightValue, 0)

				for hvIndex, heightValue := range rc.HeghtValues {
					if cancelBalanceTmp.Cmp(heightValue.CreditValue.ToInt()) >= 0 {

						//interest := getInterst(heightValue.CreditHeight, height+changeInterval, heightValue.CreditValue.ToInt())
						interestData.PrincipalData = append(interestData.PrincipalData, types.HeightValue{heightValue.CreditHeight, heightValue.CreditValue})
						//interestData.IntersetData = append(interestData.IntersetData, types.HeightValue{height + changeInterval, common.Big(*interest)})

						//cancelBalance.Add(cancelBalance, interest)
						cancelBalanceTmp.Sub(cancelBalanceTmp, heightValue.CreditValue.ToInt())

						if cancelBalanceTmp.Cmp(new(big.Int).SetUint64(0)) == 0 {
							leftHeightValues = append(leftHeightValues, rc.HeghtValues[hvIndex+1:]...)
							rc.HeghtValues = leftHeightValues
							break
						}

					} else {

						//interest := getInterst(heightValue.CreditHeight, height+changeInterval, cancelBalance)
						interestData.PrincipalData = append(interestData.PrincipalData, types.HeightValue{heightValue.CreditHeight, common.Big(*cancelBalanceTmp)})
						//interestData.IntersetData = append(interestData.IntersetData, types.HeightValue{height + changeInterval, common.Big(*interest)})
						//cancelBalance.Add(cancelBalance, interest)

						cv := heightValue.CreditValue.ToInt()
						leftHeightValues = append(leftHeightValues, types.HeightValue{heightValue.CreditHeight, common.Big(*cv.Sub(cv, cancelBalanceTmp))})
						leftHeightValues = append(leftHeightValues, rc.HeghtValues[hvIndex+1:]...)
						rc.HeghtValues = leftHeightValues
						left++
						break
					}
				}
				if len(rc.HeghtValues) == 0 {
					storage.RC = append(storage.RC[0:index], storage.RC[index+1:]...)
				} else {
					storage.RC[index] = rc
				}
			} else {
				return nil, fmt.Errorf("vote credit not enough")
			}
			break
		}
	}

	if !found {
		return nil, fmt.Errorf("not exist vote credit")
	}

	if f != nil {
		var err error
		storage, err = f(leftCredit, storage)
		if err != nil {
			return nil, err
		}
	}

	if len(storage.CC) == 0 {
		storage.CC = make([]types.CancelCredit, 0, 1)
	}

	found = false
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

	err := trieStore.putStakeStorage(fromAddr, storage)

	return &interestData, err
}

func (trieStore *trieStakeStore) CancelVoteCredit(fromAddr, toAddr *crypto.CommonAddress, cancelBalance *big.Int, endHeight uint64, changeInterval uint64) (*types.CancelCreditDetail, error) {
	if toAddr == nil || fromAddr == nil {
		return nil, errors.New("addr cannot equal nil")
	}

	if fromAddr.String() == toAddr.String() {
		return nil, errors.New("from euqal to addr")
	}

	if cancelBalance == nil || cancelBalance.Cmp(new(big.Int).SetUint64(0)) <= 0 {
		return nil, fmt.Errorf("cancel credit value(%v) <= 0", cancelBalance)
	}

	return trieStore.cancelCredit(fromAddr, toAddr, cancelBalance, endHeight, changeInterval, func(_ *big.Int, storage *types.StakeStorage) (*types.StakeStorage, error) {
		err := trieStore.putStakeStorage(toAddr, storage)
		if err != nil {
			return nil, err
		}

		//Objective stakeStorage; Store the temporarily returned currency and give it to the address storage where it belongs
		storage, _ = trieStore.getStakeStorage(fromAddr)
		if storage == nil {
			storage = &types.StakeStorage{}
		}

		return storage, nil
	})
}

//The mortgage cancellation cycle has come, and the cancelled currency can be added to the balance of the account
func (trieStore *trieStakeStore) GetCancelCreditForBalance(addr *crypto.CommonAddress, height uint64, changeInterval uint64) *big.Int {
	storage, _ := trieStore.getStakeStorage(addr)
	if storage == nil {
		return &big.Int{}
	}

	total := new(big.Int)
	for _, cc := range storage.CC {
		if height >= cc.CancelCreditHeight+changeInterval {
			for _, value := range cc.CancelCreditValue {
				total.Add(total, &value)
			}
			//storage.CC = append(storage.CC[0:index], storage.CC[index+1:]...)
		}
	}

	return total
}

//The mortgage cancellation cycle has come, and the cancelled currency can be added to the balance of the account
func (trieStore *trieStakeStore) CancelCreditToBalance(addr *crypto.CommonAddress, height uint64, changeInterval uint64) (*big.Int, error) {
	storage, _ := trieStore.getStakeStorage(addr)
	if storage == nil {
		return &big.Int{}, nil
	}

	total := new(big.Int)
	left := 0
	for _, cc := range storage.CC {
		if height >= cc.CancelCreditHeight+changeInterval {
			for _, value := range cc.CancelCreditValue {
				total.Add(total, &value)
			}
		} else {
			storage.CC[left] = cc
			left++
		}
	}
	storage.CC = storage.CC[:left]

	err := trieStore.putStakeStorage(addr, storage)
	if err != nil {
		return &big.Int{}, nil
	}
	return total, nil
}

//Obtain all the pledge money of the candidate
func (trieStore *trieStakeStore) GetCreditCount(addr *crypto.CommonAddress) *big.Int {
	if addr == nil {
		return &big.Int{}
	}
	storage, _ := trieStore.getStakeStorage(addr)
	if storage == nil {
		return &big.Int{}
	}

	total := new(big.Int)
	for _, rc := range storage.RC {
		for _, hv := range rc.HeghtValues {
			total.Add(total, hv.CreditValue.ToInt())
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
		for _, value := range rc.HeghtValues {
			total.Add(total, value.CreditValue.ToInt())
		}
		m[rc.Addr] = *total //storage.ReceivedCreditValue[index]
	}

	return m
}

func (trieStore *trieStakeStore) CandidateCredit(addresses *crypto.CommonAddress, addBalance *big.Int, data []byte, height uint64) error {
	if addresses == nil {
		return errors.New("candidate credit param err")
	}
	storage, _ := trieStore.getStakeStorage(addresses)
	if storage == nil {
		storage = &types.StakeStorage{}
	}

	update := false

	if addBalance != nil {
		hv := types.HeightValue{height, common.Big(*addBalance)}
		update = true
		totalBalance := new(big.Int).Set(addBalance)

		if len(storage.RC) == 0 {
			storage.RC = make([]types.ReceivedCredit, 0, 1)
		}

		found := false
		for index, rc := range storage.RC {
			if rc.Addr.String() == addresses.String() {
				for _, hv := range rc.HeghtValues {
					totalBalance.Add(totalBalance, hv.CreditValue.ToInt())
				}

				storage.RC[index].HeghtValues = append(storage.RC[index].HeghtValues, hv)
				found = true
				break
			}
		}

		if !found {
			rc := types.ReceivedCredit{Addr: *addresses, HeghtValues: make([]types.HeightValue, 0, 1)}
			rc.HeghtValues = append(rc.HeghtValues, hv)
			storage.RC = append(storage.RC, rc)
		}

		//Vote for yourself, and in large enough Numbers
		if totalBalance.Cmp(new(big.Int).Mul(new(big.Int).SetUint64(RegisterPledgeLimit), new(big.Int).SetUint64(params.Coin))) >= 0 {
			trieStore.AddCandidateAddr(addresses)
		}
	}
	candidataDate := &types.CandidateData{}
	err := candidataDate.Unmarshal(data)
	if err != nil {
		return err
	}
	data, err = binary.Marshal(candidataDate)
	if len(data) > 0 {
		update = true
		storage.CandidateData = data
	}

	if update {
		return trieStore.putStakeStorage(addresses, storage)
	}

	return errors.New("candidate credit param err")
}

//All the pledged currencies may be cancelled; Also, only a part of the pledged COINS can be cancelled.
// When the pledged COINS do not meet the minimum candidate requirements, the address list of the candidates will be cancelled
func (trieStore *trieStakeStore) CancelCandidateCredit(fromAddr *crypto.CommonAddress, cancelBalance *big.Int, height uint64, changeInterval uint64) (*types.CancelCreditDetail, error) {
	if fromAddr == nil || cancelBalance == nil {
		return nil, errors.New("cancel candidate credit param err")
	}
	return trieStore.cancelCredit(fromAddr, fromAddr, cancelBalance, height, changeInterval, func(leftCredit *big.Int, storage *types.StakeStorage) (*types.StakeStorage, error) {
		if leftCredit.Cmp(new(big.Int).Mul(new(big.Int).SetUint64(RegisterPledgeLimit), new(big.Int).SetUint64(params.Coin))) < 0 {
			trieStore.DelCandidateAddr(fromAddr)
		}
		return storage, nil
	})
}
