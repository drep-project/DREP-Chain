package service

import (
	"math/big"
	"math/rand"
	"sync"
	"time"

	"github.com/drep-project/drep-chain/chain/params"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/drep-project/drep-chain/app"
	"gopkg.in/urfave/cli.v1"

	"github.com/drep-project/binary"
	"github.com/drep-project/drep-chain/chain/txpool"
	"github.com/drep-project/drep-chain/common"
	"github.com/drep-project/drep-chain/common/event"
	"github.com/drep-project/drep-chain/crypto"
	"github.com/drep-project/drep-chain/crypto/secp256k1"
	"github.com/drep-project/drep-chain/crypto/sha3"
	"github.com/drep-project/drep-chain/database"
	"github.com/drep-project/drep-chain/network/p2p"
	"github.com/drep-project/drep-chain/pkgs/evm"
	"github.com/drep-project/drep-chain/rpc"

	"github.com/drep-project/dlog"
	chainTypes "github.com/drep-project/drep-chain/chain/types"
	p2pService "github.com/drep-project/drep-chain/network/service"
	rpc2 "github.com/drep-project/drep-chain/pkgs/rpc"
)

var (
	rootChain           app.ChainIdType
	DefaultOracleConfig = chainTypes.OracleConfig{
		Blocks:     20,
		Default:    big.NewInt(params.GWei).Uint64(),
		Percentile: 60,
		MaxPrice:   big.NewInt(500 * params.GWei).Uint64(),
	}
	DefaultChainConfig = &chainTypes.ChainConfig{
		RemotePort: 55556,
		ChainId:    app.ChainIdType{},
		GasPrice:   DefaultOracleConfig,
		GenesisPK:  "0x03177b8e4ef31f4f801ce00260db1b04cc501287e828692a404fdbc46c7ad6ff26",
	}
	span = uint64(params.MaxGasLimit / 360)
)

type ChainService struct {
	RpcService      *rpc2.RpcService          `service:"rpc"`
	P2pServer       p2pService.P2P            `service:"p2p"`
	DatabaseService *database.DatabaseService `service:"database"`
	VmService       evm.Vm                    `service:"vm"`
	transactionPool *txpool.TransactionPool
	isRelay         bool
	apis            []app.API

	stateProcessor *StateProcessor

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

	Index         *chainTypes.BlockIndex
	BestChain     *chainTypes.ChainView
	stateLock     sync.RWMutex
	StateSnapshot *ChainState

	Config       *chainTypes.ChainConfig
	pid          *actor.PID
	genesisBlock *chainTypes.Block
	//Events related to sync blocks
	syncBlockEvent event.Feed
	syncMut        sync.Mutex

	//提供新块订阅
	NewBlockFeed    event.Feed
	DetachBlockFeed event.Feed

	//从远端接收块头hash组
	headerHashCh chan []*syncHeaderHash

	//从远端接收到块
	blocksCh chan []*chainTypes.Block

	//所有需要同步的任务列表
	allTasks *heightSortedMap

	//正在同步中的任务列表，如果对应的块未到，会重新发布请求的
	pendingSyncTasks map[crypto.Hash]uint64
	taskTxsCh        chan tasksTxsSync

	//与此模块通信的所有Peer
	peersInfo map[string]*chainTypes.PeerInfo
	newPeerCh chan *chainTypes.PeerInfo

	gpo  *Oracle
	quit chan struct{}
}

type syncHeaderHash struct {
	headerHash *crypto.Hash
	height     uint64
}

type ChainState struct {
	chainTypes.BestState
	db *database.Database
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
	chainService.headerHashCh = make(chan []*syncHeaderHash)
	chainService.blocksCh = make(chan []*chainTypes.Block)
	chainService.allTasks = newHeightSortedMap()
	chainService.pendingSyncTasks = make(map[crypto.Hash]uint64)
	chainService.peersInfo = make(map[string]*chainTypes.PeerInfo)
	chainService.newPeerCh = make(chan *chainTypes.PeerInfo, maxLivePeer)
	chainService.taskTxsCh = make(chan tasksTxsSync, maxLivePeer)
	chainService.stateProcessor = NewStateProcessor(chainService)
	chainService.gpo = NewOracle(chainService, chainService.Config.GasPrice)

	chainService.genesisBlock = chainService.GetGenisiBlock(chainService.Config.GenesisPK)
	hash := chainService.genesisBlock.Header.Hash()
	if !chainService.DatabaseService.HasBlock(hash) {
		chainService.genesisBlock, err = chainService.ProcessGenesisBlock(chainService.Config.GenesisPK)
		err = chainService.createChainState()
		if err != nil {
			return err
		}
	}
	chainService.InitStates()
	chainService.transactionPool = txpool.NewTransactionPool(chainService.StateSnapshot.db)

	chainService.P2pServer.AddProtocols([]p2p.Protocol{
		p2p.Protocol{
			Name:   "chainService",
			Length: chainTypes.NumberOfMsg,
			Run: func(peer *p2p.Peer, rw p2p.MsgReadWriter) error {
				if len(chainService.peersInfo) >= maxLivePeer {
					return ErrEnoughPeer
				}
				pi := chainTypes.NewPeerInfo(peer, rw)
				chainService.peersInfo[peer.IP()] = pi
				defer delete(chainService.peersInfo, peer.IP())
				return chainService.receiveMsg(pi, rw)
			},
		},
	})

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
	chainService.transactionPool.Start(&chainService.NewBlockFeed)
	go chainService.synchronise()
	go chainService.syncTxs()
	return nil
}

func (chainService *ChainService) Stop(executeContext *app.ExecuteContext) error {
	if chainService.quit != nil {
		close(chainService.quit)
	}

	return nil
}

func (chainService *ChainService) SendTransaction(tx *chainTypes.Transaction) error {
	chainService.stateLock.Lock()
	db := chainService.StateSnapshot.db
	chainService.stateLock.Unlock()
	err := chainService.VerifyTransaction(db, tx)

	if err != nil {
		return err
	}
	err = chainService.transactionPool.AddTransaction(tx)
	if err != nil {
		return err
	} else {
		chainService.BroadcastTx(chainTypes.MsgTypeTransaction, tx, true)
	}
	return nil
}

func (chainService *ChainService) BroadcastBlock(msgType int32, block *chainTypes.Block, isLocal bool) {
	for _, peer := range chainService.peersInfo {
		b := peer.KnownBlock(block)
		if !b {
			if !isLocal {
				//收到远端来的消息，仅仅广播给1/3的peer
				rd := rand.Intn(broadcastRatio)
				if rd > 1 {
					continue
				}
			}
			peer.MarkBlock(block)
			chainService.P2pServer.Send(peer.GetMsgRW(), uint64(msgType), block)
		}
	}
}

func (chainService *ChainService) BroadcastTx(msgType int32, tx *chainTypes.Transaction, isLocal bool) {
	go func() {
		for _, peer := range chainService.peersInfo {
			b := peer.KnownTx(tx)
			if !b {
				if !isLocal {
					//收到远端来的消息，仅仅广播给1/3的peer
					rd := rand.Intn(broadcastRatio)
					if rd > 1 {
						continue
					}
				}
				peer.MarkTx(tx)
				chainService.P2pServer.Send(peer.GetMsgRW(), uint64(msgType), chainTypes.Transactions{*tx})
			}
		}
	}()
}

func (chainService *ChainService) blockExists(blockHash *crypto.Hash) bool {
	return chainService.Index.HaveBlock(blockHash)
}

func (chainService *ChainService) GenerateBlock(leaderKey *secp256k1.PublicKey) (*chainTypes.Block, error) {
	db := chainService.DatabaseService.BeginTransaction()
	defer db.Discard()
	parent, err := chainService.GetHighestBlock()
	if err != nil {
		return nil, err
	}
	newGasLimit := chainService.CalcGasLimit(parent.Header, params.MinGasLimit, params.MaxGasLimit)
	height := chainService.BestChain.Height() + 1
	txs := chainService.transactionPool.GetPending(newGasLimit)

	previousHash := chainService.BestChain.Tip().Hash
	timestamp := uint64(time.Now().Unix())

	blockHeader := &chainTypes.BlockHeader{
		Version:      common.Version,
		PreviousHash: *previousHash,
		ChainId:      chainService.chainId,
		GasLimit:     *newGasLimit,
		Timestamp:    timestamp,
		Height:       height,
		LeaderPubKey: *leaderKey,
	}

	finalTxs := make([]*chainTypes.Transaction, 0, len(txs))
	gasUsed := new(big.Int)
	gp := new(GasPool).AddGas(blockHeader.GasLimit.Uint64())
	stopchanel := make(chan struct{})
	time.AfterFunc(time.Second*5, func() {
		stopchanel <- struct{}{}
	})

SELECT_TX:
	for _, t := range txs {
		select {
		case <-stopchanel:
			break SELECT_TX
		default:
			g, _, err := chainService.executeTransaction(db, t, gp, blockHeader)
			if err == nil {
				finalTxs = append(finalTxs, t)
				gasUsed.Add(gasUsed, g)
			} else {
				if err.Error() == ErrReachGasLimit.Error() {
					break SELECT_TX
				} else {

					//TODO err or continue
					dlog.Warn("generate block", "exe tx err", err)
					continue
					//  return nil, err
				}
			}
		}
	}

	blockHeader.GasUsed = *new(big.Int).SetUint64(gasUsed.Uint64())
	blockHeader.StateRoot = db.GetStateRoot()
	blockHeader.TxRoot = chainService.deriveMerkleRoot(finalTxs)

	block := &chainTypes.Block{
		Header: blockHeader,
		Data: &chainTypes.BlockData{
			TxCount: uint64(len(finalTxs)),
			TxList:  finalTxs,
		},
	}
	return block, nil
}

func (chainService *ChainService) GetTxHashes(ts []*chainTypes.Transaction) ([][]byte, error) {
	txHashes := make([][]byte, len(ts))
	for i, tx := range ts {
		b, err := binary.Marshal(tx.Data)
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

// AccumulateRewards credits,The leader gets half of the reward and other ,Other participants get the average of the other half
func (chainService *ChainService) accumulateRewards(db *database.Database, b *chainTypes.Block, totalGasBalance *big.Int) {
	reward := new(big.Int).SetUint64(uint64(Rewards))
	leaderAddr := crypto.PubKey2Address(&b.Header.LeaderPubKey)

	r := new(big.Int)
	r = r.Div(reward, new(big.Int).SetInt64(2))
	r.Add(r, totalGasBalance)
	db.AddBalance(&leaderAddr, r)

	num := len(b.Header.MinorPubKeys)
	for _, memberPK := range b.Header.MinorPubKeys {
		if !memberPK.IsEqual(&b.Header.LeaderPubKey) {
			memberAddr := crypto.PubKey2Address(&memberPK)
			r.Div(reward, new(big.Int).SetInt64(int64(num*2)))
			db.AddBalance(&memberAddr, r)
		}
	}
}

func (chainService *ChainService) SubscribeSyncBlockEvent(subchan chan event.SyncBlockEvent) event.Subscription {
	return chainService.syncBlockEvent.Subscribe(subchan)
}

func (chainService *ChainService) GetTransactionCount(addr *crypto.CommonAddress) uint64 {
	return chainService.transactionPool.GetTransactionCount(addr)
}

func (chainService *ChainService) GetBlocksFrom(start, size uint64) ([]*chainTypes.Block, error) {
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

func (chainService *ChainService) GetBlockByHash(hash *crypto.Hash) (*chainTypes.Block, error) {
	block, err := chainService.DatabaseService.GetBlock(hash)
	if err != nil {
		return nil, err
	}
	return block, nil
}
func (chainService *ChainService) GetBlockHeaderByHash(hash *crypto.Hash) (*chainTypes.BlockHeader, error) {
	blockNode, ok := chainService.Index.Index[*hash]
	if !ok {
		return nil, ErrBlockNotFound
	}
	blockHeader := blockNode.Header()
	return &blockHeader, nil
}

func (chainService *ChainService) GetHeader(hash crypto.Hash, number uint64) *chainTypes.BlockHeader {
	header, _ := chainService.GetBlockHeaderByHash(&hash)
	return header
}

func (chainService *ChainService) GetBlockByHeight(number uint64) (*chainTypes.Block, error) {
	blockNode := chainService.BestChain.NodeByHeight(number)
	return chainService.GetBlockByHash(blockNode.Hash)
}

func (chainService *ChainService) GetBlockHeaderByHeight(number uint64) (*chainTypes.BlockHeader, error) {
	blockNode := chainService.BestChain.NodeByHeight(number)
	if blockNode == nil {
		return nil, ErrBlockNotFound
	}
	header := blockNode.Header()
	return &header, nil
}

//180000000/360
func (chainService *ChainService) CalcGasLimit(parent *chainTypes.BlockHeader, gasFloor, gasCeil uint64) *big.Int {
	limit := uint64(0)
	if parent.GasLimit.Uint64()*2/3 > parent.GasUsed.Uint64() {
		limit = parent.GasLimit.Uint64() - span
	} else {
		limit = parent.GasLimit.Uint64() + span
	}

	if limit < params.MinGasLimit {
		limit = params.MinGasLimit
	}
	// If we're outside our allowed gas range, we try to hone towards them
	if limit < gasFloor {
		limit = gasFloor
	} else if limit > gasCeil {
		limit = gasCeil
	}
	return new(big.Int).SetUint64(limit)
}

func (chainService *ChainService) GetPoolTransactions(addr *crypto.CommonAddress) []chainTypes.Transactions {
	return chainService.transactionPool.GetTransactions(addr)
}
func (chainService *ChainService) GetPoolMiniPendingNonce(addr *crypto.CommonAddress) uint64 {
	return chainService.transactionPool.GetMiniPendingNonce(addr)
}
