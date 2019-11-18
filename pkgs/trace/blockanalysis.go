package trace

import (
	"github.com/drep-project/DREP-Chain/chain/store"
	"github.com/drep-project/DREP-Chain/common/event"
	"github.com/drep-project/DREP-Chain/crypto"
	"github.com/drep-project/DREP-Chain/database/dbinterface"
	"github.com/drep-project/DREP-Chain/pkgs/consensus/service"
	"github.com/drep-project/DREP-Chain/pkgs/consensus/service/bft"
	"github.com/drep-project/DREP-Chain/types"
)

type GetProducer func(root []byte) ([]crypto.CommonAddress, error)

type BlockAnalysis struct {
	Config           HistoryConfig
	getBlock         func(uint64) (*types.Block, error)
	eventNewBlockSub event.Subscription
	newBlockChan     chan *types.ChainEvent

	consensusService *service.ConsensusService
	trieStore        dbinterface.KeyValueStore

	detachBlockSub  event.Subscription
	detachBlockChan chan *types.Block
	store           IStore
	readyToQuit     chan struct{}
}

func NewBlockAnalysis(config HistoryConfig, consensusService *service.ConsensusService, trieStore dbinterface.KeyValueStore, getBlock func(uint64) (*types.Block, error)) *BlockAnalysis {
	blockAnalysis := &BlockAnalysis{}
	blockAnalysis.Config = config
	blockAnalysis.getBlock = getBlock
	blockAnalysis.trieStore = trieStore
	blockAnalysis.consensusService = consensusService
	blockAnalysis.newBlockChan = make(chan *types.ChainEvent, 1000)
	blockAnalysis.detachBlockChan = make(chan *types.Block, 1000)
	blockAnalysis.readyToQuit = make(chan struct{})
	return blockAnalysis
}

func (blockAnalysis *BlockAnalysis) Start(newBlock, detachBlock *event.Feed) error {
	blockAnalysis.eventNewBlockSub = newBlock.Subscribe(blockAnalysis.newBlockChan)
	blockAnalysis.detachBlockSub = detachBlock.Subscribe(blockAnalysis.detachBlockChan)
	var err error
	getProducer := func(root []byte) ([]crypto.CommonAddress, error) {
		if blockAnalysis.consensusService.Config.ConsensusMode == "solo" {
			pk := blockAnalysis.consensusService.SoloService.Config.MyPk
			return []crypto.CommonAddress{crypto.PubkeyToAddress(pk)}, nil
		} else {
			trie, err := store.TrieStoreFromStore(blockAnalysis.trieStore, root)
			if err != nil {
				return nil, err
			}
			op := bft.ConsensusOp{trie}
			producers, err := op.GetProducer()
			if err != nil {
				return nil, err
			}
			miners := make([]crypto.CommonAddress, len(producers))
			for index, p := range producers {
				miners[index] = crypto.PubkeyToAddress(p.Pubkey)
			}
			return miners, nil
		}

	}
	if blockAnalysis.Config.DbType == "leveldb" {
		blockAnalysis.store, err = NewLevelDbStore(blockAnalysis.Config.HistoryDir, getProducer, blockAnalysis.consensusService.Config.ConsensusMode)
		if err != nil {
			log.WithField("err", err).WithField("path", blockAnalysis.Config.HistoryDir).Error("cannot open db file")
		}
	} else if blockAnalysis.Config.DbType == "mongo" {
		blockAnalysis.store, err = NewMongoDbStore(blockAnalysis.Config.Url, getProducer, blockAnalysis.consensusService.Config.ConsensusMode, DefaultDbName)
		if err != nil {
			log.WithField("err", err).WithField("url", blockAnalysis.Config.Url).Error("try connect mongo fail")
		}
	} else {
		return ErrUnSupportDbType
	}
	if err != nil {
		return err
	}

	go blockAnalysis.process()
	return nil
}

// Process used to resolve two types of signals,
// newBlockChan is the signal that blocks are added to the chain,
// the other is the detachBlockChan that blocks are withdrawn from the chain.
func (blockAnalysis *BlockAnalysis) process() error {
	for {
		select {
		case block := <-blockAnalysis.newBlockChan:
			blockAnalysis.store.InsertRecord(block.Block)
		case block := <-blockAnalysis.detachBlockChan:
			blockAnalysis.store.DelRecord(block)
		default:
			select {
			case <-blockAnalysis.readyToQuit:
				<-blockAnalysis.readyToQuit
				goto STOP
			default:
			}
		}
	}
STOP:
	return nil
}

func (blockAnalysis *BlockAnalysis) Close() error {
	if blockAnalysis.eventNewBlockSub != nil {
		blockAnalysis.eventNewBlockSub.Unsubscribe()
	}
	if blockAnalysis.detachBlockSub != nil {
		blockAnalysis.detachBlockSub.Unsubscribe()
	}
	if blockAnalysis.readyToQuit != nil {
		blockAnalysis.readyToQuit <- struct{}{} // tell process to stop in deal all blocks in chanel
		blockAnalysis.readyToQuit <- struct{}{} // wait for process is ok to stop
		blockAnalysis.store.Close()
	}
	return nil
}

func (blockAnalysis *BlockAnalysis) Rebuild(from, end int) error {
	/*currentHeight := blockAnalysis.ChainService.BestChain().Height()
	if uint64(from) > currentHeight {
		return nil
	}
	*/
	for i := from; i < end; i++ {
		block, err := blockAnalysis.getBlock(uint64(i))
		if err != nil {
			return ErrBlockNotFound
		}
		exist, err := blockAnalysis.store.ExistRecord(block)
		if err != nil {
			return err
		}
		if exist {
			blockAnalysis.store.DelRecord(block)
		}
		blockAnalysis.store.InsertRecord(block)
	}
	return nil
}
