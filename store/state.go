package store

import (
    "time"
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
    chainId    config.ChainIdType
    prvKey     *mycrypto.PrivateKey
    pubKey     *mycrypto.Point
    address    accounts.CommonAddress
    port       bean.Port
    isRelay    bool
)

func InitState(nodeConfig *config.NodeConfig)  {

    keystore := nodeConfig.Keystore
    node, _ := accounts.OpenKeystore(keystore)
    if node != nil {
        prvKey = node.PrvKey
        pubKey = node.PrvKey.PubKey
        address = node.Address()
    } else {
        panic("keystore file not exists!")
    }

    myIndex := nodeConfig.GetMyIndex()
    chainId = config.Hex2ChainId(nodeConfig.ChainId)
    bootNodes := nodeConfig.BootNodes
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
    port = bean.Port(nodeConfig.Port)
    isRelay = accounts.PubKey2Address(pubKey).Hex() == bootNodes[0].Address
}

func GenerateBlock(members []*bean.Peer) (*bean.Block, error) {
    dt := database.BeginTransaction()
    height := database.GetMaxHeight() + 1
    ts := PickTransactions(BlockGasLimit)
    //fmt.Println()
    //if lastLeader != nil {
    //    fmt.Println("last leader:   ", accounts.PubKey2Address(lastLeader))
    //} else {
    //    fmt.Println("last leader:   ")
    //}
    //fmt.Println("last minors:   ", lastMinors)
    //fmt.Println("last prize:    ", lastPrize)
    //fmt.Println()
    var bpt *bean.Transaction
    if lastPrize != nil {
        bpt = GenerateBlockPrizeTransaction()
        if bpt != nil {
            ts = append(ts, bpt)
        }
    }

    gasSum := new(big.Int)
    for _, t := range ts {
        subDt := dt.BeginTransaction()
        g, _ := execute(subDt, t)
        gasSum = new(big.Int).Add(gasSum, g)
        subDt.Commit()
    }
    timestamp := time.Now().Unix()
    stateRoot := dt.GetTotalStateRoot()
    gasUsed := gasSum
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
    previousBlock := database.GetHighestBlock()
    if previousBlock == nil {
        previousHash = []byte{}
    } else {
        h, err := previousBlock.BlockHash()
        if err != nil {
            return nil, err
        }
        previousHash = h
    }
    block := &bean.Block{
        Header: &bean.BlockHeader{
            Version:      Version,
            PreviousHash: previousHash,
            ChainId: GetChainId(),
            GasLimit: BlockGasLimit,
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
    dt.Discard()
    return block, nil
}

func GenerateBlockPrizeTransaction() *bean.Transaction {
    numMinors := len(lastMinors)
    leaderPrize := new(big.Int).Rsh(lastPrize, 1)
    leftPrize := new(big.Int).Sub(lastPrize, leaderPrize)
    var minorPrize *big.Int
    if numMinors > 0 {
        minorPrize = new(big.Int).Div(leftPrize, new(big.Int).SetInt64(int64(numMinors)))
    }
    trans := make([]*bean.Transaction, len(lastMinors) + 1)

    dataL := &bean.TransactionData{
        Version: Version,
        Type: BlockPrizeType,
        To: accounts.PubKey2Address(lastLeader).Hex(),
        DestChain: GetChainId(),
        Amount: leaderPrize,
        Timestamp: time.Now().Unix(),
        Data: []byte("block prize for leader"),
    }
    trans[0] = &bean.Transaction{Data: dataL}

    for i := 1; i < len(trans); i++ {
        dataM := &bean.TransactionData{
            Version: Version,
            Type: BlockPrizeType,
            To: accounts.PubKey2Address(lastMinors[i - 1]).Hex(),
            DestChain: GetChainId(),
            Amount: minorPrize,
            Timestamp: time.Now().Unix(),
            Data: []byte("block prize for minor"),
        }
        trans[i] = &bean.Transaction{Data: dataM}
    }

    b, err := json.Marshal(trans)
    if err != nil {
        return nil
    }
    data := &bean.TransactionData{
        Version: Version,
        Type: BlockPrizeType,
        Timestamp: time.Now().Unix(),
        Data: b,
    }

    //lastLeader = nil
    //lastMinors = nil
    //lastPrize = nil
    return &bean.Transaction{Data: data}
}

func GetPubKey() *mycrypto.Point {
    return pubKey
}

func GetAddress() accounts.CommonAddress {
    return address
}

func GetChainId() config.ChainIdType {
    return chainId
}

func GetPrvKey() *mycrypto.PrivateKey {
    return prvKey
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
