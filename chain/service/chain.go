package service

import (
	"fmt"
	"math/big"
	"math/rand"
	"sync"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/drep-project/drep-chain/app"
	"gopkg.in/urfave/cli.v1"

	"github.com/drep-project/drep-chain/chain/txpool"
	"github.com/drep-project/drep-chain/network/p2p"
	"github.com/drep-project/drep-chain/pkgs/evm"
	"github.com/drep-project/drep-chain/common"
	"github.com/drep-project/drep-chain/common/event"
	"github.com/drep-project/drep-chain/crypto"
	"github.com/drep-project/drep-chain/crypto/secp256k1"
	"github.com/drep-project/drep-chain/crypto/sha3"
	"github.com/drep-project/drep-chain/database"
	"github.com/drep-project/drep-chain/rpc"
	"github.com/drep-project/dlog"
	"github.com/drep-project/binary"

	chainTypes "github.com/drep-project/drep-chain/chain/types"
	p2pService "github.com/drep-project/drep-chain/network/service"
	rpc2 "github.com/drep-project/drep-chain/pkgs/rpc"
)

var (
	rootChain          app.ChainIdType
	DefaultChainConfig = &chainTypes.ChainConfig{
		RemotePort: 55556,
		ChainId:    app.ChainIdType{},
		GenesisPK:  "0x03177b8e4ef31f4f801ce00260db1b04cc501287e828692a404fdbc46c7ad6ff26",
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

	prvKey *secp256k1.PrivateKey

	Index         *chainTypes.BlockIndex
	BestChain     *chainTypes.ChainView
	stateLock     sync.RWMutex
	StateSnapshot *chainTypes.BestState

	Config       *chainTypes.ChainConfig
	pid          *actor.PID
	genesisBlock *chainTypes.Block
	//Events related to sync blocks
	syncBlockEvent event.Feed
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
	pendingSyncTasks map[crypto.Hash]uint64
	taskTxsCh chan tasksTxsSync

	//与此模块通信的所有Peer
	peersInfo map[string]*chainTypes.PeerInfo
	newPeerCh chan *chainTypes.PeerInfo

	quit chan struct{}
}

type syncHeaderHash struct {
	headerHash *crypto.Hash
	height     uint64
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
	chainService.genesisBlock = chainService.GenesisBlock(chainService.Config.GenesisPK)
	chainService.taskTxsCh = make(chan tasksTxsSync, maxLivePeer)

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

	chainService.P2pServer.AddProtocols([]p2p.Protocol{
		p2p.Protocol{
			Name:   "chainService",
			Length: chainTypes.NumberOfMsg,
			Run: func(peer *p2p.Peer, rw p2p.MsgReadWriter) error {
				if len(chainService.peersInfo) >= maxLivePeer {
					return fmt.Errorf("enough peer")
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
	chainService.transactionPool.Start(&chainService.newBlockFeed)
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
	//TODO validate transaction
	error := chainService.transactionPool.AddTransaction(tx)
	if error == nil {
		chainService.BroadcastTx(chainTypes.MsgTypeTransaction, tx, true)
	}

	return error
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
}

func (chainService *ChainService) blockExists(blockHash *crypto.Hash) bool {
	return chainService.Index.HaveBlock(blockHash)
}

func (chainService *ChainService) GenerateBlock(leaderKey *secp256k1.PublicKey) (*chainTypes.Block, error) {
	chainService.DatabaseService.BeginTransaction()
	defer chainService.DatabaseService.Discard()

	height := chainService.BestChain.Height() + 1
	txs := chainService.transactionPool.GetPending(BlockGasLimit)

	finalTxs := make([]*chainTypes.Transaction, 0, len(txs))
	gasUsed := new(big.Int)
	for _, t := range txs {
		g, _, err := chainService.executeTransaction(t)
		if err == nil {
			finalTxs = append(finalTxs, t)
			gasUsed.Add(gasUsed, g)
		}else {
			dlog.Info("execute tx","err", err)
		}
	}

	timestamp := uint64(time.Now().Unix())
	previousHash := chainService.BestChain.Tip().Hash
	stateRoot := chainService.DatabaseService.GetStateRoot()
	merkleRoot := chainService.deriveMerkleRoot(finalTxs)

	block := &chainTypes.Block{
		Header: &chainTypes.BlockHeader{
			Version:      common.Version,
			PreviousHash: *previousHash,
			ChainId:      chainService.chainId,
			GasLimit:     *BlockGasLimit,
			GasUsed:      *gasUsed,
			Timestamp:    timestamp,
			StateRoot:    stateRoot,
			TxRoot:       merkleRoot,
			Height:       height,
			LeaderPubKey: *leaderKey,
		},
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

func (chainService *ChainService) GenesisBlock(genesisPubkey string) *chainTypes.Block {
	chainService.DatabaseService.BeginTransaction()
	defer chainService.DatabaseService.Discard()

	merkleRoot := chainService.deriveMerkleRoot(nil)
	b := common.Bytes(genesisPubkey)
	err := b.UnmarshalText(b)
	if err != nil {
		return nil
	}
	pubkey, err := secp256k1.ParsePubKey(b)
	if err != nil {
		return nil
	}
	return &chainTypes.Block{
		Header: &chainTypes.BlockHeader{
			Version:      common.Version,
			PreviousHash: crypto.Hash{},
			GasLimit:     *BlockGasLimit,
			GasUsed:      *new(big.Int),
			Timestamp:    1545282765,
			StateRoot:    nil,
			TxRoot:       merkleRoot,
			Height:       0,
			LeaderPubKey: *pubkey,
		},
		Data: &chainTypes.BlockData{
			TxCount: 0,
			TxList:  []*chainTypes.Transaction{},
		},
	}
}

// AccumulateRewards credits,The leader gets half of the reward and other ,Other participants get the average of the other half
func (chainService *ChainService) accumulateRewards(b *chainTypes.Block, totalGasBalance *big.Int) {
	reward := new(big.Int).SetUint64(uint64(Rewards))
	leaderAddr := crypto.PubKey2Address(&b.Header.LeaderPubKey)

	r := new(big.Int)
	r = r.Div(reward, new(big.Int).SetInt64(2))
	r.Add(r, totalGasBalance)
	chainService.DatabaseService.AddBalance(&leaderAddr, r, true)

	num := len(b.Header.MinorPubKeys)
	for _, memberPK := range b.Header.MinorPubKeys {
		if !memberPK.IsEqual(&b.Header.LeaderPubKey) {
			memberAddr := crypto.PubKey2Address(&memberPK)
			r.Div(reward, new(big.Int).SetInt64(int64(num*2)))
			chainService.DatabaseService.AddBalance(&memberAddr, r, true)
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
