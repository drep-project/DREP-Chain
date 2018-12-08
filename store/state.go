package store

import (
    "BlockChainTest/bean"
    "BlockChainTest/mycrypto"
    "math/big"
    "BlockChainTest/database"
    "encoding/json"
    "errors"
    "time"
    "BlockChainTest/trie"
    "BlockChainTest/accounts"
    "BlockChainTest/config"
)

var (
    chainId  int64
    prvKey   *mycrypto.PrivateKey
    pubKey   *mycrypto.Point
    address  accounts.CommonAddress

    port bean.Port
    nodes map[string] *accounts.Node
)

func init()  {

    keystore := config.GetKeystore()
    node, _ := accounts.OpenKeystore(keystore)
    if node != nil {
        prvKey = node.PrvKey
        pubKey = node.PrvKey.PubKey
        address = node.Address()
    } else {
        panic("keystore file not exists!")
    }

    myIndex := config.GetMyIndex()
    chainId = config.GetChainId()
    debugNodes := config.GetDebugNodes()
    minerNum := len(debugNodes)

    curMiner = -1
    minerIndex = minerNum - 1

    for i := 0; i < minerNum; i++ {
        peer := &bean.Peer{
            IP:     bean.IP(debugNodes[i].IP),
            Port:   bean.Port(debugNodes[i].Port),
        }
        if i != myIndex {
            peer.PubKey = config.ParsePK(debugNodes[i].PubKey)
        } else {
            peer.PubKey = pubKey
        }
        curMiners = append(curMiners, peer)
        miners = append(miners, peer)
        AddPeer(peer)
        //database.PutBalanceOutSideTransaction(accounts.PubKey2Address(peer.PubKey), chainId, big.NewInt(100000000))
    }
    adminPubKey = miners[0].PubKey
    IsStart = myIndex < minerNum

    port = bean.Port(config.GetPort())
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

func GenerateBlock(members []*bean.Peer) (*bean.Block, error) {
    maxHeight := database.GetMaxHeight()
    height := maxHeight + 1
    ts := PickTransactions(BlockGasLimit)
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
    gasUsed := GetGasSum(ts).Bytes()
    //if ExceedGasLimit(gasUsed, gasLimit) {
    //    return nil, errors.New("gas used exceeds gas limit")
    //}
    timestamp := time.Now().Unix()
    stateRoot := GetStateRoot(ts)
    txHashes, err := GetTxHashes(ts)
    if err != nil {
        return nil, err
    }
    merkle := trie.NewMerkle(txHashes)
    merkleRoot := merkle.Root.Hash
    var memberPks []*mycrypto.Point = nil
    for _, p := range members {
        memberPks = append(memberPks, p.PubKey)
    }
    return &bean.Block{
        Header: &bean.BlockHeader{
            Version: Version,
            PreviousHash: previousHash,
            GasLimit: BlockGasLimit.Bytes(),
            GasUsed: gasUsed,
            Timestamp: timestamp,
            StateRoot: stateRoot,
            MerkleRoot: merkleRoot,
            TxHashes: txHashes,
            Height: height,
            LeaderPubKey:GetPubKey(),
            MinorPubKeys:memberPks,
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

func GetPort() bean.Port {
    return port
}

func GetGasSum(ts []*bean.Transaction) *big.Int {
    gasSum := new(big.Int)
    for _, tx := range ts {
        gasSum = gasSum.Add(gasSum, tx.GetGas())
    }
    return gasSum
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
