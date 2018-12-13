package store

import (
    "time"
    "errors"
    "math/big"
    "encoding/json"
    "BlockChainTest/trie"
    "BlockChainTest/bean"
    "BlockChainTest/mycrypto"
    "BlockChainTest/database"
    "BlockChainTest/accounts"
    "BlockChainTest/config"
    "BlockChainTest/core/common"
)

var (
    chainId  int64
    prvKey   *mycrypto.PrivateKey
    pubKey   *mycrypto.Point
    address  accounts.CommonAddress

    port bean.Port
    nodes map[string] *accounts.Node
)

func InitState(config *config.NodeConfig)  {
    keystore := config.Keystore
    node, _ := accounts.OpenKeystore(keystore)
    if node != nil {
        prvKey = node.PrvKey
        pubKey = node.PrvKey.PubKey
        address = node.Address()
    } else {
        panic("keystore file not exists!")
    }

    myIndex := config.GetMyIndex()
    chainId = config.ChainId
    bootNodes := config.BootNodes
    minerNum := len(bootNodes)

    curMiner = -1
    minerIndex = minerNum - 1

    for i := 0; i < minerNum; i++ {
        peer := &bean.Peer{
            IP:     bean.IP(bootNodes[i].IP),
            Port:   bean.Port(bootNodes[i].Port),
        }
        if i != myIndex {
            peer.PubKey = common.ParsePK(bootNodes[i].PubKey)
        } else {
            peer.PubKey = pubKey
        }
        curMiners = append(curMiners, peer)
        miners = append(miners, peer)
        AddPeer(peer)
    }
    adminPubKey = miners[0].PubKey

    port = bean.Port(config.Port)
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
    dbTran := database.BeginTransaction()
    height := database.GetMaxHeightInsideTransaction(dbTran) + 1
    ts := PickTransactions(BlockGasLimit)
    gasSum := new(big.Int)
    //fmt.Println("before generate block: ", hex.EncodeToString(database.GetStateRoot()))
    for _, t := range ts {
        g, _ := execute(dbTran, t)
        gasSum = new(big.Int).Add(gasSum, g)
    }
    //fmt.Println("after generate block: ", hex.EncodeToString(database.GetStateRoot()))
    timestamp := time.Now().Unix()
    stateRoot := database.GetStateRoot()
    gasUsed := gasSum.Bytes()
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

    var previousHash []byte
    previousBlock := database.GetHighestBlockInsideTransaction(dbTran)
    if previousBlock == nil {
        previousHash = []byte{}
    } else {
        h, err := previousBlock.BlockHash()
        if err != nil {
            return nil, err
        }
        previousHash = h
    }
    //fmt.Println("generate block height: ", height)
    block := &bean.Block{
        Header: &bean.BlockHeader{
            Version:      Version,
            PreviousHash: previousHash,
            ChainId: GetChainId(),
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
        Data: &bean.BlockData{
            TxCount: int32(len(ts)),
            TxList:  ts,
        },
    }
    dbTran.Discard()
    return block, nil
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

func GetPort() bean.Port {
    return port
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
