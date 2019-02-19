package service

import (
    "encoding/json"
    "github.com/AsynkronIT/protoactor-go/actor"
    "github.com/drep-project/drep-chain/app"
    chainTypes "github.com/drep-project/drep-chain/chain/types"
    "github.com/drep-project/drep-chain/common"
    "github.com/drep-project/drep-chain/crypto"
    "github.com/drep-project/drep-chain/crypto/secp256k1"
    "github.com/drep-project/drep-chain/database"
    "github.com/drep-project/drep-chain/log"
    p2pService "github.com/drep-project/drep-chain/network/service"
    p2pTypes "github.com/drep-project/drep-chain/network/types"
    "gopkg.in/urfave/cli.v1"
    "math/big"
    "reflect"
    "strconv"
    "sync"
    "time"
)

type ChainService struct {
    p2pServer *p2pService.P2pService  `service:"p2p"`
    databaseService *database.DatabaseService  `service:"database"`
    apis   []app.API

    chainId common.ChainIdType

    lock sync.RWMutex
    addBlockSync sync.Mutex
    StartComplete  chan struct{}
    stopChanel   chan struct{}

    prvKey *secp256k1.PrivateKey
    curMaxHeight int64

    config *chainTypes.ChainConfig
    pid *actor.PID
}


func (chainService *ChainService) Name() string {
    return "chain"
}
func (chainService *ChainService) Api() []app.API {
    return chainService.apis
}
func (chainService *ChainService) Flags() []cli.Flag {
    return []cli.Flag{}
}

func (chainService *ChainService) P2pMessages() map[int]interface{} {
    return map[int]interface{}{
        chainTypes.MsgTypeBlockReq : reflect.TypeOf(chainTypes.BlockReq{}),
        chainTypes.MsgTypeBlockResp : reflect.TypeOf(chainTypes.BlockResp{}),
        chainTypes.MsgTypeBlock : reflect.TypeOf(chainTypes.Block{}),
        chainTypes.MsgTypeTransaction : reflect.TypeOf(chainTypes.Transaction{}),
        chainTypes.MsgTypePeerState : reflect.TypeOf(chainTypes.PeerState{}),
        chainTypes.MsgTypeReqPeerState : reflect.TypeOf(chainTypes.ReqPeerState{}),
    }
}

func (chainService *ChainService) Init(executeContext *app.ExecuteContext) error {
    chainService.config = &chainTypes.ChainConfig{}
    phase := executeContext.GetConfig(chainService.Name())
    err := json.Unmarshal(phase, chainService.config)
    if err != nil {
        return err
    }

    props := actor.FromProducer(func() actor.Actor {
        return chainService
    })
    pid, err := actor.SpawnNamed(props, "chain_message")
    if err != nil {
        panic(err)
    }
    chainService.pid = pid
    router :=  chainService.p2pServer.Router
    chainP2pMessage := chainService.P2pMessages()
    for msgType, _ := range chainP2pMessage {
        router.RegisterMsgHandler(msgType,pid)
    }

    chainService.apis = []app.API{
        app.API{
            Namespace: "chain",
            Version:   "1.0",
            Service: &ChainApi{
                chain: chainService,
            },
            Public: true,
        },
    }
    return nil
}

func (chainService *ChainService) Start(executeContext *app.ExecuteContext) error {
    return nil
}

func (chainService *ChainService) Stop(executeContext *app.ExecuteContext) error {
    return nil
}

func (chainService *ChainService) sendBlock(block *chainTypes.Block) {
    chainService.p2pServer.Broadcast(block)
}

func (chainService *ChainService) ProcessBlock(block *chainTypes.Block) (*big.Int, error) {
    chainService.addBlockSync.Lock()
    chainService.addBlockSync.Unlock()
    log.Trace("Process block leader.", "LeaderPubKey", crypto.PubKey2Address(block.Header.LeaderPubKey).Hex(), " height ", strconv.FormatInt(block.Header.Height,10))
  //  return store.ExecuteTransactions(block)
    return nil,nil
}

func (chainService *ChainService) ProcessBlockReq(peer *p2pTypes.Peer, req *chainTypes.BlockReq) {
    /*
        from := req.Height + 1
        size := int64(200)

        for i := from; i <= database.GetMaxHeight(); {
            bs := database.GetBlocksFrom(i, size)
            resp := &bean.BlockResp{Height:database.GetMaxHeight(), Blocks:bs}
            n.p2pServer.Send(peer,resp)
            i += int64(len(bs))
        }
    */
}

func (chainService *ChainService) GenerateBlock(members []*secp256k1.PublicKey) (*chainTypes.Block, error) {
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
    return nil, nil
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
        Amount: leaderPrize.Bytes(),
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
            Amount: minorPrize.Bytes(),
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