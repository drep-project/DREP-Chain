package database

import (
    "BlockChainTest/bean"
    "math/big"
    "BlockChainTest/mycrypto"
    "strconv"
    "encoding/json"
    "BlockChainTest/accounts"
    "encoding/hex"
)

var (
    db = NewDatabase()
)
func GetBlock(height int64) *bean.Block {
    key := mycrypto.Hash256([]byte("block_" + strconv.FormatInt(height, 10)))
    value := db.get(key)
    block, _ := bean.UnmarshalBlock(value)
    return block
}

func GetBlocksFrom(start, size int64) []*bean.Block {
    var (
        currentBlock *bean.Block
        height = start
        blocks = make([]*bean.Block, 0)
    )
    for currentBlock != nil && (height < start + size || size == -1)  {
        currentBlock = GetBlock(start)
        if currentBlock != nil {
            blocks = append(blocks, currentBlock)
        }
        height += 1
    }
    return blocks
}

func GetAllBlocks() []*bean.Block {
    return GetBlocksFrom(int64(0), int64(-1))
}

func GetHighestBlock() *bean.Block {
    maxHeight := GetMaxHeight()
    return GetBlock(maxHeight)
}

//TODO cannot sync
func PutBlock(block *bean.Block) error {
    key := mycrypto.Hash256([]byte("block_" + strconv.FormatInt(block.Header.Height, 10)))
    value, _ := bean.MarshalBlock(block)
    return db.put(key, value)
}

func GetMaxHeight() int64 {
    key := mycrypto.Hash256([]byte("max_height"))
    if value := db.get(key); value == nil {
        return -1
    } else {
        return new(big.Int).SetBytes(value).Int64()
    }
}

func PutMaxHeight(height int64) error {
    key := mycrypto.Hash256([]byte("max_height"))
    value := new(big.Int).SetInt64(height).Bytes()
    err := db.put(key, value)
    if err != nil {
        return err
    }
    return nil
}

func GetStorageOutsideTransaction(addr accounts.CommonAddress, chainId int64) *accounts.Storage {
    key := mycrypto.Hash256([]byte("storage_" + addr.Hex() + strconv.FormatInt(chainId, 10)))
    value, err := db.Load(key)
    if err != nil {
        return &accounts.Storage{}
    }
    storage := &accounts.Storage{}
    err = json.Unmarshal(value, storage)
    if err != nil {
        return &accounts.Storage{}
    }
    return storage
}

func GetAccountOutsideTransaction(addr bean.CommonAddress) *bean.Account {
    key := mycrypto.Hash256([]byte("account_" + addr.Hex()))
    if value := db.get(key); value != nil {
        account, _ := bean.UnmarshalAccount(value)
        return account
    } else {
        account := &bean.Account{Addr: addr, Nonce: 0, Balance: big.NewInt(0)}
        PutAccountOutsideTransaction(account)
        return account
    }
}

func PutStorage(addr accounts.CommonAddress, chainId int64, storage *accounts.Storage) error {
    db := GetDatabase()
    key := mycrypto.Hash256([]byte("storage_" + addr.Hex() + strconv.FormatInt(chainId, 10)))
    value, err := json.Marshal(storage)
    if err != nil {
        return err
    }
    err = db.Store(key, value)
    if err != nil {
        return err
    }
    return nil
}
func GetAccountInsideTransaction(t *Transaction, addr bean.CommonAddress) *bean.Account {
    key := mycrypto.Hash256([]byte("account_" + addr.Hex()))
    if value := t.Get(key); value != nil {
        account, _ := bean.UnmarshalAccount(value)
        return account
    } else {
        account := &bean.Account{Addr:addr, Nonce:0, Balance:big.NewInt(0)}
        PutAccountInsideTransaction(t, account)
        return account
    }
}

func PutAccountOutsideTransaction(account *bean.Account) error {
    key := mycrypto.Hash256([]byte("account_" + account.Addr.Hex()))
    value, err := bean.MarshalAccount(account)
    err = db.put(key, value)
    if err != nil {
        return err
    }
    return nil
}

func GetBalance(addr accounts.CommonAddress, chainId int64) *big.Int {
    storage := GetStorage(addr, chainId)
    balance := storage.Balance
    if balance == nil {
        balance = new(big.Int)
    }
    return balance
}

func PutBalance(addr accounts.CommonAddress, chainId int64, balance *big.Int) error {
    storage := GetStorage(addr, chainId)
    storage.Balance = balance
    return PutStorage(addr, chainId, storage)
}

func GetNonce(addr accounts.CommonAddress, chainId int64) int64 {
    storage := GetStorage(addr, chainId)
    return storage.Nonce
}

func PutNonce(addr accounts.CommonAddress, chainId int64, nonce int64) error {
    storage := GetStorage(addr, chainId)
    storage.Nonce = nonce
    return PutStorage(addr, chainId, storage)
}

func PutLogs(txHash []byte, logs []*bean.Log) error {
    db := GetDatabase()
    key := mycrypto.Hash256([]byte("logs_" + hex.EncodeToString(txHash)))
    value, err := json.Marshal(logs)
    if err != nil {
        return err
    }
    return db.Store(key, value)
}

func GetLogs(txHash []byte) []*bean.Log {
    db := GetDatabase()
    key := mycrypto.Hash256([]byte("logs_" + hex.EncodeToString(txHash)))
    value, err := db.Load(key)
    if err != nil {
        return make([]*bean.Log, 0)
    }
    var logs []*bean.Log
    err = json.Unmarshal(value, logs)
    if err != nil {
        return make([]*bean.Log, 0)
    }
    return logs
}

func AddLog(log *bean.Log) error {
    logs := GetLogs(log.TxHash)
    logs = append(logs, log)
    return PutLogs(log.TxHash, logs)
}

func PutNodes(nodes map[string] *accounts.Node) error {
    db := GetDatabase()
    key := mycrypto.Hash256([]byte("nodes"))
    value, err := json.Marshal(nodes)
    if err != nil {
        return err
    }
    return db.Store(key, value)
}

func GetNodes() map[string] *accounts.Node {
    db := GetDatabase()
    key := mycrypto.Hash256([]byte("nodes"))
    value, err := db.Load(key)
    if err != nil {
        return make(map[string] *accounts.Node)
    }
    var nodes map[string] *accounts.Node
    err = json.Unmarshal(value, &nodes)
    if err != nil {
        return make(map[string] *accounts.Node)
    }
    return nodes
}

func AddNode(node *accounts.Node) error {
    nodes := GetNodes()
    nodes[node.Address().Hex()] = node
    return PutNodes(nodes)
}

func GetMostRecentBlocks(n int64) []*bean.Block {
    height := GetMaxHeight()
    if height == -1 {
        return nil
    }
    return GetBlocksFrom(height - n, n)
}

func PutAccountInsideTransaction(t *Transaction, account *bean.Account) {
    key := mycrypto.Hash256([]byte("account_" + account.Addr.Hex()))
    value, _ := bean.MarshalAccount(account)
    t.Put(key, value)
}

func GetBalanceOutsideTransaction(addr bean.CommonAddress) *big.Int {
    account := GetAccountOutsideTransaction(addr)
    return account.Balance
}

func GetBalanceInsideTransaction(t *Transaction, addr bean.CommonAddress) *big.Int {
    account := GetAccountInsideTransaction(t, addr)
    return account.Balance
}

func GetNonceOutsideTransaction(addr bean.CommonAddress) int64 {
    account := GetAccountOutsideTransaction(addr)
    return account.Nonce
}

func GetNonceInsideTransaction(t *Transaction, addr bean.CommonAddress) int64 {
    account := GetAccountInsideTransaction(t, addr)
    return account.Nonce
}

func PutNonceOutsideTransaction(addr bean.CommonAddress, nonce int64) error {
    account := GetAccountOutsideTransaction(addr)
    account.Nonce = nonce
    return PutAccountOutsideTransaction(account)
}

func PutNonceInsideTransaction(t *Transaction, addr bean.CommonAddress, nonce int64) {
    account := GetAccountInsideTransaction(t, addr)
    account.Nonce = nonce
    PutAccountInsideTransaction(t, account)
}
