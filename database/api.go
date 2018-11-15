package database

import (
    "BlockChainTest/bean"
    "math/big"
    "BlockChainTest/mycrypto"
    "strconv"
    "encoding/json"
)

func GetBlock(height int64) *bean.Block {
    db := GetDatabase()
    key := mycrypto.Hash256([]byte("block_" + strconv.FormatInt(height, 10)))
    value, _ := db.Load(key)
    block, _ := bean.UnmarshalBlock(value)
    return block
}

func GetBlocksFrom(start, size int64) []*bean.Block {
    var (
        currentBlock *bean.Block
        err error
        height = start
        blocks = make([]*bean.Block, 0)
    )
    for err == nil && (height < start + size || size == -1)  {
        currentBlock = GetBlock(start)
        if err == nil {
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

func PutBlock(block *bean.Block) error {
    db := GetDatabase()
    key := mycrypto.Hash256([]byte("block_" + strconv.FormatInt(block.Header.Height, 10)))
    value, _ := bean.MarshalBlock(block)
    return db.Store(key, value)
}

func GetMaxHeight() int64 {
    db := GetDatabase()
    key := mycrypto.Hash256([]byte("max_height"))
    value, err := db.Load(key)
    if err != nil {
        return -1
    }
    return new(big.Int).SetBytes(value).Int64()
}

func PutMaxHeight(height int64) error {
    db := GetDatabase()
    key := mycrypto.Hash256([]byte("max_height"))
    value := new(big.Int).SetInt64(height).Bytes()
    err := db.Store(key, value)
    if err != nil {
        return err
    }
    db.Trie.Insert(key, value)
    return nil
}

func GetAccount(addr bean.CommonAddress) *bean.Account {
    db := GetDatabase()
    key := mycrypto.Hash256([]byte("account_" + addr.Hex()))
    if value, err := db.Load(key); err == nil {
        account, _ := bean.UnmarshalAccount(value)
        return account
    } else {
        account := &bean.Account{Addr:addr, Nonce:0, Balance:big.NewInt(0)}
        PutAccount(account)
        return account
    }
}

func PutAccount(account *bean.Account) error {
    db := GetDatabase()
    key := mycrypto.Hash256([]byte("account_" + account.Addr.Hex()))
    value, err := bean.MarshalAccount(account)
    err = db.Store(key, value)
    if err != nil {
        return err
    }
    db.Trie.Insert(key, value)
    return AddAccountsAddress(account)
}

func GetBalance(addr bean.CommonAddress) *big.Int {
    account := GetAccount(addr)
    return account.Balance
}

func PutBalance(addr bean.CommonAddress, balance *big.Int) error {
    account := GetAccount(addr)
    account.Balance = balance
    return PutAccount(account)
}

func GetNonce(addr bean.CommonAddress) int64 {
    account := GetAccount(addr)
    return account.Nonce
}

func PutNonce(addr bean.CommonAddress, nonce int64) error {
    account := GetAccount(addr)
    account.Nonce = nonce
    return PutAccount(account)
}

func GetAccountsAddress() []bean.CommonAddress {
    db := GetDatabase()
    key := mycrypto.Hash256([]byte("accounts"))
    value, err := db.Load(key)
    if err != nil {
        return make([]bean.CommonAddress, 0)
    }
    var ca []bean.CommonAddress
    err = json.Unmarshal(value, &ca)
    if err != nil {
        return make([]bean.CommonAddress, 0)
    }
    return ca
}

func AddAccountsAddress(account *bean.Account) error {
    db := GetDatabase()
    key := mycrypto.Hash256([]byte("accounts"))
    ca := GetAccountsAddress()
    ca = append(ca, account.Addr)
    value, err := json.Marshal(ca)
    if err != nil {
        return err
    }
    return db.Store(key, value)
}