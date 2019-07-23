package chainservice

import (
	"fmt"
	"github.com/drep-project/drep-chain/app"
	"github.com/drep-project/drep-chain/chain/params"
	"gopkg.in/urfave/cli.v1"
	"math/big"
	"sync"

	"github.com/drep-project/binary"
	"github.com/drep-project/drep-chain/common"
	"github.com/drep-project/drep-chain/common/event"
	"github.com/drep-project/drep-chain/crypto"
	"github.com/drep-project/drep-chain/crypto/sha3"
	"github.com/drep-project/drep-chain/database"
	"github.com/drep-project/drep-chain/pkgs/evm"

	chainTypes "github.com/drep-project/drep-chain/chain/types"
	rpc2 "github.com/drep-project/drep-chain/pkgs/rpc"
)

var (
	rootChain          app.ChainIdType
	DefaultChainConfig = &chainTypes.ChainConfig{
		RemotePort:       55556,
		ChainId:          app.ChainIdType{},
		GenesisPK:        "0x0373654ccdb250f2cfcfe64c783a44b9ea85bc47f2f00c480d05082428d277d6d0",
		SkipCheckMutiSig: true,
	}
	span = uint64(params.MaxGasLimit / 360)
)

type ChainServiceInterface interface {
	app.Service
	ChainID() app.ChainIdType
	DeriveMerkleRoot(txs []*chainTypes.Transaction) []byte
	GetBlockByHash(hash *crypto.Hash) (*chainTypes.Block, error)
	GetBlockByHeight(number uint64) (*chainTypes.Block, error)

	//DefaultChainConfig
	GetBlockHeaderByHash(hash *crypto.Hash) (*chainTypes.BlockHeader, error)
	GetBlockHeaderByHeight(number uint64) (*chainTypes.BlockHeader, error)
	GetBlocksFrom(start, size uint64) ([]*chainTypes.Block, error)

	GetCurrentState() *database.Database
	GetHeader(hash crypto.Hash, number uint64) *chainTypes.BlockHeader
	GetHighestBlock() (*chainTypes.Block, error)
	RootChain() app.ChainIdType
	BestChain() *ChainView
	CalcGasLimit(parent *chainTypes.BlockHeader, gasFloor, gasCeil uint64) *big.Int
	ProcessBlock(block *chainTypes.Block) (bool, bool, error)
	NewBlockFeed() *event.Feed
	BlockExists(blockHash *crypto.Hash) bool
	TransactionValidator() ITransactionValidator
	GetDatabaseService() *database.DatabaseService
	Index() *BlockIndex
	BlockValidator() IBlockValidator
	Config() *chainTypes.ChainConfig
	AccumulateRewards(db *database.Database, b *chainTypes.Block, totalGasBalance *big.Int) error
	DetachBlockFeed() *event.Feed
}

var cs ChainServiceInterface = &ChainService{}

//xxx
type ChainService struct {
	RpcService      *rpc2.RpcService          `service:"rpc"`
	DatabaseService *database.DatabaseService `service:"database"`
	VmService       evm.Vm                    `service:"vm"`
	apis            []app.API

	stateProcessor *StateProcessor

	chainId app.ChainIdType

	lock         sync.RWMutex
	addBlockSync sync.Mutex

	// These fields are related to handling of orphan blocks.  They are
	// protected by a combination of the chain lock and the orphan lock.
	orphanLock   sync.RWMutex
	orphans      map[crypto.Hash]*chainTypes.OrphanBlock
	prevOrphans  map[crypto.Hash][]*chainTypes.OrphanBlock
	oldestOrphan *chainTypes.OrphanBlock

	blockIndex    *BlockIndex
	bestChain     *ChainView
	stateLock     sync.RWMutex
	StateSnapshot *ChainState

	config       *chainTypes.ChainConfig
	genesisBlock *chainTypes.Block

	//提供新块订阅
	newBlockFeed    event.Feed
	detachBlockFeed event.Feed

	blockValidator       IBlockValidator
	transactionValidator ITransactionValidator

	//blockDb *database.Database
}

type ChainState struct {
	chainTypes.BestState
	db *database.Database
}

func (chainService *ChainService) GetDatabaseService() *database.DatabaseService {
	return chainService.DatabaseService
}

func (chainService *ChainService) DetachBlockFeed() *event.Feed {
	return &chainService.detachBlockFeed

}

func (chainService *ChainService) Config() *chainTypes.ChainConfig {
	return chainService.config
}

func (chainService *ChainService) BlockValidator() IBlockValidator {
	return chainService.blockValidator
}

func (chainService *ChainService) Index() *BlockIndex {
	return chainService.blockIndex
}

func (chainService *ChainService) TransactionValidator() ITransactionValidator {
	return chainService.transactionValidator
}

func (chainService *ChainService) NewBlockFeed() *event.Feed {
	return &chainService.newBlockFeed
}

func (chainService *ChainService) BestChain() *ChainView {
	return chainService.bestChain
}

func (chainService *ChainService) ChainID() app.ChainIdType {
	return chainService.chainId
}

func (chainService *ChainService) Name() string {
	return MODULENAME
}

func (chainService *ChainService) Api() []app.API {
	return chainService.apis
}

func (chainService *ChainService) CommandFlags() ([]cli.Command, []cli.Flag) {
	return nil, []cli.Flag{}
}

func NewChainService(config *chainTypes.ChainConfig, ds *database.DatabaseService) *ChainService {
	chainService := &ChainService{}
	chainService.config = config
	var err error
	chainService.blockIndex = NewBlockIndex()
	chainService.bestChain = NewChainView(nil)
	chainService.orphans = make(map[crypto.Hash]*chainTypes.OrphanBlock)
	chainService.prevOrphans = make(map[crypto.Hash][]*chainTypes.OrphanBlock)
	chainService.stateProcessor = NewStateProcessor(chainService)
	chainService.transactionValidator = NewTransactionValidator(chainService)
	chainService.blockValidator = NewChainBlockValidator(chainService)
	chainService.DatabaseService = ds
	//chainService.blockDb = chainService.DatabaseService.BeginTransaction()
	if chainService.config.GenesisPK == "" {
		return nil
	}
	if len(chainService.config.Producers) == 0 {
		return nil
	}
	chainService.genesisBlock = chainService.GetGenisiBlock(chainService.config.GenesisPK)
	hash := chainService.genesisBlock.Header.Hash()
	if !chainService.DatabaseService.HasBlock(hash) {
		chainService.genesisBlock, err = chainService.ProcessGenesisBlock(chainService.config.GenesisPK)
		err = chainService.createChainState()
		if err != nil {
			return nil
		}
	}

	err = chainService.InitStates()
	if err != nil {
		return nil
	}

	chainService.apis = []app.API{
		app.API{
			Namespace: MODULENAME,
			Version:   "1.0",
			Service: &ChainApi{
				chainService: chainService,
				dbService:    chainService.DatabaseService,
			},
			Public: true,
		},
	}
	return chainService
}

func (chainService *ChainService) Init(executeContext *app.ExecuteContext) error {
	chainService.config = DefaultChainConfig

	err := executeContext.UnmashalConfig(chainService.Name(), chainService.config)
	if err != nil {
		return err
	}
	chainService.blockIndex = NewBlockIndex()
	chainService.bestChain = NewChainView(nil)
	chainService.orphans = make(map[crypto.Hash]*chainTypes.OrphanBlock)
	chainService.prevOrphans = make(map[crypto.Hash][]*chainTypes.OrphanBlock)
	chainService.stateProcessor = NewStateProcessor(chainService)
	chainService.transactionValidator = NewTransactionValidator(chainService)
	chainService.blockValidator = NewChainBlockValidator(chainService)

	if chainService.config.GenesisPK == "" {
		return ErrGenesisPkNotFound
	}
	if len(chainService.config.Producers) == 0 {
		return ErrBlockProducerNotFound
	}
	chainService.genesisBlock = chainService.GetGenisiBlock(chainService.config.GenesisPK)
	hash := chainService.genesisBlock.Header.Hash()
	if !chainService.DatabaseService.HasBlock(hash) {
		chainService.genesisBlock, err = chainService.ProcessGenesisBlock(chainService.config.GenesisPK)
		err = chainService.createChainState()
		if err != nil {
			return err
		}
	}

	err = chainService.InitStates()
	if err != nil {
		return err
	}

	chainService.apis = []app.API{
		app.API{
			Namespace: MODULENAME,
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
	return nil
}

func (chainService *ChainService) Stop(executeContext *app.ExecuteContext) error {
	return nil
}

func (chainService *ChainService) BlockExists(blockHash *crypto.Hash) bool {
	return chainService.blockIndex.HaveBlock(blockHash)
}

func (chainService *ChainService) RootChain() app.ChainIdType {
	return rootChain
}

func (chainService *ChainService) GetBlocksFrom(start, size uint64) ([]*chainTypes.Block, error) {
	blocks := []*chainTypes.Block{}
	for i := start; i < start+size; i++ {
		node := chainService.bestChain.NodeByHeight(i)
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
	heighestBlockBode := chainService.bestChain.Tip()
	if heighestBlockBode == nil {
		return nil, fmt.Errorf("chain not init")
	}
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
	blockNode, ok := chainService.blockIndex.Index[*hash]
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
	blockNode := chainService.bestChain.NodeByHeight(number)
	return chainService.GetBlockByHash(blockNode.Hash)
}

func (chainService *ChainService) GetBlockHeaderByHeight(number uint64) (*chainTypes.BlockHeader, error) {
	blockNode := chainService.bestChain.NodeByHeight(number)
	if blockNode == nil {
		return nil, ErrBlockNotFound
	}
	header := blockNode.Header()
	return &header, nil
}

func (chainService *ChainService) getTxHashes(ts []*chainTypes.Transaction) ([][]byte, error) {
	txHashes := make([][]byte, len(ts))
	for i, tx := range ts {
		b, err := binary.Marshal(tx.Data)
		if err != nil {
			return nil, err
		}
		txHashes[i] = sha3.Keccak256(b)
	}
	return txHashes, nil
}

func (cs *ChainService) DeriveMerkleRoot(txs []*chainTypes.Transaction) []byte {
	if len(txs) == 0 {
		return []byte{}
	}
	ts, _ := cs.getTxHashes(txs)
	merkle := common.NewMerkle(ts)
	return merkle.Root.Hash
}

func (chainService *ChainService) createChainState() error {
	node := chainTypes.NewBlockNode(chainService.genesisBlock.Header, nil)
	node.Status = chainTypes.StatusDataStored | chainTypes.StatusValid
	chainService.bestChain.SetTip(node)

	// Add the new node to the index which is used for faster lookups.
	chainService.blockIndex.AddNode(node)

	// Initialize the state related to the best block.  Since it is the
	// genesis block, use its timestamp for the median time.
	chainService.stateLock.Lock()
	chainService.StateSnapshot = &ChainState{
		BestState: *chainTypes.NewBestState(node),
		db:        chainService.DatabaseService.BeginTransaction(true),
	}
	chainService.stateLock.Unlock()

	// Save the genesis block to the block index database.
	err := chainService.DatabaseService.PutBlockNode(node)
	if err != nil {
		return err
	}

	// Store the current best chain state into the database.
	chainService.stateLock.Lock()
	state := chainService.StateSnapshot.BestState
	chainService.stateLock.Unlock()
	err = chainService.DatabaseService.PutChainState(&state)
	if err != nil {
		return err
	}
	err = chainService.DatabaseService.PutBlock(chainService.genesisBlock)
	if err != nil {
		return err
	} else {
		return nil
	}
}

func (chainService *ChainService) GetCurrentState() *database.Database {
	chainService.stateLock.Lock()
	defer chainService.stateLock.Unlock()
	return chainService.StateSnapshot.db

}
