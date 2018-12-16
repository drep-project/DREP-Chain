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
    chainId  config.ChainIdType
    prvKey   *mycrypto.PrivateKey
    pubKey   *mycrypto.Point
    address  accounts.CommonAddress

    port bean.Port
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
}

func GenerateBlock(members []*bean.Peer) (*bean.Block, error) {
    dt := database.BeginTransaction()
    height := database.GetMaxHeight() + 1
    ts := PickTransactions(BlockGasLimit)
    gasSum := new(big.Int)
    //fmt.Println("before generate block: ", hex.EncodeToString(database.GetStateRoot()))
    for _, t := range ts {
        subDt := dt.BeginTransaction()
        g, _ := execute(subDt, t)
        gasSum = new(big.Int).Add(gasSum, g)
        subDt.Commit()
    }
    //fmt.Println("after generate block: ", hex.EncodeToString(database.GetStateRoot()))
    timestamp := time.Now().Unix()
    stateRoot := dt.GetTotalStateRoot()
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
    dt.Discard()
    return block, nil
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
