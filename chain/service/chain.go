package service

import (
	"encoding/json"
	"fmt"
	"github.com/drep-project/drep-chain/pkgs/evm"
	"github.com/drep-project/drep-chain/transaction/txpool"
	"math/big"
	"strconv"
	"sync"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/drep-project/dlog"
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
	p2pTypes "github.com/drep-project/drep-chain/network/types"
	rpc2 "github.com/drep-project/drep-chain/pkgs/rpc"
	txType "github.com/drep-project/drep-chain/transaction/types"
)

var (
	rootChain     app.ChainIdType
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

	prvKey        *secp256k1.PrivateKey
	CurrentHeight int64
	peerStateMap  map[string]*chainTypes.PeerState

	Config *chainTypes.ChainConfig
	pid    *actor.PID

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
		chainTypes.MsgTypeTransaction:  txType.Transaction{},
		chainTypes.MsgTypePeerState:    chainTypes.PeerState{},
		chainTypes.MsgTypeReqPeerState: chainTypes.ReqPeerState{},
	}
}

func (chainService *ChainService) Init(executeContext *app.ExecuteContext) error {
	chainService.Config = &chainTypes.ChainConfig{}
	err := executeContext.UnmashalConfig(chainService.Name(), chainService.Config)
	if err != nil {
		return err
	}
	chainService.peerStateMap = make(map[string]*chainTypes.PeerState)
	chainService.CurrentHeight = chainService.DatabaseService.GetMaxHeight()
	if chainService.CurrentHeight == -1 {
		//generate genisis block
		genesisBlock := chainService.GenesisBlock(chainService.Config.GenesisPK)
		if genesisBlock == nil {
			return fmt.Errorf("genesis block err")
		}
		chainService.ProcessBlock(genesisBlock)
	}
	chainService.transactionPool = txpool.NewTransactionPool(chainService.DatabaseService)
	chainService.transactionPool.Start(&chainService.newBlockFeed)
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
	go chainService.fetchBlocks()
	return nil
}

func (chainService *ChainService) Stop(executeContext *app.ExecuteContext) error {
	return nil
}

func (chainService *ChainService) SendTransaction(tx *txType.Transaction) error {
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

func (chainService *ChainService) ProcessBlock(block *chainTypes.Block) (*big.Int, error) {
	chainService.addBlockSync.Lock()
	defer chainService.addBlockSync.Unlock()
	dlog.Trace("Process block leader.", "LeaderPubKey", crypto.PubKey2Address(block.Header.LeaderPubKey).Hex(), " height ", strconv.FormatInt(block.Header.Height, 10))
	gasUsed, err := chainService.ExecuteTransactions(block)
	if err == nil {
		chainService.CurrentHeight = block.Header.Height
	}

	addrMap := make(map[crypto.CommonAddress]struct{})
	var addrs []*crypto.CommonAddress
	for _,tx := range block.Data.TxList {
		addr := tx.From()
		if _,ok:=addrMap[*addr]; !ok{
			addrMap[*addr] = struct{}{}
			addrs = append(addrs, addr)
		}
	}

	if len(addrs) > 0 {
		chainService.newBlockFeed.Send(addrs)
	}

	return gasUsed, err
}

func (chainService *ChainService) ProcessBlockReq(peer *p2pTypes.Peer, req *chainTypes.BlockReq) {
	from := req.Height + 1
	size := int64(200)
	for i := from; i <= chainService.DatabaseService.GetMaxHeight(); {
		bs := chainService.DatabaseService.GetBlocksFrom(i, size)
		resp := &chainTypes.BlockResp{Height: chainService.DatabaseService.GetMaxHeight(), Blocks: bs}
		chainService.P2pServer.Send(peer, resp)
		i += int64(len(bs))
	}
}

func (chainService *ChainService) GenerateBlock(leaderKey *secp256k1.PublicKey, members []*secp256k1.PublicKey) (*chainTypes.Block, error) {
	chainService.DatabaseService.BeginTransaction()
	defer chainService.DatabaseService.Discard()

	height := chainService.DatabaseService.GetMaxHeight() + 1
	txs := chainService.transactionPool.GetPending(BlockGasLimit)

	gasUsed := new(big.Int)
	for _, t := range txs {
		g, _ := chainService.execute(t)
		gasUsed.Add(gasUsed, g)
	}

	timestamp := time.Now().Unix()
	previousHash := chainService.DatabaseService.GetPreviousBlockHash()

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
			TxHashes:     txHashes,
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

func (chainService *ChainService) GetTxHashes(ts []*txType.Transaction) ([][]byte, error) {
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
			PreviousHash: []byte{},
			GasLimit:     BlockGasLimit,
			GasUsed:      new(big.Int),
			Timestamp:    1545282765,
			StateRoot:    stateRoot,
			MerkleRoot:   merkleRoot,
			TxHashes:     [][]byte{},
			Height:       0,
			LeaderPubKey: pubkey,
			MinorPubKeys: memberPks,
		},
		Data: &chainTypes.BlockData{
			TxCount: 0,
			TxList:  []*txType.Transaction{},
		},
	}
}

// AccumulateRewards credits,The leader gets half of the reward and other ,Other participants get the average of the other half
func (chainService *ChainService) accumulateRewards(b *chainTypes.Block, chainId app.ChainIdType) {
	chainService.DatabaseService.BeginTransaction()
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

	chainService.DatabaseService.Commit()
}

func (chainService *ChainService) Subscribe(subchan chan event.SyncBlockEvent) event.Subscription {
	return chainService.syncBlockEvent.Subscribe(subchan)
}

func (chainService *ChainService)GetTransactionCount(addr * crypto.CommonAddress)int64{
	return chainService.transactionPool.GetTransactionCount(addr)
}