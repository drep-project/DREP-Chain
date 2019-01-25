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
    "fmt"
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
    }
    adminPubKey = miners[0].PubKey
    port = bean.Port(config.GetPort())
}

func GenerateBlock(members []*bean.Peer) (*bean.Block, error) {
    maxHeight := database.GetMaxHeight()
    height := maxHeight + 1
    ts := PickTransactions(BlockGasLimit)

    var bpt []*bean.Transaction
    if lastPrize != nil {
        bpt = GenerateBlockPrizeTransaction()
        if bpt != nil {
            ts = append(ts, bpt...)
        }
    }

    if height == 0 || height % 5 > 0 {
        increments := generateInc(height)
        rtx := GenerateGainTransaction(GetAddress(), GetChainId(), increments)
        ts = append(ts, rtx)
    }

    timestamp := time.Now().Unix()
    previousHash := database.GetPreviousHash()

    dbTran := database.BeginTransaction()
    total := new(big.Int)
    for _, t := range ts {
        if t.Data.Type == BlockPrizeType {
            continue
        }
        if t.Data.Type == GainType {
            continue
        }
        gasUsed := executeWithT(t, dbTran)
        fmt.Println("Delete transaction ", *t)
        fmt.Println(removeTransaction(t))
        if gasUsed != nil {
            total.Add(total, gasUsed)
        }
    }
    dbTran.Discard()

    stateRoot := GetStateRoot(ts)
    txHashes, err := GetTxHashes(ts)
    if err != nil {
        return nil, err
    }
    merkle := trie.NewMerkle(txHashes)
    merkleRoot := merkle.Root.Hash

    adminNodes := config.GetAdminNodes()
    var leaderPk *mycrypto.Point
    var memberPks = make([]*mycrypto.Point, 0)
    for _, p := range members {
        memberPks = append(memberPks, p.PubKey)
    }
    if lastIndex == 0 {
        leaderPk = GetPubKey()
        for _, n := range adminNodes {
            memberPks = append(memberPks, config.ParsePK(n.PubKey))
        }
    } else {
        leaderPk = config.ParsePK(adminNodes[lastIndex - 1].PubKey)
        for i := lastIndex; i < 6; i++ {
            memberPks = append(memberPks, config.ParsePK(adminNodes[i].PubKey))
        }
        memberPks = append(memberPks, GetPubKey())
        for i := 0; i < lastIndex - 1; i++ {
            memberPks = append(memberPks, config.ParsePK(adminNodes[i].PubKey))
        }
    }

    return &bean.Block{
        Header: &bean.BlockHeader{
            Version: Version,
            ChainId: GetChainId(),
            PreviousHash: previousHash,
            GasLimit: BlockGasLimit.Bytes(),
            GasUsed: total.Bytes(),
            Timestamp: timestamp,
            StateRoot: stateRoot,
            MerkleRoot: merkleRoot,
            TxHashes: txHashes,
            Height: height,
            LeaderPubKey: leaderPk,
            MinorPubKeys: memberPks,
        },
        Data:&bean.BlockData{
            TxCount:int32(len(ts)),
            TxList:ts,
        },
    }, nil
}

func GenerateBlockPrizeTransaction() []*bean.Transaction {
    numMinors := len(lastMinors)
    leaderPrize := new(big.Int).Rsh(lastPrize, 1)
    leftPrize := new(big.Int).Sub(lastPrize, leaderPrize)
    var minorPrize *big.Int
    if numMinors > 0 {
        minorPrize = new(big.Int).Div(leftPrize, new(big.Int).SetInt64(int64(numMinors)))
    }
    trans := make([]*bean.Transaction, 1 + len(lastMinors))

    dataL := &bean.TransactionData{
        Version: Version,
        Type: BlockPrizeType,
        To: accounts.PubKey2Address(lastLeader).Hex(),
        ChainId: GetChainId(),
        DestChain: GetChainId(),
        Amount: leaderPrize.Bytes(),
        Timestamp: time.Now().Unix(),
    }
    trans[0] = &bean.Transaction{Data: dataL}

    for i := 0; i < len(lastMinors); i++ {
        dataM := &bean.TransactionData{
            Version: Version,
            Type: BlockPrizeType,
            To: accounts.PubKey2Address(lastMinors[i]).Hex(),
            ChainId: GetChainId(),
            DestChain: GetChainId(),
            Amount: minorPrize.Bytes(),
            Timestamp: time.Now().Unix(),
        }
        trans[i + 1] = &bean.Transaction{Data: dataM}
    }

    return trans
}

func GenerateGainTransaction(addr accounts.CommonAddress, chainId int64, increments []byte) *bean.Transaction {
    data := &bean.TransactionData{
        Version: Version,
        Type: GainType,
        ChainId: chainId,
        DestChain: chainId,
        To: addr.Hex(),
        GasLimit: GainGas.Bytes(),
        GasPrice: DefaultGasPrice.Bytes(),
        Timestamp: time.Now().Unix(),
        PubKey: GetPubKey(),
        Data: increments,
    }
    tx := &bean.Transaction{Data: data}
    return tx
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

func GetStateRoot(ts []*bean.Transaction) []byte {
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
