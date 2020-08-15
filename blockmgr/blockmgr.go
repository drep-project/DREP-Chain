package blockmgr

import (
	"math/big"
	"math/rand"
	"path"
	"sync"

	"github.com/drep-project/DREP-Chain/chain/store"
	"github.com/drep-project/DREP-Chain/common/trie"

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
	// DefaultOracleConfig define default config of oracle
	DefaultOracleConfig = OracleConfig{
		Blocks:     20,
		Default:    30000,
		Percentile: 60,
		MaxPrice:   big.NewInt(500 * params.GWei).Uint64(),
	}
	// DefaultChainConfig define default config of chain
	DefaultChainConfig = &BlockMgrConfig{
		GasPrice:    DefaultOracleConfig,
		JournalFile: "txpool/txs",
	}
	span = uint64(params.MaxGasLimit / 360)
	_    = IBlockMgr((*BlockMgr)(nil)) //compile check
)

// IBlockMgr interface
type IBlockMgr interface {
	app.Service
	IBlockMgrPool
	IBlockBlockGenerator
	IBlockNotify
	ISendMessage
}

// IBlockMgrPool interface
type IBlockMgrPool interface {
	//query tx pool message
	GetTransactionCount(addr *crypto.CommonAddress) uint64
	GetPoolTransactions(addr *crypto.CommonAddress) []types.Transactions
	GetPoolMiniPendingNonce(addr *crypto.CommonAddress) uint64
	GetTxInPool(hash string) (*types.Transaction, error)
}

// IBlockBlockGenerator interface
type IBlockBlockGenerator interface {
	//generate block template
	GenerateTemplate(trieStore store.StoreInterface, leaderAddr crypto.CommonAddress, blockInterval int) (*types.Block, *big.Int, error)
}

// IBlockNotify interface
type IBlockNotify interface {
	//notify
	SubscribeSyncBlockEvent(subchan chan event.SyncBlockEvent) event.Subscription
	NewTxFeed() *event.Feed
}

// ISendMessage interface
type ISendMessage interface {
	// send
	SendTransaction(tx *types.Transaction, islocal bool) error
	BroadcastBlock(msgType int32, block *types.Block, isLocal bool)
	BroadcastTx(msgType int32, tx *types.Transaction, isLocal bool)
}

// BlockMgr is an overarching block manager that can communicate with various
// backends for preducing blocks.
type BlockMgr struct {
	// ChainService define service interface of chain
	ChainService chain.ChainServiceInterface `service:"chain"`
	// RpcService define service interface of rpc
	RpcService *rpc2.RpcService `service:"rpc"`
	// P2pServer define interface of p2p
	P2pServer p2pService.P2P `service:"p2p"`
	// P2pServer define service interface of database
	DatabaseService *database.DatabaseService `service:"database"`
	transactionPool *txpool.TransactionPool
	chainStore      *chain.ChainStore
	apis            []app.API

	lock   sync.RWMutex
	Config *BlockMgrConfig

	//Events related to sync blocks
	syncBlockEvent event.Feed
	syncMut        sync.Mutex

	//Receive the bulk hash group from the remote
	headerHashCh chan []*syncHeaderHash

	//The block is received from the far end
	blocksCh chan []*types.Block

	//List of all tasks that need to be synchronized
	allTasks *heightSortedMap

	//The list of tasks being synchronized will be reissued if the corresponding block does not arrive
	pendingSyncTasks sync.Map //map[*time.Timer]map[crypto.Hash]uint64
	taskTxsCh        chan tasksTxsSync
	syncTimerCh      chan *time.Timer
	state            event.EventType

	//All peers that communicate with this module
	//peersInfo map[string]types.PeerInfoInterface
	peersInfo sync.Map //key: node.ID(),value PeerInfo

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

// Name return package name
func (blockMgr *BlockMgr) Name() string {
	return "blockmgr"
}

// Api return an array of block management api
func (blockMgr *BlockMgr) Api() []app.API {
	return blockMgr.apis
}

// CommandFlags return an array interface of flag
func (blockMgr *BlockMgr) CommandFlags() ([]cli.Command, []cli.Flag) {
	return nil, []cli.Flag{}
}

// NewBlockMgr init all need of block management
func NewBlockMgr(config *BlockMgrConfig, homeDir string, cs chain.ChainServiceInterface, p2pservice p2pService.P2P) *BlockMgr {
	blockMgr := &BlockMgr{}
	blockMgr.Config = config
	blockMgr.ChainService = cs
	blockMgr.P2pServer = p2pservice

	blockMgr.headerHashCh = make(chan []*syncHeaderHash)
	blockMgr.blocksCh = make(chan []*types.Block)
	blockMgr.allTasks = newHeightSortedMap()
	//blockMgr.pendingSyncTasks = make(map[*time.Timer]map[crypto.Hash]uint64)
	blockMgr.state = event.StopSyncBlock
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
				blockMgr.peersInfo.Store(peer.ID().String(), pi)

				defer blockMgr.peersInfo.Delete(peer.ID().String()) // (blockMgr.peersInfo, peer.IP())
				return blockMgr.receiveMsg(pi, rw)
			},
		},
	})

	blockMgr.apis = []app.API{
		app.API{
			Namespace: "blockmgr",
			Version:   "1.0",
			Service: &BlockMgrAPI{
				blockMgr:  blockMgr,
				dbService: blockMgr.DatabaseService,
			},
			Public: true,
		},
	}
	return blockMgr
}

// Init function init block from initial config.
func (blockMgr *BlockMgr) Init(executeContext *app.ExecuteContext) error {
	blockMgr.headerHashCh = make(chan []*syncHeaderHash)
	blockMgr.blocksCh = make(chan []*types.Block)
	blockMgr.allTasks = newHeightSortedMap()
	blockMgr.syncTimerCh = make(chan *time.Timer, 1)
	blockMgr.state = event.StopSyncBlock
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
				//blockMgr.lock.Lock()
				//defer blockMgr.lock.Unlock()

				if getPeersCount(blockMgr.peersInfo) >= maxLivePeer {
					return ErrEnoughPeer
				}
				pi := types.NewPeerInfo(peer, rw)
				blockMgr.peersInfo.Store(peer.ID().String(), pi)

				defer blockMgr.peersInfo.Delete(peer.ID().String())
				return blockMgr.receiveMsg(pi, rw)
			},
		},
	})

	blockMgr.apis = []app.API{
		app.API{
			Namespace: "blockmgr",
			Version:   "1.0",
			Service: &BlockMgrAPI{
				blockMgr:  blockMgr,
				dbService: blockMgr.DatabaseService,
			},
			Public: true,
		},
	}
	return nil
}

// Start syn block and transactions.
func (blockMgr *BlockMgr) Start(executeContext *app.ExecuteContext) error {
	blockMgr.transactionPool.Start(blockMgr.ChainService.NewBlockFeed(), blockMgr.ChainService.BestChain().Tip().StateRoot)
	go blockMgr.synchronise()
	go blockMgr.syncTxs()
	return nil
}

// Stop blockchain.
func (blockMgr *BlockMgr) Stop(executeContext *app.ExecuteContext) error {
	if blockMgr.quit != nil {
		close(blockMgr.quit)
	}
	return nil
}

// GetTransactionCount gets the total number of transactions, that is, the nonce corresponding to the address.
func (blockMgr *BlockMgr) GetTransactionCount(addr *crypto.CommonAddress) uint64 {
	return blockMgr.transactionPool.GetTransactionCount(addr)
}

// SendTransaction adds local signed transaction and broadcast it.
func (blockMgr *BlockMgr) SendTransaction(tx *types.Transaction, islocal bool) error {
	//from, err := tx.From()
	//nonce := blockMgr.transactionPool.GetTransactionCount(from)
	//if nonce > tx.Nonce() {
	//	return fmt.Errorf("SendTransaction local nonce:%d != comming tx nonce:%d", nonce, tx.Nonce())
	//}
	err := blockMgr.verifyTransaction(tx)
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

// BroadcastBlock broadcasts block until receive more than 2/3 of peers.
func (blockMgr *BlockMgr) BroadcastBlock(msgType int32, block *types.Block, isLocal bool) {
	blockMgr.peersInfo.Range(func(key, value interface{}) bool {
		peer := value.(types.PeerInfoInterface)
		b := peer.KnownBlock(block)
		if !b {
			if !isLocal {
				//Receive a message from the remote end and only broadcast it to 2/3 of peers
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

// BroadcastTx broadcasts transaction until receive more than 2/3 of peers.
func (blockMgr *BlockMgr) BroadcastTx(msgType int32, tx *types.Transaction, isLocal bool) {
	go func() {

		blockMgr.peersInfo.Range(func(key, value interface{}) bool {
			peer := value.(types.PeerInfoInterface)
			b := peer.KnownTx(tx)
			if !b {
				if !isLocal {
					//Receive a message from the remote end and only broadcast it to 2/3 of peers
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

// GetPoolTransactions gets all the trades in the current pool.
func (blockMgr *BlockMgr) GetPoolTransactions(addr *crypto.CommonAddress) []types.Transactions {
	return blockMgr.transactionPool.GetTransactions(addr)
}

// GetPoolMiniPendingNonce gets the smallest nonce in the Pending queue
func (blockMgr *BlockMgr) GetPoolMiniPendingNonce(addr *crypto.CommonAddress) uint64 {
	return blockMgr.transactionPool.GetMiniPendingNonce(addr)
}

// GetTxInPool gets transactions in the trading pool.
func (blockMgr *BlockMgr) GetTxInPool(hash string) (*types.Transaction, error) {
	return blockMgr.transactionPool.GetTxInPool(hash)
}

// SubscribeSyncBlockEvent gets a channel from the feed.
func (blockMgr *BlockMgr) SubscribeSyncBlockEvent(subchan chan event.SyncBlockEvent) event.Subscription {
	return blockMgr.syncBlockEvent.Subscribe(subchan)
}

// NewTxFeed gets transaction feed in the trading pool.
func (blockMgr *BlockMgr) NewTxFeed() *event.Feed {
	return blockMgr.transactionPool.NewTxFeed()
}

// DefaultConfig gets default config of blockchain.
func (blockMgr *BlockMgr) DefaultConfig(netType params.NetType) *BlockMgrConfig {
	return DefaultChainConfig
}
