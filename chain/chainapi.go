package chain

import (
	"encoding/json"
	"fmt"

	"github.com/drep-project/DREP-Chain/chain/store"
	"github.com/drep-project/DREP-Chain/common"
	"github.com/drep-project/DREP-Chain/common/hexutil"
	"github.com/drep-project/DREP-Chain/common/trie"
	"github.com/drep-project/DREP-Chain/crypto"
	"github.com/drep-project/DREP-Chain/crypto/sha3"
	"github.com/drep-project/DREP-Chain/database/dbinterface"
	"github.com/drep-project/DREP-Chain/params"
	"github.com/drep-project/DREP-Chain/types"
	"github.com/drep-project/binary"
	"math/big"
)

/*
name: Block chain API
usage: Used to obtain block information
prefix:chain

*/
type ChainApi struct {
	store     dbinterface.KeyValueStore
	chainView *ChainView
	dbQuery   *ChainStore
}

func NewChainApi(store dbinterface.KeyValueStore, chainView *ChainView, dbQuery *ChainStore) *ChainApi {
	return &ChainApi{
		store:     store,
		chainView: chainView,
		dbQuery:   dbQuery,
	}
}

/*
 name: getblock
 usage: Used to obtain block information
 params:
	1. height  usage: Current block height
 return: Block detail information
 example: curl http://localhost:10085 -X POST --data '{"jsonrpc":"2.0","method":"chain_getBlock","params":[1], "id": 3}' -H "Content-Type:application/json"
 response:
{
    "jsonrpc":"2.0",
    "id":3,
    "result":{
        "Header":{
            "ChainId":0,
            "Version":1,
            "PreviousHash":"0x1fbae528a8eed0f09201bfd2c7e52fef66f5f35619e9868cd6d02dabac60e4e6",
            "GasLimit":18000000,
            "GasUsed":0,
            "Height":1,
            "Timestamp":1592365562,
            "StateRoot":"UpMnHA5WmmTxU4T4jFQvpt6bFwigN+fg1Jx0fSD91MA=",
            "TxRoot":null,
            "ReceiptRoot":"0x0000000000000000000000000000000000000000000000000000000000000000",
            "Bloom":"0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"
        },
        "Data":{
            "TxCount":0,
            "TxList":null
        },
        "Proof":{
            "Type":0,
            "Evidence":"MEUCIQDIZnsow/WbAmQ7jJ21EcVxzQkKA33LJfw8anhzkNjBzAIgMexycsYJlYEv0rbPvleoAx1iahzUx6FrMNZhh8uq6Lg="
        }
    }
}
*/
func (chain *ChainApi) GetBlock(height uint64) (*types.Block, error) {
	node := chain.chainView.NodeByHeight(height)
	if node == nil {
		return nil, ErrBlockNotFound
	}
	block, err := chain.dbQuery.GetBlock(node.Hash)
	if err != nil {
		return nil, err
	}
	return block, nil
}

/*
 name: getMaxHeight
 usage: To get the current highest block
 params:
	1. 无
 return: Current maximum block height value
 example: curl http://localhost:10085 -X POST --data '{"jsonrpc":"2.0","method":"chain_getMaxHeight","params":[], "id": 3}' -H "Content-Type:application/json"
 response:
   {"jsonrpc":"2.0","id":3,"result":193005}
*/
func (chain *ChainApi) GetMaxHeight() uint64 {
	return chain.chainView.Tip().Height
}

/*
 name: getBlockGasInfo
 usage: Obtain gas related information
 params:
	1. 无
 return: Gas minimum value and maximum value required by the system; And the maximum gas value that the current block is set to
 example: curl http://localhost:10085 -X POST --data '{"jsonrpc":"2.0","method":"chain_getBlockGasInfo","params":[], "id": 3}' -H "Content-Type:application/json"
 response:
   {"jsonrpc":"2.0","id":3,"result":193005}
*/
func (chain *ChainApi) GetBlockGasInfo() string {
	height := chain.chainView.Tip().Height
	node := chain.chainView.NodeByHeight(height)
	if node == nil {
		return ""
	}
	block, err := chain.dbQuery.GetBlock(node.Hash)
	if err != nil {
		return ""
	}

	type gasInfo struct {
		MinGas          int
		MaxGas          int
		CurrentBlockGas int
	}
	gi := gasInfo{
		MinGas:          int(params.MinGasLimit),
		MaxGas:          int(params.MaxGasLimit),
		CurrentBlockGas: int(block.GasLimit()),
	}

	ret, _ := json.Marshal(&gi)

	return string(ret)
}

/*
 name: getBalance
 usage: Query address balance
 params:
	1. Query address
 return: The account balance in the address
 example: curl http://localhost:10085 -X POST --data '{"jsonrpc":"2.0","method":"chain_getBalance","params":["0x8a8e541ddd1272d53729164c70197221a3c27486"], "id": 3}' -H "Content-Type:application/json"
 response:
   {"jsonrpc":"2.0","id":3,"result":9987999999999984000000}
*/
func (chain *ChainApi) GetBalance(addr crypto.CommonAddress) string {
	store, err := store.TrieStoreFromStore(chain.store, chain.chainView.Tip().StateRoot)
	if err != nil {
		return ""
	}
	big := store.GetBalance(&addr, chain.chainView.tip().Height)

	return big.String()
}

/*
 name: getNonce
 usage: Query the nonce whose address is on the chain
 params:
	1. Query address
 return: nonce
 example: curl http://localhost:10085 -X POST --data '{"jsonrpc":"2.0","method":"chain_getNonce","params":["0x8a8e541ddd1272d53729164c70197221a3c27486"], "id": 3}' -H "Content-Type:application/json"
 response:
   {"jsonrpc":"2.0","id":3,"result":0}
*/
func (chain *ChainApi) GetNonce(addr crypto.CommonAddress) uint64 {
	trieQuery, _ := NewTrieQuery(chain.store, chain.chainView.Tip().StateRoot)
	return trieQuery.GetNonce(&addr)
}

/*
 name: GetReputation
 usage: Query the reputation value of the address
 params:
	1.  Query address
 return: The reputation value corresponding to the address
 example: curl http://localhost:10085 -X POST --data '{"jsonrpc":"2.0","method":"chain_getReputation","params":["0x8a8e541ddd1272d53729164c70197221a3c27486"], "id": 3}' -H "Content-Type:application/json"
 response:
   {"jsonrpc":"2.0","id":3,"result":1}
*/
func (chain *ChainApi) GetReputation(addr crypto.CommonAddress) *big.Int {
	trieQuery, _ := NewTrieQuery(chain.store, chain.chainView.Tip().StateRoot)
	return trieQuery.GetReputation(&addr)
}

/*
 name: getTransactionByBlockHeightAndIndex
 usage: Gets a particular sequence of transactions in a block
 params:
	1. block height
    2. Transaction sequence
 return: transaction
 example: curl http://localhost:10085 -X POST --data '{"jsonrpc":"2.0","method":"chain_getTransactionByBlockHeightAndIndex","params":[10000,1], "id": 3}' -H "Content-Type:application/json"
 response:
   {
  "jsonrpc": "2.0",
  "id": 3,
  "result": {
    "Hash": "0xfa5c34114ff459b4c97e7cd268c507c0ccfcfc89d3ccdcf71e96402f9899d040",
    "From": "0x7923a30bbfbcb998a6534d56b313e68c8e0c594a",
    "Version": 1,
    "Nonce": 15632,
    "Type": 0,
    "To": "0x7923a30bbfbcb998a6534d56b313e68c8e0c594a",
    "ChainId": "00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
    "Amount": "0x111",
    "GasPrice": "0x110",
    "GasLimit": "0x30000",
    "Timestamp": 1559322808,
    "Data": null,
    "Sig": "0x20f25b86c4bf73aa4fa0bcb01e2f5731de3a3917c8861d1ce0574a8d8331aedcf001e678000f6afc95d35a53ef623a2055fce687f85c2fd752dc455ab6db802b1f"
  }
}
*/
func (chain *ChainApi) GetTransactionByBlockHeightAndIndex(height uint64, index int) (*types.Transaction, error) {
	block, err := chain.GetBlock(height)
	if err != nil {
		return nil, err
	}
	if index > int(block.Data.TxCount) {
		return nil, ErrTxIndexOutOfRange
	}
	return block.Data.TxList[index], nil
}

/*
 name: getAliasByAddress
 usage: Gets the alias corresponding to the address according to the address
 params:
	1. address
 return: Address the alias
 example: curl http://localhost:10085 -X POST --data '{"jsonrpc":"2.0","method":"chain_getAliasByAddress","params":["0x8a8e541ddd1272d53729164c70197221a3c27486"], "id": 3}' -H "Content-Type:application/json"
 response:
	{"jsonrpc":"2.0","id":3,"result":"tom"}
*/
func (chain *ChainApi) GetAliasByAddress(addr *crypto.CommonAddress) string {
	trieQuery, _ := NewTrieQuery(chain.store, chain.chainView.Tip().StateRoot)
	return trieQuery.GetStorageAlias(addr)
}

/*
 name: getAddressByAlias
 usage: Gets the address corresponding to the alias based on the alias
 params:
	1. Alias to be queried
 return: The address corresponding to the alias
 example: curl http://localhost:10085 -X POST --data '{"jsonrpc":"2.0","method":"chain_getAddressByAlias","params":["tom"], "id": 3}' -H "Content-Type:application/json"
 response:
   {"jsonrpc":"2.0","id":3,"result":"0x8a8e541ddd1272d53729164c70197221a3c27486"}
*/
func (chain *ChainApi) GetAddressByAlias(alias string) (*crypto.CommonAddress, error) {
	trieQuery, _ := NewTrieQuery(chain.store, chain.chainView.Tip().StateRoot)
	return trieQuery.AliasGet(alias)
}

/*
 name: getReceipt
 usage: Get the receipt information based on txhash
 params:
	1. txhash
 return: receipt
 example: curl http://localhost:10085 -X POST --data '{"jsonrpc":"2.0","method":"chain_getReceipt","params":["0x7d9dd32ca192e765ff2abd7c5f8931cc3f77f8f47d2d52170c7804c2ca2c5dd9"], "id": 3}' -H "Content-Type:application/json"
 response:
   {"jsonrpc":"2.0","id":3,"result":""}
*/
func (chain *ChainApi) GetReceipt(txHash crypto.Hash) *types.Receipt {
	return chain.dbQuery.GetReceipt(txHash)
}

/*
 name: getLogs
 usage: Get the transaction log information based on txhash
 params:
	1. txhash
 return: []log
 example: curl http://localhost:10085 -X POST --data '{"jsonrpc":"2.0","method":"chain_getLogs","params":["0x7d9dd32ca192e765ff2abd7c5f8931cc3f77f8f47d2d52170c7804c2ca2c5dd9"], "id": 3}' -H "Content-Type:application/json"
 response:
   {"jsonrpc":"2.0","id":3,"result":""}
*/
func (chain *ChainApi) GetLogs(txHash crypto.Hash) []*types.Log {
	//return chain.chainService.chainStore.GetLogs(txHash)
	rt := chain.dbQuery.GetReceipt(txHash)
	if rt != nil {
		//for _, log := range rt.Logs {
		//	if log.TxType == types.CancelVoteCreditType || log.TxType == types.CancelCandidateType {
		//		id := types.CancelCreditDetail{}
		//		err := json.Unmarshal(log.Data, &id)
		//		if err == nil {
		//			ids = append(ids, &id)
		//		}
		//	}
		//}

		return rt.Logs
	}

	return nil
}

/*
 name: getCancelCreditDetail
 usage: Get the back pledge or back vote information according to txhash
 params:
	1. txhash
 return: {}
 example: curl http://localhost:10085 -X POST --data '{"jsonrpc":"2.0","method":"chain_getCancelCreditDetail","params":["0x7d9dd32ca192e765ff2abd7c5f8931cc3f77f8f47d2d52170c7804c2ca2c5dd9"], "id": 3}' -H "Content-Type:application/json"
 response:
   {"jsonrpc":"2.0","id":3,"result":""}
*/
func (chain *ChainApi) GetCancelCreditDetail(txHash crypto.Hash) []*types.CancelCreditDetail {
	//return chain.chainService.chainStore.GetLogs(txHash)
	ids := make([]*types.CancelCreditDetail, 0)
	rt := chain.dbQuery.GetReceipt(txHash)
	if rt != nil {
		for _, log := range rt.Logs {
			if log.TxType == types.CancelVoteCreditType || log.TxType == types.CancelCandidateType {
				id := types.CancelCreditDetail{}
				err := json.Unmarshal(log.Data, &id)
				if err == nil {
					ids = append(ids, &id)
				}
			}
		}

		return ids
	}

	return nil
}

/*
 name: getByteCode
 usage: Get bytecode by address
 params:
	1. address
 return: bytecode
 example: curl http://localhost:10085 -X POST --data '{"jsonrpc":"2.0","method":"chain_getByteCode","params":["0x8a8e541ddd1272d53729164c70197221a3c27486"], "id": 3}' -H "Content-Type:application/json"
 response:
   {"jsonrpc":"2.0","id":3,"result":"0x00"}
*/
func (chain *ChainApi) GetByteCode(addr *crypto.CommonAddress) hexutil.Bytes {
	trieQuery, _ := NewTrieQuery(chain.store, chain.chainView.Tip().StateRoot)
	return trieQuery.GetByteCode(addr)
}

/*
 name: getVoteCreditDetails
 usage: Get all the details of the stake according to the address
 params:
	1. address
 return: bytecode
 example: curl http://localhost:10085 -X POST --data '{"jsonrpc":"2.0","method":"chain_getCreditDetails","params":["0x8a8e541ddd1272d53729164c70197221a3c27486"], "id": 3}' -H "Content-Type:application/json"
 response:
   {"jsonrpc":"2.0","id":3,"result":"[{\"Addr\":\"DREPd05d5f324ada3c418e14cd6b497f2f36d60ba607\",\"HeghtValues\":[{\"CreditHeight\":1329,\"CreditValue\":\"0x11135\"}]}]"}
*/
func (chain *ChainApi) GetCreditDetails(addr *crypto.CommonAddress) string {
	trieQuery, _ := NewTrieQuery(chain.store, chain.chainView.Tip().StateRoot)
	return trieQuery.GetVoteCreditDetails(addr)
}

/*
 name: GetCancelCreditDetails
 usage: Get the details of all refund requests
 params:
	1. address
 return: bytecode
 example: curl http://localhost:10085 -X POST --data '{"jsonrpc":"2.0","method":"chain_getCancelCreditDetails","params":["0x8a8e541ddd1272d53729164c70197221a3c27486"], "id": 3}' -H "Content-Type:application/json"
 response:
   {"jsonrpc":"2.0","id":3,"result":"{\"DREP300fc5a14e578be28c64627c0e7e321771c58cd4\":\"0x3641100\"}"}
*/
func (chain *ChainApi) GetCancelCreditDetails(addr *crypto.CommonAddress) string {
	trieQuery, _ := NewTrieQuery(chain.store, chain.chainView.Tip().StateRoot)
	return trieQuery.GetCancelCreditDetails(addr)
}

/*
 name: GetCandidateAddrs
 usage: Gets the addresses of all candidate nodes and the corresponding trust values
 params:
	1. address
 return:  []
 example: curl http://localhost:10085 -X POST --data '{"jsonrpc":"2.0","method":"chain_getCandidateAddrs","params":[""], "id": 3}' -H "Content-Type:application/json"
 response:
   {"jsonrpc":"2.0","id":3,"result":"{\"DREP300fc5a14e578be28c64627c0e7e321771c58cd4\":\"0x3641100\"}"}
*/
func (chain *ChainApi) GetCandidateAddrs() string {
	trieQuery, _ := NewTrieQuery(chain.store, chain.chainView.Tip().StateRoot)
	return trieQuery.GetCandidateAddrs()
}

///*
// name: getInterestRate
// usage: 获取3个月内、3-6个月、6-12个月、12个月以上的利率
// params:
//	无
// return:  年华后三个月利息, 年华后六个月利息, 一年期利息, 一年以上期利息
// example: curl http://localhost:10085 -X POST --data '{"jsonrpc":"2.0","method":"chain_getInterestRate","params":"", "id": 3}' -H "Content-Type:application/json"
// response:
//   {"jsonrpc":"2.0","id":3,"result":"{\"ThreeMonthRate\":4,\"SixMonthRate\":12,\"OneYearRate\":25,\"MoreOneYearRate\":51}"}
//*/
//func (chain *ChainApi) GetInterestRate() (string, error) {
//
//	threeMonth, sixMonth, oneYear, moreOneYear := store.GetInterestRate()
//
//	type InterestRateInfo struct {
//		ThreeMonthRate  uint64
//		SixMonthRate    uint64
//		OneYearRate     uint64
//		MoreOneYearRate uint64
//	}
//
//	iri := InterestRateInfo{
//		ThreeMonthRate:  threeMonth,
//		SixMonthRate:    sixMonth,
//		OneYearRate:     oneYear,
//		MoreOneYearRate: moreOneYear,
//	}
//
//	fmt.Println(threeMonth, sixMonth, oneYear, moreOneYear)
//
//	ret, err := json.Marshal(&iri)
//
//	return string(ret), err
//}

/*
 name: getChangeCycle
 usage: Gets the transition period of the out - of - block node
 params:
	none
 return:  Transition period
 example: curl http://localhost:10085 -X POST --data '{"jsonrpc":"2.0","method":"chain_getChangeCycle","params":"", "id": 3}' -H "Content-Type:application/json"
 response:
   {"jsonrpc":"2.0","id":3,"result":"{100}"}
*/
func (chain *ChainApi) GetChangeCycle() (int, error) {
	store, err := store.TrieStoreFromStore(chain.store, chain.chainView.Tip().StateRoot)
	if err != nil {
		return 0, err
	}

	changeInterval, err := store.GetChangeInterval()
	return int(changeInterval), err
}

func (chain *ChainApi) GetReward() (int, error) {
	return params.Rewards, nil
}

func (chain *ChainApi) GetAvgPrice(height uint64) (*big.Int, error) {
	block, err := chain.GetBlock(height)
	if err != nil {
		return nil, err
	}
	txCount := big.NewInt(int64(block.Data.TxCount))
	gasUsed := &block.Header.GasUsed

	avgPrice := block.Header.GasUsed.Div(gasUsed, txCount)
	return avgPrice, nil

}

type TrieQuery struct {
	dbinterface.KeyValueStore
	trie *trie.SecureTrie
	root []byte
}

func NewTrieQuery(store dbinterface.KeyValueStore, root []byte) (*TrieQuery, error) {
	trieQuery := &TrieQuery{store, nil, root}
	trieDb := trie.NewDatabaseWithCache(store, 0)
	var err error
	trieQuery.trie, err = trie.NewSecure(crypto.Bytes2Hash(root), trieDb)
	if err != nil {
		return nil, err
	}
	return trieQuery, nil
}

func (trieQuery *TrieQuery) Get(key []byte) ([]byte, error) {
	return trieQuery.trie.TryGet(key)
}

func (trieQuery *TrieQuery) GetStorage(addr *crypto.CommonAddress) (types.Storage, error) {
	key := sha3.Keccak256([]byte(store.AddressStorage + addr.Hex()))
	value, err := trieQuery.trie.TryGet(key)
	storage := types.Storage{}
	if value == nil {
		return storage, nil
	} else {
		err = binary.Unmarshal(value, &storage)
		if err != nil {
			return storage, err
		}
	}
	return storage, nil
}

func (trieQuery *TrieQuery) GetStorageAlias(addr *crypto.CommonAddress) string {
	storage, _ := trieQuery.GetStorage(addr)
	return storage.Alias
}

func (trieQuery *TrieQuery) AliasGet(alias string) (*crypto.CommonAddress, error) {
	buf, err := trieQuery.Get([]byte(store.AliasPrefix + alias))
	if err != nil {
		return nil, err
	}
	if buf == nil {
		return nil, fmt.Errorf("alias :%s not set", alias)
	}
	addr := crypto.CommonAddress{}
	addr.SetBytes(buf)
	return &addr, nil
}

func (trieQuery *TrieQuery) AliasExist(alias string) bool {
	_, err := trieQuery.Get([]byte(store.AliasPrefix + alias))
	if err != nil {
		return false
	}
	return true
}

func (trieQuery *TrieQuery) GetBalance(addr *crypto.CommonAddress) *big.Int {
	storage, _ := trieQuery.GetStorage(addr)
	return &storage.Balance
}

func (trieQuery *TrieQuery) GetNonce(addr *crypto.CommonAddress) uint64 {
	storage, _ := trieQuery.GetStorage(addr)
	return storage.Nonce
}

func (trieQuery *TrieQuery) GetByteCode(addr *crypto.CommonAddress) []byte {
	storage, _ := trieQuery.GetStorage(addr)
	return storage.ByteCode
}

func (trieQuery *TrieQuery) GetCodeHash(addr *crypto.CommonAddress) crypto.Hash {
	storage, _ := trieQuery.GetStorage(addr)
	return storage.CodeHash
}

func (trieQuery *TrieQuery) GetReputation(addr *crypto.CommonAddress) *big.Int {
	storage, _ := trieQuery.GetStorage(addr)
	return &storage.Reputation
}

func (trieQuery *TrieQuery) GetVoteCreditDetails(addr *crypto.CommonAddress) string {
	key := sha3.Keccak256([]byte(store.StakeStorage + addr.Hex()))

	storage := &types.StakeStorage{}

	value, err := trieQuery.trie.TryGet(key)
	if err != nil {
		log.Errorf("get storage err:%v", err)
		return ""
	}
	if value == nil {
		return ""
	} else {
		err = binary.Unmarshal(value, storage)
		if err != nil {
			return ""
		}
	}

	if len(storage.RC) == 0 {
		return ""
	}
	b, _ := json.Marshal(storage.RC)
	return string(b)
}

func (trieQuery *TrieQuery) GetCandidateAddrs() string {
	var addrsBuf []byte
	var err error

	key := []byte(store.CandidateAddrs)

	addrs := []crypto.CommonAddress{}
	addrsBuf = trieQuery.trie.Get(key)
	if err != nil {
		log.Errorf("GetCandidateAddrs:%v", err)
		return ""
	}

	if addrsBuf == nil {
		return ""
	}

	err = binary.Unmarshal(addrsBuf, &addrs)
	if err != nil {
		log.Errorf("GetCandidateAddrs, Unmarshal:%v", err)
		return ""
	}

	type AddrAndCrit struct {
		Addr   string
		Credit *common.Big
	}

	ac := make([]AddrAndCrit, 0)
	for _, addr := range addrs {
		addr := addr
		storage := &types.StakeStorage{}
		key := sha3.Keccak256([]byte(store.StakeStorage + addr.Hex()))

		value, _ := trieQuery.trie.TryGet(key)
		if value == nil {
			return ""
		} else {
			err = binary.Unmarshal(value, storage)
			if err != nil {
				return ""
			}
		}

		total := new(big.Int)
		for _, rc := range storage.RC {
			for _, hv := range rc.HeghtValues {
				total.Add(total, hv.CreditValue.ToInt())
			}
		}

		cb := new(common.Big)
		cb.SetMathBig(*total)
		drepAddr := crypto.EthToDrep(&addr)
		ac = append(ac, AddrAndCrit{Addr: drepAddr, Credit: cb})
	}

	b, err := json.Marshal(ac)
	return string(b)
}

func (trieQuery *TrieQuery) GetCancelCreditDetails(addr *crypto.CommonAddress) string {
	key := sha3.Keccak256([]byte(store.StakeStorage + addr.Hex()))

	storage := &types.StakeStorage{}
	value, err := trieQuery.trie.TryGet(key)
	if err != nil {
		log.Errorf("get storage err:%v", err)
		return ""
	}
	if value == nil {
		return ""
	} else {
		err = binary.Unmarshal(value, storage)
		if err != nil {
			return ""
		}
	}

	if len(storage.CC) == 0 {
		return ""
	}

	//for _, rc := range storage.RC {
	//	total := new(big.Int)
	//	for _, value := range rc.HeghtValues {
	//		total.Add(total, &value.CreditValue)
	//	}
	//	m[rc.Addr] = common.Big(*total)
	//}
	b, _ := json.Marshal(storage.CC)
	return string(b)
}
