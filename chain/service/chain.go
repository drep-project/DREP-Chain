package service

import (
    "encoding/json"
    "github.com/AsynkronIT/protoactor-go/actor"
    "github.com/drep-project/drep-chain/app"
    chainTypes "github.com/drep-project/drep-chain/chain/types"
    "github.com/drep-project/drep-chain/common"
    "github.com/drep-project/drep-chain/crypto"
    "github.com/drep-project/drep-chain/crypto/secp256k1"
    "github.com/drep-project/drep-chain/crypto/sha3"
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
    "encoding/hex"
    "github.com/drep-project/drep-chain/chain/service"
    "BlockChainTest/chain/types"
)
var (
    rootChain common.ChainIdType
)

type ChainService struct {
    p2pServer *p2pService.P2pService  `service:"p2p"`
    databaseService *database.DatabaseService  `service:"database"`
    transactionPool *TransactionPool
    isRelay bool
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
    defer chainService.addBlockSync.Unlock()
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

func (chainService *ChainService) GenerateBlock(leaderKey *secp256k1.PublicKey, members []*secp256k1.PublicKey) (*chainTypes.Block, error) {
    chainService.databaseService.BeginTransaction()
    height := chainService.databaseService.GetMaxHeight()
    height++

    ts := chainService.transactionPool.PickTransactions(BlockGasLimit)
    if lastPrize != nil {
        bpt := chainService.GenerateBlockPrizeTransaction()
        if bpt != nil {
            ts = append(ts, bpt)
        }
    }

    gasUsed := new(big.Int)
    for _, t := range ts {
        g, _ := chainService.execute(t)
        gasUsed.Add(gasUsed, g)
    }

    timestamp := time.Now().Unix()
    previousHash := chainService.databaseService.GetPreviousBlockHash()

    stateRoot := chainService.databaseService.GetStateRoot()
    txHashes, _ := chainService.GetTxHashes(ts)
    merkle := chainService.databaseService.NewMerkle(txHashes)
    merkleRoot := merkle.Root.Hash

    var memberPks []*secp256k1.PublicKey
    for _, p := range members {
        memberPks = append(memberPks, p)
    }

    block := &chainTypes.Block{
        Header: &chainTypes.BlockHeader{
            Version:       Version,
            PreviousHash:  previousHash,
            ChainId:       chainService.chainId,
            GasLimit:      BlockGasLimit,
            GasUsed:       gasUsed,
            Timestamp:     timestamp,
            StateRoot:     stateRoot,
            MerkleRoot:    merkleRoot,
            TxHashes:      txHashes,
            Height:        height,
            LeaderPubKey : leaderKey,
            MinorPubKeys:  memberPks,
        },
        Data: &chainTypes.BlockData{
            TxCount: int32(len(ts)),
            TxList:  ts,
        },
    }

    chainService.databaseService.Discard()
    return block, nil
}

func (chainService *ChainService) GenerateBlockPrizeTransaction() *chainTypes.Transaction {
    numMinors := len(lastMinors)
    leaderPrize := new(big.Int).Rsh(lastPrize, 1)
    leftPrize := new(big.Int).Sub(lastPrize, leaderPrize)
    var minorPrize *big.Int
    if numMinors > 0 {
        minorPrize = new(big.Int).Div(leftPrize, new(big.Int).SetInt64(int64(numMinors)))
    }
    trans := make([]*chainTypes.Transaction, len(lastMinors) + 1)

    dataL := &chainTypes.TransactionData{
        Version:   Version,
        Type:      BlockPrizeType,
        To:        crypto.PubKey2Address(lastLeader).Hex(),
        DestChain: chainService.chainId,
        Amount:    *leaderPrize,
        Timestamp: time.Now().Unix(),
        Data:      []byte("block prize for leader"),
    }
    trans[0] = &chainTypes.Transaction{Data: dataL}

    for i := 1; i < len(trans); i++ {
        dataM := &chainTypes.TransactionData{
            Version:   Version,
            Type:      BlockPrizeType,
            To:        crypto.PubKey2Address(lastMinors[i - 1]).Hex(),
            DestChain: chainService.chainId,
            Amount:    *minorPrize,
            Timestamp: time.Now().Unix(),
            Data:      []byte("block prize for minor"),
        }
        trans[i] = &chainTypes.Transaction{Data: dataM}
    }

    b, err := json.Marshal(trans)
    if err != nil {
        return nil
    }
    data := &chainTypes.TransactionData{
        Version: Version,
        Type: BlockPrizeType,
        Timestamp: time.Now().Unix(),
        Data: b,
    }

    return &chainTypes.Transaction{Data: data}
}

func (chainService *ChainService)GetTxHashes(ts []*chainTypes.Transaction) ([][]byte, error) {
    txHashes := make([][]byte, len(ts))
    for i, tx := range ts {
        b, err := json.Marshal(tx.Data)
        if err != nil {
            return nil, err
        }
        txHashes[i] = sha3.Hash256(b)
    }
    return txHashes, nil
}

func (chainService *ChainService) RootChain() common.ChainIdType {
    return rootChain
}

func (chainService *ChainService) GenerateBalanceTransaction(from *secp256k1.PublicKey, to crypto.CommonAddress, chainId common.ChainIdType, amount *big.Int) *chainTypes.Transaction {
    address := crypto.PubKey2Address(from)
    nonce := chainService.databaseService.GetNonce(address, chainId, false)
    data := &chainTypes.TransactionData{
        Version:   Version,
        Nonce:     nonce,
        Type:      TransferType,
        To:        to,
        ChainId:   chainId,
        Amount:    amount,
        GasPrice:  DefaultGasPrice,
        GasLimit:  TransferGas,
        Timestamp: time.Now().Unix(),
        PubKey:    from,
    }
    return &chainTypes.Transaction{Data: data}
}

func (chainService *ChainService) GenerateCreateContractTransaction(from *secp256k1.PublicKey, to crypto.CommonAddress, chainId common.ChainIdType, byteCode []byte) *chainTypes.Transaction {
    address := crypto.PubKey2Address(from)
    nonce := chainService.databaseService.GetNonce(address, chainId, false)
    nonce++
    data := &chainTypes.TransactionData{
        Nonce:     nonce,
        Type:      store.CreateContractType,
        ChainId:   chainId,
        GasPrice:  DefaultGasPrice,
        GasLimit:  CreateContractGas,
        Timestamp: time.Now().Unix(),
        Data:      make([]byte, len(byteCode) + 1),
        PubKey:    from,
    }
    copy(data.Data[1:], byteCode)
    data.Data[0] = 2
    return &chainTypes.Transaction{Data: data}
}

func (chainService *ChainService) GenerateCallContractTransaction(from *secp256k1.PublicKey, to crypto.CommonAddress, chainId common.ChainIdType, input []byte, amount *big.Int, readOnly bool) *chainTypes.Transaction {
    address := crypto.PubKey2Address(from)
    nonce := chainService.databaseService.GetNonce(address, chainId, false)
    nonce++
    data := &chainTypes.TransactionData{
        Nonce:     nonce,
        Type:      store.CallContractType,
        ChainId:   chainId,
        To:        to,
        Amount:    amount,
        GasPrice:  DefaultGasPrice,
        GasLimit:  CallContractGas,
        Timestamp: time.Now().Unix(),
        PubKey:    from,
        Data:      make([]byte, len(input) + 1),
    }
    copy(data.Data[1:], input)
    if readOnly {
        data.Data[0] = 1
    } else {
        data.Data[0] = 0
    }
    return &chainTypes.Transaction{Data: data}
}