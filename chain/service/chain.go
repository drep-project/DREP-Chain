package service

import (
	"encoding/json"
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
		GenesisPK:  "0x03177b8e4ef31f4f801ce00260db1b04cc501287e828692a404fdbc46c7ad6ff26",
	}
	//genesisPubkey = "0x03177b8e4ef31f4f801ce00260db1b04cc501287e828692a404fdbc46c7ad6ff26"
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
	syncingMaxHeight int64
	syncMaxHeightMut sync.Mutex

	//提供新块订阅
	newBlockFeed event.Feed
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

	chainService.genesisBlock = chainService.GenesisBlock(chainService.Config.GenesisPK)
	block, err := chainService.DatabaseService.GetBlock(chainService.genesisBlock.Header.Hash())
	if err != nil && err.Error() != "leveldb: not found" {
		return nil
	}
	if block == nil {
		//generate genisis block
		chainService.createChainState()
		chainService.ProcessBlock(chainService.genesisBlock)
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
	chainService.syncingMaxHeight = -1
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
	go chainService.fetchBlocks()
	return nil
}

func (chainService *ChainService) Stop(executeContext *app.ExecuteContext) error {
	return nil
}

func (chainService *ChainService) SendTransaction(tx *chainTypes.Transaction) error {
	if id, err := tx.TxId(); err == nil {
		ForwardTransaction(id)
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

func (chainService *ChainService) GenerateBlock(leaderKey *secp256k1.PublicKey, members []*secp256k1.PublicKey) (*chainTypes.Block, error) {
	chainService.DatabaseService.BeginTransaction()
	defer chainService.DatabaseService.Discard()

	height := chainService.BestChain.Height() + 1
	txs := chainService.transactionPool.GetPending(BlockGasLimit)

	gasUsed := new(big.Int)
	for _, t := range txs {
		g, _ := chainService.execute(t)
		gasUsed.Add(gasUsed, g)
	}

	timestamp := time.Now().Unix()
	previousHash := chainService.BestChain.Tip().Hash

	stateRoot := chainService.DatabaseService.GetStateRoot()
	txHashes, _ := chainService.GetTxHashes(txs)
	merkle := chainService.DatabaseService.NewMerkle(txHashes)
	merkleRoot := merkle.Root.Hash

	var memberPks []*secp256k1.PublicKey
	for _, p := range members {
		memberPks = append(memberPks, p)
	}

	block := &chainTypes.Block{
		Header: &chainTypes.BlockHeader{
			Version:      common.Version,
			PreviousHash: previousHash,
			ChainId:      chainService.chainId,
			GasLimit:     BlockGasLimit,
			GasUsed:      gasUsed,
			Timestamp:    timestamp,
			StateRoot:    stateRoot,
			MerkleRoot:   merkleRoot,
			Height:       height,
			LeaderPubKey: leaderKey,
			MinorPubKeys: memberPks,
		},
		Data: &chainTypes.BlockData{
			TxCount: int32(len(txs)),
			TxList:  txs,
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

//func (chainService *ChainService) GenerateBalanceTransaction(from *secp256k1.PublicKey, to crypto.CommonAddress, amount *big.Int) *txType.Transaction {
//	address := crypto.PubKey2Address(from)
//	nonce := chainService.DatabaseService.GetNonce(&address, false)
//	data := &txType.TransactionData{
//		Version:   Version,
//		Nonce:     nonce,
//		Type:      TransferType,
//		To:        to,
//		Amount:    amount,
//		GasPrice:  DefaultGasPrice,
//		GasLimit:  TransferGas,
//		Timestamp: time.Now().Unix(),
//		PubKey:    from,
//	}
//	return &txType.Transaction{Data: data}
//}
//
//func (chainService *ChainService) GenerateCreateContractTransaction(from *secp256k1.PublicKey, to crypto.CommonAddress, byteCode []byte) *txType.Transaction {
//	address := crypto.PubKey2Address(from)
//	nonce := chainService.DatabaseService.GetNonce(&address, false)
//	nonce++
//	data := &txType.TransactionData{
//		Nonce:     nonce,
//		Type:      CreateContractType,
//		GasPrice:  DefaultGasPrice,
//		GasLimit:  CreateContractGas,
//		Timestamp: time.Now().Unix(),
//		Data:      make([]byte, len(byteCode)+1),
//		PubKey:    from,
//	}
//	copy(data.Data[1:], byteCode)
//	data.Data[0] = 2
//	return &txType.Transaction{Data: data}
//}
//
//func (chainService *ChainService) GenerateCallContractTransaction(from *secp256k1.PublicKey, to crypto.CommonAddress, input []byte, amount *big.Int, readOnly bool) *chainTypes.Transaction {
//	address := crypto.PubKey2Address(from)
//	nonce := chainService.DatabaseService.GetNonce(&address, false)
//	nonce++
//	data := &txType.TransactionData{
//		Nonce:     nonce,
//		Type:      CallContractType,
//		To:        to,
//		Amount:    amount,
//		GasPrice:  DefaultGasPrice,
//		GasLimit:  CallContractGas,
//		Timestamp: time.Now().Unix(),
//		PubKey:    from,
//		Data:      make([]byte, len(input)+1),
//	}
//	copy(data.Data[1:], input)
//	if readOnly {
//		data.Data[0] = 1
//	} else {
//		data.Data[0] = 0
//	}
//	return &txType.Transaction{Data: data}
//}

func (chainService *ChainService) GenesisBlock(genesisPubkey string) *chainTypes.Block {
	chainService.DatabaseService.BeginTransaction()
	defer chainService.DatabaseService.Discard()
	stateRoot := chainService.DatabaseService.GetStateRoot()
	merkle := chainService.DatabaseService.NewMerkle([][]byte{})
	merkleRoot := merkle.Root.Hash

	b := common.Bytes(genesisPubkey)
	err := b.UnmarshalText(b)
	if err != nil {
		return nil
	}
	pubkey, err := secp256k1.ParsePubKey(b)
	if err != nil {
		return nil
	}
	var memberPks []*secp256k1.PublicKey = nil
	return &chainTypes.Block{
		Header: &chainTypes.BlockHeader{
			Version:      common.Version,
			PreviousHash: &crypto.Hash{},
			GasLimit:     BlockGasLimit,
			GasUsed:      new(big.Int),
			Timestamp:    1545282765,
			StateRoot:    stateRoot,
			MerkleRoot:   merkleRoot,
			Height:       0,
			LeaderPubKey: pubkey,
			MinorPubKeys: memberPks,
		},
		Data: &chainTypes.BlockData{
			TxCount: 0,
			TxList:  []*chainTypes.Transaction{},
		},
	}
}

// AccumulateRewards credits,The leader gets half of the reward and other ,Other participants get the average of the other half
func (chainService *ChainService) accumulateRewards(b *chainTypes.Block, chainId app.ChainIdType) {
	//chainService.DatabaseService.BeginTransaction()
	reward := new(big.Int).SetUint64(uint64(Rewards))
	leaderAddr := crypto.PubKey2Address(b.Header.LeaderPubKey)

	r := new(big.Int)
	chainService.DatabaseService.AddBalance(&leaderAddr, r.Div(reward, new(big.Int).SetInt64(2)), true)

	num := len(b.Header.MinorPubKeys)
	for _, memberPK := range b.Header.MinorPubKeys {
		memberAddr := crypto.PubKey2Address(memberPK)
		r.Div(r, new(big.Int).SetInt64(int64(num)))
		chainService.DatabaseService.AddBalance(&memberAddr, r, true)
	}

	//chainService.DatabaseService.Commit()
}

func (chainService *ChainService) Subscribe(subchan chan event.SyncBlockEvent) event.Subscription {
	return chainService.syncBlockEvent.Subscribe(subchan)
}

func (chainService *ChainService) GetTransactionCount(addr *crypto.CommonAddress) int64 {
	return chainService.transactionPool.GetTransactionCount(addr)
}

func (chainService *ChainService) GetBlocksFrom(start, size int64) ( []*chainTypes.Block, error){
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
