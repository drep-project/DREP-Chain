package chain

import (
	"encoding/json"
	"fmt"
	"github.com/drep-project/DREP-Chain/chain/store"
	"github.com/drep-project/DREP-Chain/common"

	//"github.com/drep-project/DREP-Chain/common"
	"github.com/drep-project/DREP-Chain/common/hexutil"
	"github.com/drep-project/DREP-Chain/common/trie"
	"github.com/drep-project/DREP-Chain/crypto"
	"github.com/drep-project/DREP-Chain/crypto/sha3"
	"github.com/drep-project/DREP-Chain/database/dbinterface"
	"github.com/drep-project/DREP-Chain/types"
	"github.com/drep-project/binary"
	"math/big"
)

/*
name: 链接口
usage: 用于获取区块信息
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
 usage: 用于获取区块信息
 params:
	1. height  usage: 当前区块高度
 return: 区块明细信息
 example: curl http://localhost:15645 -X POST --data '{"jsonrpc":"2.0","method":"chain_getBlock","params":[1], "id": 3}' -H "Content-Type:application/json"
 response:
{
  "jsonrpc": "2.0",
  "id": 3,
  "result": {
    "Hash": "0xcfa283a5b591da5a15971bf62fffae87e649bcf749776f4c83ffe50e65920f8e",
    "ChainId": "00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
    "Version": 1,
    "PreviousHash": "0x1717b4b9f740cebeb2659886122a29c0876ed906dd05370319fee4ecf219b1e9",
    "GasLimit": 180000000,
    "GasUsed": 0,
    "Height": 1,
    "Timestamp": 1559272779,
    "StateRoot": "0xd7bd5b3af4f2f1fb3d484743052c2e911f9fb7b04131660912244347508f16a9",
    "TxRoot": "0x",
    "LeaderAddress": "0x0374bf9c8ea268b5548686685dda4a74fc95903ca7c440e5b187a718b595c1f374",
    "MinorAddresses": [
      "0x0374bf9c8ea268b5548686685dda4a74fc95903ca7c440e5b187a718b595c1f374",
      "0x02f11cfd138eaaaba5f8c0a7f1f2791bdabd0b0c404734dceac820aa9b683bfb1a",
      "0x03949aad279a32536ce20f0957c9c6ba592532ea70e5f174332bed4c94382354e3",
      "0x0263bc5628fa7033727d14b5d6714ac7d6a5d34bc5db994a896f54499f12db9b0b"
    ],
    "Txs": [

    ]
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
 usage: 用于获取当前最高区块
 params:
	1. 无
 return: 当前最高区块高度数值
 example: curl http://localhost:15645 -X POST --data '{"jsonrpc":"2.0","method":"chain_getMaxHeight","params":[], "id": 3}' -H "Content-Type:application/json"
 response:
   {"jsonrpc":"2.0","id":3,"result":193005}
*/
func (chain *ChainApi) GetMaxHeight() uint64 {
	return chain.chainView.Tip().Height
}

/*
 name: getBalance
 usage: 查询地址余额
 params:
	1. 待查询地址
 return: 地址中的账号余额
 example: curl http://localhost:15645 -X POST --data '{"jsonrpc":"2.0","method":"chain_getBalance","params":["0x8a8e541ddd1272d53729164c70197221a3c27486"], "id": 3}' -H "Content-Type:application/json"
 response:
   {"jsonrpc":"2.0","id":3,"result":9987999999999984000000}
*/
func (chain *ChainApi) GetBalance(addr crypto.CommonAddress) *big.Int {
	trieQuery, _ := NewTrieQuery(chain.store, chain.chainView.Tip().StateRoot)
	return trieQuery.GetBalance(&addr)
}

/*
 name: getNonce
 usage: 查询地址在链上的nonce
 params:
	1. 待查询地址
 return: 链上nonce
 example: curl http://localhost:15645 -X POST --data '{"jsonrpc":"2.0","method":"chain_getNonce","params":["0x8a8e541ddd1272d53729164c70197221a3c27486"], "id": 3}' -H "Content-Type:application/json"
 response:
   {"jsonrpc":"2.0","id":3,"result":0}
*/
func (chain *ChainApi) GetNonce(addr crypto.CommonAddress) uint64 {
	trieQuery, _ := NewTrieQuery(chain.store, chain.chainView.Tip().StateRoot)
	return trieQuery.GetNonce(&addr)
}

/*
 name: getNonce
 usage: 查询地址的名誉值
 params:
	1. 待查询地址
 return: 地址对应的名誉值
 example: curl http://localhost:15645 -X POST --data '{"jsonrpc":"2.0","method":"chain_getReputation","params":["0x8a8e541ddd1272d53729164c70197221a3c27486"], "id": 3}' -H "Content-Type:application/json"
 response:
   {"jsonrpc":"2.0","id":3,"result":1}
*/
func (chain *ChainApi) GetReputation(addr crypto.CommonAddress) *big.Int {
	trieQuery, _ := NewTrieQuery(chain.store, chain.chainView.Tip().StateRoot)
	return trieQuery.GetReputation(&addr)
}

/*
 name: getTransactionByBlockHeightAndIndex
 usage: 获取区块中特定序列的交易
 params:
	1. 区块高度
    2. 交易序列
 return: 交易信息
 example: curl http://localhost:15645 -X POST --data '{"jsonrpc":"2.0","method":"chain_getTransactionByBlockHeightAndIndex","params":[10000,1], "id": 3}' -H "Content-Type:application/json"
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
 usage: 根据地址获取地址对应的别名
 params:
	1. 待查询地址
 return: 地址别名
 example: curl http://localhost:15645 -X POST --data '{"jsonrpc":"2.0","method":"chain_getAliasByAddress","params":["0x8a8e541ddd1272d53729164c70197221a3c27486"], "id": 3}' -H "Content-Type:application/json"
 response:
	{"jsonrpc":"2.0","id":3,"result":"tom"}
*/
func (chain *ChainApi) GetAliasByAddress(addr *crypto.CommonAddress) string {
	trieQuery, _ := NewTrieQuery(chain.store, chain.chainView.Tip().StateRoot)
	return trieQuery.GetStorageAlias(addr)
}

/*
 name: getAddressByAlias
 usage: 根据别名获取别名对应的地址
 params:
	1. 待查询地别名
 return: 别名对应的地址
 example: curl http://localhost:15645 -X POST --data '{"jsonrpc":"2.0","method":"chain_getAddressByAlias","params":["tom"], "id": 3}' -H "Content-Type:application/json"
 response:
   {"jsonrpc":"2.0","id":3,"result":"0x8a8e541ddd1272d53729164c70197221a3c27486"}
*/
func (chain *ChainApi) GetAddressByAlias(alias string) (*crypto.CommonAddress, error) {
	trieQuery, _ := NewTrieQuery(chain.store, chain.chainView.Tip().StateRoot)
	return trieQuery.AliasGet(alias)
}

/*
 name: getReceipt
 usage: 根据txhash获取receipt信息
 params:
	1. txhash
 return: receipt
 example: curl http://localhost:15645 -X POST --data '{"jsonrpc":"2.0","method":"chain_getReceipt","params":["0x7d9dd32ca192e765ff2abd7c5f8931cc3f77f8f47d2d52170c7804c2ca2c5dd9"], "id": 3}' -H "Content-Type:application/json"
 response:
   {"jsonrpc":"2.0","id":3,"result":""}
*/
func (chain *ChainApi) GetReceipt(txHash crypto.Hash) *types.Receipt {
	return chain.dbQuery.GetReceipt(txHash)
}

/*
 name: getLogs
 usage: 根据txhash获取交易log信息
 params:
	1. txhash
 return: []log
 example: curl http://localhost:15645 -X POST --data '{"jsonrpc":"2.0","method":"chain_getLogs","params":["0x7d9dd32ca192e765ff2abd7c5f8931cc3f77f8f47d2d52170c7804c2ca2c5dd9"], "id": 3}' -H "Content-Type:application/json"
 response:
   {"jsonrpc":"2.0","id":3,"result":""}
*/
func (chain *ChainApi) GetLogs(txHash crypto.Hash) []*types.Log {
	//return chain.chainService.chainStore.GetLogs(txHash)
	rt := chain.dbQuery.GetReceipt(txHash)
	if rt != nil {
		return rt.Logs
	}

	return nil
}

/*
 name: getByteCode
 usage: 根据地址获取bytecode
 params:
	1. 地址
 return: bytecode
 example: curl http://localhost:15645 -X POST --data '{"jsonrpc":"2.0","method":"chain_getByteCode","params":["0x8a8e541ddd1272d53729164c70197221a3c27486"], "id": 3}' -H "Content-Type:application/json"
 response:
   {"jsonrpc":"2.0","id":3,"result":"0x00"}
*/
func (chain *ChainApi) GetByteCode(addr *crypto.CommonAddress) hexutil.Bytes {
	trieQuery, _ := NewTrieQuery(chain.store, chain.chainView.Tip().StateRoot)
	return trieQuery.GetByteCode(addr)
}

/*
 name: getVoteCreditDetails
 usage: 根据地址获取stake 所有细节信息
 params:
	1. 地址
 return: bytecode
 example: curl http://localhost:15645 -X POST --data '{"jsonrpc":"2.0","method":"chain_getCreditDetails","params":["0x8a8e541ddd1272d53729164c70197221a3c27486"], "id": 3}' -H "Content-Type:application/json"
 response:
   {"jsonrpc":"2.0","id":3,"result":"[{\"Addr\":\"0xd05d5f324ada3c418e14cd6b497f2f36d60ba607\",\"HeghtValues\":[{\"CreditHeight\":1329,\"CreditValue\":\"0x11135\"}]}]"}
*/
func (chain *ChainApi) GetCreditDetails(addr *crypto.CommonAddress) string {
	trieQuery, _ := NewTrieQuery(chain.store, chain.chainView.Tip().StateRoot)
	return trieQuery.GetVoteCreditDetails(addr)
}

/*
 name: GetCancelCreditDetails
 usage: 获取所有退票请求的细节
 params:
	1. 地址
 return: bytecode
 example: curl http://localhost:15645 -X POST --data '{"jsonrpc":"2.0","method":"chain_getCancelCreditDetails","params":["0x8a8e541ddd1272d53729164c70197221a3c27486"], "id": 3}' -H "Content-Type:application/json"
 response:
   {"jsonrpc":"2.0","id":3,"result":"{\"0x300fc5a14e578be28c64627c0e7e321771c58cd4\":\"0x3641100\"}"}
*/
func (chain *ChainApi) GetCancelCreditDetails(addr *crypto.CommonAddress) string {
	trieQuery, _ := NewTrieQuery(chain.store, chain.chainView.Tip().StateRoot)
	return trieQuery.GetCancelCreditDetails(addr)
}

/*
 name: GetCandidateAddrs
 usage: 获取所有候选节点地址和对应的信任值
 params:
	1. 地址
 return:  []
 example: curl http://localhost:15645 -X POST --data '{"jsonrpc":"2.0","method":"chain_getCandidateAddrs","params":[""], "id": 3}' -H "Content-Type:application/json"
 response:
   {"jsonrpc":"2.0","id":3,"result":"{\"0x300fc5a14e578be28c64627c0e7e321771c58cd4\":\"0x3641100\"}"}
*/
func (chain *ChainApi) GetCandidateAddrs() string {
	trieQuery, _ := NewTrieQuery(chain.store, chain.chainView.Tip().StateRoot)
	return trieQuery.GetCandidateAddrs()
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
		Addr  *crypto.CommonAddress
		Cridt *common.Big
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

		ac = append(ac, AddrAndCrit{Addr: &addr, Cridt: cb})
	}

	b, err := json.Marshal(ac)
	fmt.Println(string(b), err)
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
