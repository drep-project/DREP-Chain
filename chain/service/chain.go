package service

import (
	"encoding/json"
	"errors"
	"github.com/drep-project/drep-chain/chain/txpool"
	"github.com/drep-project/drep-chain/pkgs/evm"
	"math/big"
	"sync"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/drep-project/drep-chain/app"
	"gopkg.in/urfave/cli.v1"

	"github.com/drep-project/drep-chain/common"
	"github.com/drep-project/drep-chain/common/event"
	"github.com/drep-project/drep-chain/crypto"
	"github.com/drep-project/drep-chain/crypto/secp256k1"
	"github.com/drep-project/drep-chain/crypto/sha3"
	"github.com/drep-project/drep-chain/database"
	"github.com/drep-project/drep-chain/rpc"

	chainTypes "github.com/drep-project/drep-chain/chain/types"
	p2pService "github.com/drep-project/drep-chain/network/service"
	rpc2 "github.com/drep-project/drep-chain/pkgs/rpc"
)

var (
	rootChain          app.ChainIdType
	DefaultChainConfig = &chainTypes.ChainConfig{
		RemotePort: 55555,
		ChainId:    app.ChainIdType{},
	}
)

type ChainService struct {
	RpcService      *rpc2.RpcService          `service:"rpc"`
	P2pServer       p2pService.P2P            `service:"p2p"`
	DatabaseService *database.DatabaseService `service:"database"`
	VmService       evm.Vm                    `service:"vm"`
	transactionPool *txpool.TransactionPool
	isRelay         bool
	apis            []app.API

	chainId app.ChainIdType

	lock          sync.RWMutex
	addBlockSync  sync.Mutex
	StartComplete chan struct{}
	stopChanel    chan struct{}

	// These fields are related to handling of orphan blocks.  They are
	// protected by a combination of the chain lock and the orphan lock.
	orphanLock   sync.RWMutex
	orphans      map[crypto.Hash]*chainTypes.OrphanBlock
	prevOrphans  map[crypto.Hash][]*chainTypes.OrphanBlock
	oldestOrphan *chainTypes.OrphanBlock

	prvKey       *secp256k1.PrivateKey
	peerStateMap map[string]*chainTypes.PeerState

	Index         *chainTypes.BlockIndex
	BestChain     *chainTypes.ChainView
	stateLock     sync.RWMutex
	StateSnapshot *chainTypes.BestState

	Config       *chainTypes.ChainConfig
	pid          *actor.PID
	genesisBlock *chainTypes.Block
	//Events related to sync blocks
	syncBlockEvent event.Feed
	//Maximum block height being synced
	//syncingMaxHeight int64
	syncMut sync.Mutex

	//提供新块订阅
	newBlockFeed event.Feed

	//从远端接收块头hash组
	headerHashCh chan []*syncHeaderHash

	//从远端接收到块
	blocksCh chan []*chainTypes.Block

	//所有需要同步的任务列表
	allTasks *heightSortedMap

	//正在同步中的任务列表，如果对应的块未到，会重新发布请求的
	pendingSyncTasks map[crypto.Hash]int64
}

type syncHeaderHash struct {
	headerHash *crypto.Hash
	height     int64
}

func (chainService *ChainService) ChainID() app.ChainIdType {
	return chainService.chainId
}

func (chainService *ChainService) Name() string {
	return "chain"
}

func (chainService *ChainService) Api() []app.API {
	return chainService.apis
}

func (chainService *ChainService) CommandFlags() ([]cli.Command, []cli.Flag) {
	return nil, []cli.Flag{}
}

func (chainService *ChainService) P2pMessages() map[int]interface{} {
	return map[int]interface{}{
		chainTypes.MsgTypeBlockReq:     chainTypes.BlockReq{},
		chainTypes.MsgTypeBlockResp:    chainTypes.BlockResp{},
		chainTypes.MsgTypeBlock:        chainTypes.Block{},
		chainTypes.MsgTypeTransaction:  chainTypes.Transaction{},
		chainTypes.MsgTypePeerState:    chainTypes.PeerState{},
		chainTypes.MsgTypeReqPeerState: chainTypes.ReqPeerState{},
		chainTypes.MsgTypeHeaderReq:    chainTypes.HeaderReq{},
		chainTypes.MsgTypeHeaderRsp:    chainTypes.HeaderRsp{},
	}
}

func (chainService *ChainService) Init(executeContext *app.ExecuteContext) error {
	chainService.Config = DefaultChainConfig
	err := executeContext.UnmashalConfig(chainService.Name(), chainService.Config)
	if err != nil {
		return err
	}
	chainService.Index = chainTypes.NewBlockIndex()
	chainService.BestChain = chainTypes.NewChainView(nil)
	chainService.orphans = make(map[crypto.Hash]*chainTypes.OrphanBlock)
	chainService.prevOrphans = make(map[crypto.Hash][]*chainTypes.OrphanBlock)
	chainService.peerStateMap = make(map[string]*chainTypes.PeerState)
	chainService.headerHashCh = make(chan []*syncHeaderHash)
	chainService.blocksCh = make(chan []*chainTypes.Block)
	chainService.allTasks = newHeightSortedMap()
	chainService.pendingSyncTasks = make(map[crypto.Hash]int64)

	chainService.genesisBlock, err = chainService.GenesisBlock()
	if err != nil {
		return err
	}
	hash := chainService.genesisBlock.Header.Hash()
	block, err := chainService.DatabaseService.GetBlock(hash)
	if err != nil && err.Error() != "leveldb: not found" {
		return nil
	}
	if block == nil {
		//generate genisis block
		chainService.ProcessGenisisBlock()
		chainService.createChainState()
	}

	chainService.InitStates()
	chainService.transactionPool = txpool.NewTransactionPool(chainService.DatabaseService)

	props := actor.FromProducer(func() actor.Actor {
		return chainService
	})

	pid, err := actor.SpawnNamed(props, "chain_message")
	if err != nil {
		panic(err)
	}

	chainService.pid = pid
	router := chainService.P2pServer.GetRouter()
	chainP2pMessage := chainService.P2pMessages()
	for msgType, _ := range chainP2pMessage {
		router.RegisterMsgHandler(msgType, pid)
	}

	chainService.apis = []app.API{
		app.API{
			Namespace: "chain",
			Version:   "1.0",
			Service: &ChainApi{
				chainService: chainService,
				dbService:    chainService.DatabaseService,
			},
			Public: true,
		},
	}
	return nil
}

func (chainService *ChainService) Start(executeContext *app.ExecuteContext) error {
	chainService.transactionPool.Start(&chainService.newBlockFeed)
	go chainService.synchronise()
	return nil
}

func (chainService *ChainService) Stop(executeContext *app.ExecuteContext) error {
	return nil
}

func (chainService *ChainService) SendTransaction(tx *chainTypes.Transaction) error {
	//if id, err := tx.TxId(); err == nil {
	//	ForwardTransaction(id)
	//}

	err := chainService.ValidateTransaction(tx)
	if err != nil {
		return err
	}
	//TODO validate transaction
	error := chainService.transactionPool.AddTransaction(tx)
	if error == nil {
		chainService.P2pServer.Broadcast(tx)
	}

	return error
}

func (chainService *ChainService) sendBlock(block *chainTypes.Block) {
	chainService.P2pServer.Broadcast(block)
}

func (chainService *ChainService) blockExists(blockHash *crypto.Hash) bool {
	return chainService.Index.HaveBlock(blockHash)
}

func (chainService *ChainService) GenerateBlock(leaderKey string) (*chainTypes.Block, error) {
	chainService.DatabaseService.BeginTransaction()
	defer chainService.DatabaseService.Discard()

	height := chainService.BestChain.Height() + 1
	txs := chainService.transactionPool.GetPending(BlockGasLimit)

	finalTxs := make([]*chainTypes.Transaction, 0, len(txs))
	gasUsed := new(big.Int)
	for _, tx := range txs {
		g, _, err := chainService.executeTransaction(tx)
		if err == nil {
			finalTxs = append(finalTxs, tx)
			gasUsed.Add(gasUsed, g)
		}
	}

	timestamp := time.Now().Unix()
	previousHash := chainService.BestChain.Tip().Hash
	stateRoot := chainService.DatabaseService.GetStateRoot()
	merkleRoot := chainService.deriveMerkleRoot(finalTxs)

	block := &chainTypes.Block{
		Header: &chainTypes.BlockHeader{
			Version:      common.Version,
			PreviousHash: previousHash,
			ChainId:      chainService.chainId,
			GasLimit:     BlockGasLimit,
			GasUsed:      gasUsed,
			Timestamp:    timestamp,
			StateRoot:    stateRoot,
			TxRoot:       merkleRoot,
			Height:       height,
			LeaderPubKey: leaderKey,
		},
		Data: &chainTypes.BlockData{
			TxCount: int32(len(finalTxs)),
			TxList:  finalTxs,
		},
	}
	return block, nil
}

func (chainService *ChainService) GetTxHashes(ts []*chainTypes.Transaction) ([][]byte, error) {
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

func (chainService *ChainService) Attach() (*rpc.Client, error) {
	chainService.lock.RLock()
	defer chainService.lock.RUnlock()

	return rpc.DialInProc(chainService.RpcService.IpcHandler), nil
}

func (chainService *ChainService) RootChain() app.ChainIdType {
	return rootChain
}

func (chainService *ChainService) GenesisBlock() (*chainTypes.Block, error) {
	chainService.DatabaseService.BeginTransaction()
	defer chainService.DatabaseService.Discard()

	merkle := chainService.DatabaseService.NewMerkle([][]byte{})
	merkleRoot := merkle.Root.Hash

	//TODO check genesis account
	/*
	b := common.Bytes(genesisPubkey)
	err := b.UnmarshalText(b)
	if err != nil {
		return nil
	}

	pubkey, err := secp256k1.ParsePubKey(b)
	if err != nil {
		return nil
	}
	*/

	return &chainTypes.Block{
		Header: &chainTypes.BlockHeader{
			Version:      common.Version,
			PreviousHash: &crypto.Hash{},
			GasLimit:     BlockGasLimit,
			GasUsed:      new(big.Int),
			Timestamp:    1545282765,
			StateRoot:    []byte{0},
			TxRoot:       merkleRoot,
			Height:       0,
			LeaderPubKey: "",
			MinorPubKeys: []string{},
		},
		Data: &chainTypes.BlockData{
			TxCount: 0,
			TxList:  []*chainTypes.Transaction{},
		},
	}, nil
}

// AccumulateRewards credits,The leader gets half of the reward and other ,Other participants get the average of the other half
func (chainService *ChainService) accumulateRewards(b *chainTypes.Block, totalGasBalance *big.Int) error {
	if b.Header.LeaderPubKey == "" {
		return errors.New("invalidate leader account")
	}
	reward := new(big.Int).SetUint64(uint64(Rewards))
	r := new(big.Int)
	r = r.Div(reward, new(big.Int).SetInt64(2))
	r.Add(r, totalGasBalance)
	chainService.DatabaseService.AddBalance(b.Header.LeaderPubKey, r, true)

	num := len(b.Header.MinorPubKeys)
	for _, member := range b.Header.MinorPubKeys {
		r.Div(reward, new(big.Int).SetInt64(int64(num*2)))
		if member == "" {
			return errors.New("invalidate leader account")
		}
		chainService.DatabaseService.AddBalance(member, r, true)
	}
	return  nil
}

func (chainService *ChainService) SubscribeSyncBlockEvent(subchan chan event.SyncBlockEvent) event.Subscription {
	return chainService.syncBlockEvent.Subscribe(subchan)
}

func (chainService *ChainService) GetTransactionCount(accountName string) int64 {
	return chainService.transactionPool.GetTransactionCount(accountName)
}

func (chainService *ChainService) GetBlocksFrom(start, size int64) ([]*chainTypes.Block, error) {
	blocks := []*chainTypes.Block{}
	for i := start; i < start+size; i++ {
		node := chainService.BestChain.NodeByHeight(i)
		if node == nil {
			continue
		}
		block, err := chainService.DatabaseService.GetBlock(node.Hash)
		if err != nil {
			return nil, err
		}
		blocks = append(blocks, block)
	}
	return blocks, nil
}

func (chainService *ChainService) GetHighestBlock() (*chainTypes.Block, error) {
	heighestBlockBode := chainService.BestChain.Tip()
	block, err := chainService.DatabaseService.GetBlock(heighestBlockBode.Hash)
	if err != nil {
		return nil, err
	}
	return block, nil
}
