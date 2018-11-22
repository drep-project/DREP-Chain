package database

import (
    "BlockChainTest/bean"
    "math/big"
    "BlockChainTest/mycrypto"
    "strconv"
    "encoding/json"
    "BlockChainTest/accounts"
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

func GetAccount(addr accounts.CommonAddress) accounts.Account {
    db := GetDatabase()
    key := mycrypto.Hash256([]byte("account_" + addr.Hex()))
    value, err := db.Load(key)
    if err != nil {
        return nil
    }
    var m *accounts.MainAccount
    var s *accounts.SubAccount
    mErr := json.Unmarshal(value, m)
    sErr := json.Unmarshal(value, s)
    if mErr == nil {
        return m
    }
    if sErr == nil {
        return s
    }
    return nil
}

func PutAccount(acc accounts.Account) error {
    db := GetDatabase()
    key := mycrypto.Hash256([]byte("account_" + acc.Address().Hex()))
    value, err := json.Marshal(acc)
    if err != nil {
        return err
    }
    err = db.Store(key, value)
    if err != nil {
        return err
    }
    db.Trie.Insert(key, value)
    return nil
}

func GetBalance(addr accounts.CommonAddress) *big.Int {
    acc := GetAccount(addr)
    return acc.GetStorage().Balance
}

func PutBalance(addr accounts.CommonAddress, balance *big.Int) error {
    acc := GetAccount(addr)
    storage := acc.GetStorage()
    storage.Balance = balance
    return PutAccount(acc)
}

func GetNonce(addr accounts.CommonAddress) int64 {
    acc := GetAccount(addr)
    return acc.GetStorage().Nonce
}

func PutNonce(addr accounts.CommonAddress, nonce int64) error {
    acc := GetAccount(addr)
    storage := acc.GetStorage()
    storage.Nonce = nonce
    return PutAccount(acc)
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
    nodes[node.Address.Hex()] = node
    return PutNodes(nodes)
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
//        return errors.New("sender accounts not found")
//    }
//    addrTo := bean.Hex2Address(to)
//    receiver := GetAccount(addrTo)
//    if receiver == nil {
//        return errors.New("receiver accounts not found")
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
//        return errors.New("failed to modify sender accounts, error: " + err.Error())
//    }
//    err = PutAccount(receiver)
//    if err != nil {
//        return errors.New("failed to modify receiver accounts, error: " + err.Error())
//    }
//    return nil
//}