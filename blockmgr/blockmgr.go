package blockmgr

import (
	"fmt"
	"github.com/drep-project/DREP-Chain/chain/store"
	"github.com/drep-project/DREP-Chain/common/trie"
	"math/big"
	"math/rand"
	"path"
	"sync"

	"github.com/drep-project/DREP-Chain/params"

	"github.com/drep-project/DREP-Chain/app"
	"gopkg.in/urfave/cli.v1"

	"github.com/drep-project/DREP-Chain/blockmgr/txpool"
	"github.com/drep-project/DREP-Chain/chain"
	"github.com/drep-project/DREP-Chain/common/event"
	"github.com/drep-project/DREP-Chain/crypto"
	"github.com/drep-project/DREP-Chain/database"
	"github.com/drep-project/DREP-Chain/network/p2p"
	p2pService "github.com/drep-project/DREP-Chain/network/service"
	"github.com/drep-project/DREP-Chain/types"

	"time"

	rpc2 "github.com/drep-project/DREP-Chain/pkgs/rpc"
)

var (
	rootChain           types.ChainIdType
	DefaultOracleConfig = OracleConfig{
		Blocks:     20,
		Default:    30000,
		Percentile: 60,
		MaxPrice:   big.NewInt(500 * params.GWei).Uint64(),
	}
	DefaultChainConfig = &BlockMgrConfig{
		GasPrice:    DefaultOracleConfig,
		JournalFile: "txpool/txs",
	}
	span = uint64(params.MaxGasLimit / 360)
	_    = IBlockMgr((*BlockMgr)(nil)) //compile check
)

type IBlockMgr interface {
	app.Service
	IBlockMgrPool
	IBlockBlockGenerator
	IBlockNotify
	ISendMessage
}

type IBlockMgrPool interface {
	//query tx pool message
	GetTransactionCount(addr *crypto.CommonAddress) uint64
	GetPoolTransactions(addr *crypto.CommonAddress) []types.Transactions
	GetPoolMiniPendingNonce(addr *crypto.CommonAddress) uint64
	GetTxInPool(hash string) (*types.Transaction, error)
}

type IBlockBlockGenerator interface {
	//generate block template
	GenerateTemplate(trieStore store.StoreInterface, leaderAddr crypto.CommonAddress) (*types.Block, *big.Int, error)
}

type IBlockNotify interface {
	//notify
	SubscribeSyncBlockEvent(subchan chan event.SyncBlockEvent) event.Subscription
	NewTxFeed() *event.Feed
}

type ISendMessage interface {
	// send
	SendTransaction(tx *types.Transaction, islocal bool) error
	BroadcastBlock(msgType int32, block *types.Block, isLocal bool)
	BroadcastTx(msgType int32, tx *types.Transaction, isLocal bool)
}

type BlockMgr struct {
	ChainService    chain.ChainServiceInterface `service:"chain"`
	RpcService      *rpc2.RpcService            `service:"rpc"`
	P2pServer       p2pService.P2P              `service:"p2p"`
	DatabaseService *database.DatabaseService   `service:"database"`
	transactionPool *txpool.TransactionPool
	chainStore      *chain.ChainStore
	apis            []app.API

	lock   sync.RWMutex
	Config *BlockMgrConfig

	//Events related to sync blocks
	syncBlockEvent event.Feed
	syncMut        sync.Mutex

	//从远端接收块头hash组
	headerHashCh chan []*syncHeaderHash

	//从远端接收到块
	blocksCh chan []*types.Block

	//所有需要同步的任务列表
	allTasks *heightSortedMap

	//正在同步中的任务列表，如果对应的块未到，会重新发布请求的
	pendingSyncTasks sync.Map //map[*time.Timer]map[crypto.Hash]uint64
	taskTxsCh        chan tasksTxsSync
	syncTimerCh      chan *time.Timer

	//与此模块通信的所有Peer
	//peersInfo map[string]types.PeerInfoInterface
	peersInfo sync.Map

	newPeerCh chan *types.PeerInfo

	gpo  *Oracle
	quit chan struct{}
}

func getPeersCount(peerInfos sync.Map) int {
	count := 0
	peerInfos.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	return count
}

type syncHeaderHash struct {
	headerHash *crypto.Hash
	height     uint64
}

func (blockMgr *BlockMgr) Name() string {
	return "blockmgr"
}

func (blockMgr *BlockMgr) Api() []app.API {
	return blockMgr.apis
}

func (blockMgr *BlockMgr) CommandFlags() ([]cli.Command, []cli.Flag) {
	return nil, []cli.Flag{}
}

func NewBlockMgr(config *BlockMgrConfig, homeDir string, cs chain.ChainServiceInterface, p2pservice p2pService.P2P) *BlockMgr {
	blockMgr := &BlockMgr{}
	blockMgr.Config = config
	blockMgr.ChainService = cs
	blockMgr.P2pServer = p2pservice

	blockMgr.headerHashCh = make(chan []*syncHeaderHash)
	blockMgr.blocksCh = make(chan []*types.Block)
	blockMgr.allTasks = newHeightSortedMap()
	//blockMgr.pendingSyncTasks = make(map[*time.Timer]map[crypto.Hash]uint64)
	blockMgr.syncTimerCh = make(chan *time.Timer, pendingTimerCount)
	//blockMgr.peersInfo = sync.Map{} //make(map[string]types.PeerInfoInterface)
	blockMgr.newPeerCh = make(chan *types.PeerInfo, maxLivePeer)
	blockMgr.taskTxsCh = make(chan tasksTxsSync, maxLivePeer)

	blockMgr.gpo = NewOracle(blockMgr.ChainService, blockMgr.Config.GasPrice)

	store, err := store.TrieStoreFromStore(blockMgr.DatabaseService.LevelDb(), trie.EmptyRoot[:])
	if err != nil {
		return nil
	}
	blockMgr.transactionPool = txpool.NewTransactionPool(store, path.Join(homeDir, blockMgr.Config.JournalFile))

	blockMgr.P2pServer.AddProtocols([]p2p.Protocol{
		p2p.Protocol{
			Name:   "blockMgr",
			Length: types.NumberOfMsg,
			Run: func(peer *p2p.Peer, rw p2p.MsgReadWriter) error {
				if getPeersCount(blockMgr.peersInfo) >= maxLivePeer {
					return ErrEnoughPeer
				}
				pi := types.NewPeerInfo(peer, rw)
				blockMgr.peersInfo.Store(peer.IP(), pi)

				defer blockMgr.peersInfo.Delete(peer.IP()) // delete(blockMgr.peersInfo, peer.IP())
				return blockMgr.receiveMsg(pi, rw)
			},
		},
	})

	blockMgr.apis = []app.API{
		app.API{
			Namespace: "blockmgr",
			Version:   "1.0",
			Service: &BlockMgrApi{
				blockMgr:  blockMgr,
				dbService: blockMgr.DatabaseService,
			},
			Public: true,
		},
	}
	return blockMgr
}

func (blockMgr *BlockMgr) Init(executeContext *app.ExecuteContext) error {
	blockMgr.headerHashCh = make(chan []*syncHeaderHash)
	blockMgr.blocksCh = make(chan []*types.Block)
	blockMgr.allTasks = newHeightSortedMap()
	//blockMgr.pendingSyncTasks = make(map[*time.Timer]map[crypto.Hash]uint64)
	blockMgr.syncTimerCh = make(chan *time.Timer, 1)
	//blockMgr.peersInfo = make(map[string]types.PeerInfoInterface)
	blockMgr.newPeerCh = make(chan *types.PeerInfo, maxLivePeer)
	blockMgr.taskTxsCh = make(chan tasksTxsSync, maxLivePeer)

	blockMgr.gpo = NewOracle(blockMgr.ChainService, blockMgr.Config.GasPrice)

	store, err := store.TrieStoreFromStore(blockMgr.DatabaseService.LevelDb(), trie.EmptyRoot[:])
	if err != nil {
		return err
	}
	blockMgr.transactionPool = txpool.NewTransactionPool(store, path.Join(executeContext.CommonConfig.HomeDir, blockMgr.Config.JournalFile))
	blockMgr.chainStore = &chain.ChainStore{blockMgr.DatabaseService.LevelDb()}
	blockMgr.P2pServer.AddProtocols([]p2p.Protocol{
		p2p.Protocol{
			Name:   "blockMgr",
			Length: types.NumberOfMsg,
			Run: func(peer *p2p.Peer, rw p2p.MsgReadWriter) error {
				if getPeersCount(blockMgr.peersInfo) >= maxLivePeer {
					return ErrEnoughPeer
				}
				pi := types.NewPeerInfo(peer, rw)
				blockMgr.peersInfo.Store(peer.IP(), pi)

				defer blockMgr.peersInfo.Delete(peer.IP())
				return blockMgr.receiveMsg(pi, rw)
			},
		},
	})

	blockMgr.apis = []app.API{
		app.API{
			Namespace: "blockmgr",
			Version:   "1.0",
			Service: &BlockMgrApi{
				blockMgr:  blockMgr,
				dbService: blockMgr.DatabaseService,
			},
			Public: true,
		},
	}
	return nil
}

func (blockMgr *BlockMgr) Start(executeContext *app.ExecuteContext) error {
	blockMgr.transactionPool.Start(blockMgr.ChainService.NewBlockFeed(), blockMgr.ChainService.BestChain().Tip().StateRoot)
	go blockMgr.synchronise()
	go blockMgr.syncTxs()
	return nil
}

func (blockMgr *BlockMgr) Stop(executeContext *app.ExecuteContext) error {
	if blockMgr.quit != nil {
		close(blockMgr.quit)
	}
	return nil
}

func (blockMgr *BlockMgr) GetTransactionCount(addr *crypto.CommonAddress) uint64 {
	return blockMgr.transactionPool.GetTransactionCount(addr)
}

func (blockMgr *BlockMgr) SendTransaction(tx *types.Transaction, islocal bool) error {
	from, err := tx.From()
	nonce := blockMgr.transactionPool.GetTransactionCount(from)
	if nonce > tx.Nonce() {
		return fmt.Errorf("error nounce db nonce:%d != %d", nonce, tx.Nonce())
	}
	err = blockMgr.verifyTransaction(tx)

	if err != nil {
		return err
	}
	err = blockMgr.transactionPool.AddTransaction(tx, islocal)
	if err != nil {
		return err
	}

	blockMgr.BroadcastTx(types.MsgTypeTransaction, tx, true)

	return nil
}

func (blockMgr *BlockMgr) BroadcastBlock(msgType int32, block *types.Block, isLocal bool) {
	blockMgr.peersInfo.Range(func(key, value interface{}) bool {
		peer := value.(types.PeerInfoInterface)
		b := peer.KnownBlock(block)
		if !b {
			if !isLocal {
				//收到远端来的消息，仅仅广播给2/3的peer
				rd := rand.Intn(broadcastRatio)
				if rd > 2 {
					return false
				}
			}
			peer.MarkBlock(block)
			blockMgr.P2pServer.Send(peer.GetMsgRW(), uint64(msgType), block)
		}
		return true
	})
}

func (blockMgr *BlockMgr) BroadcastTx(msgType int32, tx *types.Transaction, isLocal bool) {
	go func() {

		blockMgr.peersInfo.Range(func(key, value interface{}) bool {
			peer := value.(types.PeerInfoInterface)
			b := peer.KnownTx(tx)
			if !b {
				if !isLocal {
					//收到远端来的消息，仅仅广播给2/3的peer
					rd := rand.Intn(broadcastRatio)
					if rd > 2 {
						return false
					}
				}

				peer.MarkTx(tx)
				blockMgr.P2pServer.Send(peer.GetMsgRW(), uint64(msgType), []*types.Transaction{tx})
			}
			return true
		})
	}()
}

func (blockMgr *BlockMgr) GetPoolTransactions(addr *crypto.CommonAddress) []types.Transactions {
	return blockMgr.transactionPool.GetTransactions(addr)
}

func (blockMgr *BlockMgr) GetPoolMiniPendingNonce(addr *crypto.CommonAddress) uint64 {
	return blockMgr.transactionPool.GetMiniPendingNonce(addr)
}

func (blockMgr *BlockMgr) GetTxInPool(hash string) (*types.Transaction, error) {
	return blockMgr.transactionPool.GetTxInPool(hash)
}

func (blockMgr *BlockMgr) SubscribeSyncBlockEvent(subchan chan event.SyncBlockEvent) event.Subscription {
	return blockMgr.syncBlockEvent.Subscribe(subchan)
}

func (blockMgr *BlockMgr) NewTxFeed() *event.Feed {
	return blockMgr.transactionPool.NewTxFeed()
}

func (blockMgr *BlockMgr) DefaultConfig() *BlockMgrConfig {
	return DefaultChainConfig
}
