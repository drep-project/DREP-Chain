package chain

import (
	"fmt"
	"github.com/drep-project/drep-chain/chain/store"
	"math/big"
	"path"
	"sync"

	"github.com/drep-project/drep-chain/app"
	"github.com/drep-project/drep-chain/params"
	"gopkg.in/urfave/cli.v1"

	"github.com/drep-project/binary"
	"github.com/drep-project/drep-chain/common"
	"github.com/drep-project/drep-chain/common/event"
	"github.com/drep-project/drep-chain/crypto"
	"github.com/drep-project/drep-chain/crypto/sha3"
	"github.com/drep-project/drep-chain/database"

	rpc2 "github.com/drep-project/drep-chain/pkgs/rpc"
	"github.com/drep-project/drep-chain/types"
)

var (
	RootChain          types.ChainIdType
	DefaultChainConfig = &ChainConfig{
		RemotePort:  55556,
		ChainId:     RootChain,
		GenesisAddr: params.HoleAddress,
	}
	span = uint64(params.MaxGasLimit / 360)
)

type ChainServiceInterface interface {
	app.Service
	ChainID() types.ChainIdType
	DeriveMerkleRoot(txs []*types.Transaction) []byte
	DeriveReceiptRoot(receipts []*types.Receipt) crypto.Hash
	GetBlockByHash(hash *crypto.Hash) (*types.Block, error)
	GetBlockByHeight(number uint64) (*types.Block, error)

	//DefaultChainConfig
	GetBlockHeaderByHash(hash *crypto.Hash) (*types.BlockHeader, error)
	GetBlockHeaderByHeight(number uint64) (*types.BlockHeader, error)

	GetHeader(hash crypto.Hash, number uint64) *types.BlockHeader
	GetCurrentHeader() *types.BlockHeader
	GetHighestBlock() (*types.Block, error)
	RootChain() types.ChainIdType
	BestChain() *ChainView
	CalcGasLimit(parent *types.BlockHeader, gasFloor, gasCeil uint64) *big.Int
	ProcessBlock(block *types.Block) (bool, bool, error)
	NewBlockFeed() *event.Feed
	GetLogsFeed() *event.Feed
	GetRMLogsFeed() *event.Feed
	BlockExists(blockHash *crypto.Hash) bool
	Index() *BlockIndex
	BlockValidator() []IBlockValidator
	AddBlockValidator(validator IBlockValidator)
	TransactionValidators() map[ITransactionSelector]ITransactionValidator
	AddTransactionValidator(selector ITransactionSelector, validator ITransactionValidator)
	AddGenesisProcess(validator IGenesisProcess)
	GetConfig() *ChainConfig
	DetachBlockFeed() *event.Feed
}

var cs ChainServiceInterface = &ChainService{}

//xxx
type ChainService struct {
	RpcService      *rpc2.RpcService          `service:"rpc"`
	DatabaseService *database.DatabaseService `service:"database"`
	apis            []app.API

	chainId types.ChainIdType

	lock         sync.RWMutex
	addBlockSync sync.Mutex

	// These fields are related to handling of orphan blocks.  They are
	// protected by a combination of the chain lock and the orphan lock.
	orphanLock   sync.RWMutex
	orphans      map[crypto.Hash]*types.OrphanBlock
	prevOrphans  map[crypto.Hash][]*types.OrphanBlock
	oldestOrphan *types.OrphanBlock

	blockIndex *BlockIndex
	bestChain  *ChainView

	Config       *ChainConfig
	genesisBlock *types.Block

	//提供新块订阅
	newBlockFeed    event.Feed
	detachBlockFeed event.Feed
	logsFeed        event.Feed
	rmLogsFeed      event.Feed

	blockValidator       []IBlockValidator
	transactionValidator map[ITransactionSelector]ITransactionValidator
	genesisProcess       []IGenesisProcess
	genesisConfig        string
	chainStore           *ChainStore
}

type ChainState struct {
	types.BestState
	db *store.StoreInterface
}

func NewChainService(config *ChainConfig, ds *database.DatabaseService) *ChainService {
	var err error
	chainService := &ChainService{}
	chainService.Config = config
	chainService.blockIndex = NewBlockIndex()
	chainService.bestChain = NewChainView(nil)
	chainService.orphans = make(map[crypto.Hash]*types.OrphanBlock)
	chainService.prevOrphans = make(map[crypto.Hash][]*types.OrphanBlock)
	chainService.chainStore = &ChainStore{ds.LevelDb()}

	chainService.transactionValidator = map[ITransactionSelector]ITransactionValidator{
		&TransferTxSelector{}: &TransferTransactionProcessor{},
		&AliasTxSelector{}:    &AliasTransactionProcessor{},
	}
	chainService.blockValidator = []IBlockValidator{NewChainBlockValidator(chainService)}
	chainService.genesisProcess = []IGenesisProcess{NewPreminerGenesisProcessor()}

	chainService.genesisBlock, err = chainService.GetGenisiBlock(chainService.Config.GenesisAddr)
	if err != nil {
		return nil
	}
	hash := chainService.genesisBlock.Header.Hash()
	if !chainService.chainStore.HasBlock(hash) {
		chainService.genesisBlock, err = chainService.ProcessGenesisBlock(chainService.Config.GenesisAddr)
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
		{
			Namespace: MODULENAME,
			Version:   "1.0",
			Service:   NewChainApi(chainService.DatabaseService.LevelDb(), chainService.BestChain(), chainService.chainStore),
			Public:    true,
		},
	}
	return chainService
}

func (chainService *ChainService) Init(executeContext *app.ExecuteContext) error {
	chainService.blockIndex = NewBlockIndex()
	chainService.bestChain = NewChainView(nil)
	chainService.chainStore = &ChainStore{chainService.DatabaseService.LevelDb()}
	chainService.orphans = make(map[crypto.Hash]*types.OrphanBlock)
	chainService.prevOrphans = make(map[crypto.Hash][]*types.OrphanBlock)

	chainService.blockValidator = []IBlockValidator{NewChainBlockValidator(chainService)}
	chainService.genesisProcess = []IGenesisProcess{NewPreminerGenesisProcessor()}
	chainService.transactionValidator = map[ITransactionSelector]ITransactionValidator{
		&TransferTxSelector{}:        &TransferTransactionProcessor{},
		&AliasTxSelector{}:           &AliasTransactionProcessor{},
		&StakeTxSelector{}:           &StakeTransactionProcessor{},
		&CancelVoteTxSelector{}:      &CancelVoteTransactionProcessor{},
		&CandidateTxSelector{}:       &CandidateTransactionProcessor{},
		&CancelCandidateTxSelector{}: &CancelCandidateTransactionProcessor{},
	}

	var err error
	chainService.genesisConfig = path.Join(executeContext.CommonConfig.HomeDir, "genesis.json")
	chainService.genesisBlock, err = chainService.GetGenisiBlock(chainService.Config.GenesisAddr)
	if err != nil {
		return err
	}
	hash := chainService.genesisBlock.Header.Hash()
	if !chainService.chainStore.HasBlock(hash) {
		chainService.genesisBlock, err = chainService.ProcessGenesisBlock(chainService.Config.GenesisAddr)
		err = chainService.createChainState()
		if err != nil {
			log.Error("createChainState err", err)
			return err
		}
		_, err = store.TrieStoreFromStore(chainService.DatabaseService.LevelDb(), chainService.genesisBlock.Header.StateRoot)
		if err != nil {
			return err
		}
	}
	err = chainService.InitStates()
	if err != nil {
		log.Error("InitStates err:", err)
		return err
	}
	chainService.apis = []app.API{
		{
			Namespace: MODULENAME,
			Version:   "1.0",
			Service:   NewChainApi(chainService.DatabaseService.LevelDb(), chainService.BestChain(), chainService.chainStore),
			Public:    true,
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

func (chainService *ChainService) RootChain() types.ChainIdType {
	return RootChain
}

func (chainService *ChainService) GetCurrentHeader() *types.BlockHeader {
	heighestBlockBode := chainService.bestChain.Tip()
	if heighestBlockBode == nil {
		return nil
	}
	block, err := chainService.chainStore.GetBlock(heighestBlockBode.Hash)
	if err != nil {
		return nil
	}
	return block.Header
}

func (chainService *ChainService) GetHighestBlock() (*types.Block, error) {
	heighestBlockBode := chainService.bestChain.Tip()
	if heighestBlockBode == nil {
		return nil, fmt.Errorf("chain not init")
	}
	block, err := chainService.chainStore.GetBlock(heighestBlockBode.Hash)
	if err != nil {
		return nil, err
	}
	return block, nil
}

func (chainService *ChainService) GetBlockByHash(hash *crypto.Hash) (*types.Block, error) {
	block, err := chainService.chainStore.GetBlock(hash)
	if err != nil {
		return nil, err
	}
	return block, nil
}

func (chainService *ChainService) GetBlockHeaderByHash(hash *crypto.Hash) (*types.BlockHeader, error) {
	blockNode, ok := chainService.blockIndex.Index[*hash]
	if !ok {
		return nil, ErrBlockNotFound
	}
	blockHeader := blockNode.Header()
	return &blockHeader, nil
}

func (chainService *ChainService) GetHeader(hash crypto.Hash, number uint64) *types.BlockHeader {
	header, _ := chainService.GetBlockHeaderByHash(&hash)
	return header
}

func (chainService *ChainService) GetBlockByHeight(number uint64) (*types.Block, error) {
	blockNode := chainService.bestChain.NodeByHeight(number)
	return chainService.GetBlockByHash(blockNode.Hash)
}

func (chainService *ChainService) GetBlockHeaderByHeight(number uint64) (*types.BlockHeader, error) {
	blockNode := chainService.bestChain.NodeByHeight(number)
	if blockNode == nil {
		return nil, ErrBlockNotFound
	}
	header := blockNode.Header()
	return &header, nil
}

func (chainService *ChainService) getTxHashes(ts []*types.Transaction) ([][]byte, error) {
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

func (cs *ChainService) DeriveMerkleRoot(txs []*types.Transaction) []byte {
	if len(txs) == 0 {
		return []byte{}
	}
	ts, _ := cs.getTxHashes(txs)
	merkle := common.NewMerkle(ts)
	return merkle.Root.Hash
}

func (cs *ChainService) DeriveReceiptRoot(receipts []*types.Receipt) crypto.Hash {
	if len(receipts) == 0 {
		return crypto.Hash{}
	}
	receiptsHashes := make([][]byte, len(receipts))
	for i, receipt := range receipts {
		b, _ := binary.Marshal(receipt)
		receiptsHashes[i] = sha3.Keccak256(b)
	}
	merkle := common.NewMerkle(receiptsHashes)
	receiptRoot := crypto.Hash{}
	receiptRoot.SetBytes(merkle.Root.Hash)
	return receiptRoot
}

func (chainService *ChainService) createChainState() error {
	node := types.NewBlockNode(chainService.genesisBlock.Header, nil)
	node.Status = types.StatusDataStored | types.StatusValid
	chainService.bestChain.SetTip(node)

	// Add the new node to the index which is used for faster lookups.
	chainService.blockIndex.AddNode(node)

	// Save the genesis block to the block index database.
	err := chainService.chainStore.PutBlockNode(node)
	if err != nil {
		return err
	}

	err = chainService.chainStore.PutBlock(chainService.genesisBlock)
	if err != nil {
		return err
	} else {
		return nil
	}
}

//GETTER

func (chainService *ChainService) GetDatabaseService() *database.DatabaseService {
	return chainService.DatabaseService
}

func (chainService *ChainService) DetachBlockFeed() *event.Feed {
	return &chainService.detachBlockFeed

}

func (chainService *ChainService) GetConfig() *ChainConfig {
	return chainService.Config
}

func (chainService *ChainService) BlockValidator() []IBlockValidator {
	return chainService.blockValidator
}

func (chainService *ChainService) TransactionValidators() map[ITransactionSelector]ITransactionValidator {
	return chainService.transactionValidator
}

func (chainService *ChainService) AddBlockValidator(validator IBlockValidator) {
	chainService.blockValidator = append(chainService.blockValidator, validator)
}

func (chainService *ChainService) AddTransactionValidator(selector ITransactionSelector, validator ITransactionValidator) {
	chainService.transactionValidator[selector] = validator
}

func (chainService *ChainService) AddGenesisProcess(validator IGenesisProcess) {
	chainService.genesisProcess = append(chainService.genesisProcess, validator)
}

func (chainService *ChainService) Index() *BlockIndex {
	return chainService.blockIndex
}

func (chainService *ChainService) NewBlockFeed() *event.Feed {
	return &chainService.newBlockFeed
}

func (chainService *ChainService) GetLogsFeed() *event.Feed {
	return &chainService.logsFeed
}

func (chainService *ChainService) GetRMLogsFeed() *event.Feed {
	return &chainService.rmLogsFeed
}

func (chainService *ChainService) BestChain() *ChainView {
	return chainService.bestChain
}

func (chainService *ChainService) ChainID() types.ChainIdType {
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

// Config
func (chainService *ChainService) DefaultConfig() *ChainConfig {
	return DefaultChainConfig
}
