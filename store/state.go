package store

import (
    "sync"
    "BlockChainTest/network"
    "BlockChainTest/bean"
    "BlockChainTest/mycrypto"
    "math/big"
    "BlockChainTest/database"
    "encoding/json"
    "errors"
    "time"
    "BlockChainTest/trie"
    "BlockChainTest/accounts"
    "fmt"
    "BlockChainTest/config"
    "path"
)

var (
    lock     sync.Locker
    chainId  int64
    prvKey   *mycrypto.PrivateKey
    pubKey   *mycrypto.Point
    address  accounts.CommonAddress

    port network.Port
    nodes map[string] *accounts.Node
)

func init()  {
    lock = &sync.Mutex{}

    configDir := config.GetConfigDir()
    keystore := config.GetKeystore()
    keystorePath := path.Join(configDir, keystore)
    fmt.Println("keystorePath: ", keystorePath)
    node, _ := accounts.OpenKeystore(keystorePath)
    existed := node != nil
    if existed {
        prvKey = node.PrvKey
        pubKey = node.PrvKey.PubKey
        address = node.Address()
    } else {
        fmt.Println("keystore file not exists!")
        return
    }

    curMiner = -1

    minerNum := config.GetMinerNum()
    myIndex := config.GetMyIndex()
    chainId = config.GetChainId()
    debugNodes := config.GetDebugNodes()

    for i := 0; i < minerNum; i++ {
        peer := &network.Peer{
            IP:     network.IP(debugNodes[i].IP),
            Port:   network.Port(debugNodes[i].Port),
        }
        if i != myIndex {
            peer.PubKey = config.ParsePK(debugNodes[i].PubKey)
        } else {
            peer.PubKey = pubKey
        }
        curMiners = append(curMiners, peer)
        miners = append(miners, peer)
        AddPeer(peer)
        database.PutBalanceOutSideTransaction(accounts.PubKey2Address(peer.PubKey), chainId, big.NewInt(100000000))
    }
    if myIndex == 0 {
        adminPubKey = pubKey
    }
    IsStart = myIndex < minerNum

    //if Solo {
    //    minerNum = 1
    //    ip0 = network.IP("127.0.0.1")
    //    port0 = network.Port(55555)
    //} else if LocalTest {
    //    ip0 = network.IP("127.0.0.1")
    //    ip1 = network.IP("127.0.0.1")
    //    port0 = network.Port(55555)
    //    port1 = network.Port(55556)
    //    port2 = network.Port(55557)
    //} else {
    //    ip0 = network.IP("192.168.3.231")
    //    ip1 = network.IP("192.168.3.197")
    //    ip2 = network.IP("192.168.3.236")
    //    port0 = network.Port(55555)
    //    port1 = network.Port(55555)
    //    port2 = network.Port(55555)
    //}
    //
}

func GenerateBlock() (*bean.Block, error) {
    maxHeight := database.GetMaxHeight()
    height := maxHeight + 1
    ts := PickTransactions(BlockGasLimit)
    fmt.Println("ts: ", ts)
    previousBlock := database.GetHighestBlock()
    var b, previousHash []byte
    var err error
    if previousBlock != nil {
        b, err = json.Marshal(previousBlock.Header)
        if err != nil {
            return nil, err
        }
        previousHash = mycrypto.Hash256(b)
    } else {
        previousHash = []byte{}
    }
    gasLimit := new(big.Int).SetInt64(int64(10000000)).Bytes()
    gasUsed := GetGasSum(ts).Bytes()
    if ExceedGasLimit(gasUsed, gasLimit) {
        return nil, errors.New("gas used exceeds gas limit")
    }
    timestamp := time.Now().Unix()
    stateRoot := GetStateRoot(ts)
    txHashes, err := GetTxHashes(ts)
    if err != nil {
        return nil, err
    }
    fmt.Println("txHashes: ", txHashes, len(txHashes))
    merkle := trie.NewMerkle(txHashes)
    merkleRoot := merkle.Root.Hash
    return &bean.Block{
        Header: &bean.BlockHeader{
            Version: Version,
            PreviousHash: previousHash,
            GasLimit: gasLimit,
            GasUsed: gasUsed,
            Timestamp: timestamp,
            StateRoot: stateRoot,
            MerkleRoot: merkleRoot,
            TxHashes: txHashes,
            Height: height,
            LeaderPubKey:GetPubKey(),
        },
        Data:&bean.BlockData{
            TxCount:int32(len(ts)),
            TxList:ts,
        },
    }, nil
}

func GetPubKey() *mycrypto.Point {
    return pubKey
}

func GetAddress() accounts.CommonAddress {
    return address
}

func GetChainId() int64 {
    return chainId
}

func GetPrvKey() *mycrypto.PrivateKey {
    return prvKey
}

func CreateAccount(addr string, chainId int64) (string, error) {
    IsRoot := chainId == accounts.RootChainID
    var (
        parent *accounts.Node
        parentFound bool
    )
    if !IsRoot {
        parent, parentFound = nodes[addr]
        if !parentFound {
            return "", errors.New("no parent account " + addr + " is found on the root chain")
        }
    }
    account, err := accounts.NewNormalAccount(parent, chainId)
    if err != nil {
        return "", err
    }
    database.PutStorageOutsideTransaction(account.Storage, account.Address, chainId)
    return account.Address.Hex(), nil
}

func GetAccounts() []string {
    acs := make([]string, len(nodes))
    i := 0
    for addr, _ := range nodes {
        acs[i] = addr
        i++
    }
    return acs
}

func GetPort() network.Port {
    return port
}

func GetGasSum(ts []*bean.Transaction) *big.Int {
    gasSum := new(big.Int)
    for _, tx := range ts {
        gasSum = gasSum.Add(gasSum, tx.GetGas())
    }
    return gasSum
}

func ExceedGasLimit(used, limit []byte) bool {
    if new(big.Int).SetBytes(used).Cmp(new(big.Int).SetBytes(limit)) > 0 {
        return true
    }
    return false
}

func GetStateRoot(ts []*bean.Transaction) []byte {
    //for _, tx := range ts {
    //    from := bean.PubKey2Address(tx.Data.PubKey)
    //    to := bean.Hex2Address(tx.Data.To)
    //    gasUsed := tx.GetGas()
    //    nonce := tx.Data.Nonce
    //    amount := new(big.Int).SetBytes(tx.Data.Amount)
    //    prevSenderBalance := database.GetBalance(from)
    //    prevReceiverBalance := database.GetBalance(to)
    //    newSenderBalance := new(big.Int).Sub(prevSenderBalance, amount)
    //    newSenderBalance = newSenderBalance.Sub(newSenderBalance, gasUsed)
    //    newReceiverBalance := new(big.Int).Add(prevReceiverBalance, amount)
    //    database.PutBalance(from, newSenderBalance)
    //    database.PutBalance(to, newReceiverBalance)
    //    database.PutNonce(from, nonce)
    //}
    return database.GetDB().GetStateRoot()
}

func GetTxHashes(ts []*bean.Transaction) ([][]byte, error) {
    txHashes := make([][]byte, len(ts))
    for i, tx := range ts {
        b, err := json.Marshal(tx.Data)
        if err != nil {
            return nil, err
        }
        txHashes[i] = mycrypto.Hash256(b)
    }
    return txHashes, nil
}
