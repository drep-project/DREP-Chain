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
        account, err := bean.UnmarshalAccount(value)
        if err == nil {
            return account
        }
        return &bean.Account{Addr:addr, Nonce:0, Balance:big.NewInt(0)}
    }
    return &bean.Account{Addr:addr, Nonce:0, Balance:big.NewInt(0)}
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
    return nil
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

func PutPrv(prv map[string] *mycrypto.PrivateKey) error {
    db := GetDatabase()
    key := mycrypto.Hash256([]byte("private_keys"))
    value, err := json.Marshal(prv)
    if err != nil {
        return err
    }
    return db.Store(key, value)
}

func GetPrv() map[string] *mycrypto.PrivateKey {
    db := GetDatabase()
    key := mycrypto.Hash256([]byte("private_keys"))
    value, err := db.Load(key)
    if err != nil {
        return nil
    }
    var prv map[string] *mycrypto.PrivateKey
    err = json.Unmarshal(value, &prv)
    if err != nil {
        return nil
    }
    return prv
}

func AddAccount(hexStr string, prv map[string] *mycrypto.PrivateKey) error {
    addr := bean.Hex2Address(hexStr)
    account := &bean.Account{
        Addr: addr,
        Nonce: 0,
        Balance: new(big.Int),
    }
    err := PutAccount(account)
    if err != nil {
        return err
    }
    return PutPrv(prv)
}

func GetMostRecentBlocks(n int64) []*bean.Block {
    height := GetMaxHeight()
    if height == -1 {
        return nil
    }
    return GetBlocksFrom(height - n, n)
}

//func SendTransaction(from, to, amount string) error {
//    addrFrom := bean.Hex2Address(from)
//    sender := GetAccount(addrFrom)
//    if sender == nil {
//        return errors.New("sender account not found")
//    }
//    addrTo := bean.Hex2Address(to)
//    receiver := GetAccount(addrTo)
//    if receiver == nil {
//        return errors.New("receiver account not found")
//    }
//    value, ok := new(big.Int).SetString(amount, 10)
//    if !ok {
//        return errors.New("wrong amount value")
//    }
//    if sender.Balance.Cmp(value) < 0 {
//        return errors.New("do not have enough balance to send")
//    }
//    sender.Balance = new(big.Int).Sub(sender.Balance, value)
//    sender.Nonce++
//    receiver.Balance = new(big.Int).Add(receiver.Balance, value)
//    err := PutAccount(sender)
//    if err != nil {
//        return errors.New("failed to modify sender account, error: " + err.Error())
//    }
//    err = PutAccount(receiver)
//    if err != nil {
//        return errors.New("failed to modify receiver account, error: " + err.Error())
//    }
//    return nil
//}