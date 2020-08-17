package bft

import (
	"fmt"
	"github.com/drep-project/DREP-Chain/app"
	blockMgrService "github.com/drep-project/DREP-Chain/blockmgr"
	chainService "github.com/drep-project/DREP-Chain/chain"
	"github.com/drep-project/DREP-Chain/chain/store"
	"github.com/drep-project/DREP-Chain/common/event"
	"github.com/drep-project/DREP-Chain/crypto"
	"github.com/drep-project/DREP-Chain/crypto/secp256k1"
	"github.com/drep-project/DREP-Chain/database"
	"github.com/drep-project/DREP-Chain/network/p2p"
	p2pService "github.com/drep-project/DREP-Chain/network/service"
	"github.com/drep-project/DREP-Chain/params"
	accountService "github.com/drep-project/DREP-Chain/pkgs/accounts/service"
	consensusTypes "github.com/drep-project/DREP-Chain/pkgs/consensus/types"
	chainTypes "github.com/drep-project/DREP-Chain/types"
	"gopkg.in/urfave/cli.v1"
	"io/ioutil"
	"time"
)

var (
	MinerFlag = cli.BoolFlag{
		Name:  "miner",
		Usage: "is miner",
	}

	DefaultConfigMainnet = BftConfig{
		MyPk:           nil,
		StartMiner:     true,
		ProducerNum:    params.GenesisProducerNumMainnet,
		BlockInterval:  params.BlockInterval,
		ChangeInterval: params.ChangeInterval,
	}

	DefaultConfigTestnet = BftConfig{
		MyPk:           nil,
		StartMiner:     true,
		ProducerNum:    params.GenesisProducerNumTestnet,
		BlockInterval:  params.BlockInterval,
		ChangeInterval: params.ChangeInterval,
	}
)

type BftConsensusService struct {
	P2pServer        p2pService.P2P                       `service:"p2p"`
	ChainService     chainService.ChainServiceInterface   `service:"chain"`
	BroadCastor      blockMgrService.ISendMessage         `service:"blockmgr"`
	BlockMgrNotifier blockMgrService.IBlockNotify         `service:"blockmgr"`
	BlockGenerator   blockMgrService.IBlockBlockGenerator `service:"blockmgr"`
	DatabaseService  *database.DatabaseService            `service:"database"`
	WalletService    *accountService.AccountService       `service:"accounts"`

	BftConsensus *BftConsensus
	NetType      params.NetType

	apis   []app.API
	Config *BftConfig

	syncBlockEventSub  event.Subscription
	syncBlockEventChan chan event.SyncBlockEvent
	ConsensusEngine    consensusTypes.IConsensusEngine
	Miner              *secp256k1.PrivateKey
	//During the process of synchronizing blocks, the miner stopped mining
	pauseForSync bool
	start        bool
	quit         chan struct{}
}

func (bftConsensusService *BftConsensusService) Name() string {
	return "bft"
}

func (bftConsensusService *BftConsensusService) Api() []app.API {
	return bftConsensusService.apis
}

func (bftConsensusService *BftConsensusService) CommandFlags() ([]cli.Command, []cli.Flag) {
	return nil, []cli.Flag{MinerFlag}
}

func (bftConsensusService *BftConsensusService) Init(executeContext *app.ExecuteContext) error {
	if executeContext.Cli.GlobalIsSet(MinerFlag.Name) {
		bftConsensusService.Config.StartMiner = executeContext.Cli.GlobalBool(MinerFlag.Name)
	}

	var addPeerFeed event.Feed
	var removePeerFeed event.Feed
	bftConsensusService.BftConsensus = NewBftConsensus(
		bftConsensusService.ChainService,
		bftConsensusService.BlockGenerator,
		bftConsensusService.DatabaseService,
		bftConsensusService.P2pServer,
		bftConsensusService.Config,
		&addPeerFeed,
		&removePeerFeed,
	)

	bftConsensusService.ChainService.AddBlockValidator(&BlockMultiSigValidator{bftConsensusService.BftConsensus.GetProducers, bftConsensusService.ChainService.GetBlockByHash, bftConsensusService.Config.ProducerNum})
	bftConsensusService.ChainService.AddGenesisProcess(NewMinerGenesisProcessor())

	if bftConsensusService.WalletService.Wallet == nil {
		return ErrWalletNotOpen
	}

	bftConsensusService.P2pServer.AddProtocols([]p2p.Protocol{
		p2p.Protocol{
			Name:   "bftConsensusService",
			Length: NumberOfMsg,
			Run: func(peer *p2p.Peer, rw p2p.MsgReadWriter) error {
				log.WithField("newpeer ip", peer.IP()).Info("consensuse protocol")
				pi := consensusTypes.NewPeerInfo(peer, rw)

				addPeerFeed.Send(pi)
				defer func() {
					select {
					case <-bftConsensusService.quit:
						log.Info("consensuse protocol ,remove peer, ip", peer.IP())
					default:
						log.WithField("protocol out, remove peer ip", peer.IP()).Info("consensuse protocol")
						removePeerFeed.Send(pi)
						log.WithField("protocol out , remove peer ip", peer.IP()).Info("consensuse protocol,remove ok")
					}
				}()
				for {
					select {
					case <-bftConsensusService.quit:
						return fmt.Errorf("bft consensus service been stop")
					default:

						msg, err := rw.ReadMsg()
						if err != nil {
							log.WithField("Reason", err).WithField("Ip", pi.IP()).Error("consensus receive msg")
							return err
						}
						buf, err := ioutil.ReadAll(msg.Payload)
						if err != nil {
							return err
						}

						bftConsensusService.BftConsensus.ReceiveMsg(pi, msg.Code, buf)
					}
				}
			},
		},
	})
	bftConsensusService.syncBlockEventChan = make(chan event.SyncBlockEvent)
	bftConsensusService.syncBlockEventSub = bftConsensusService.BlockMgrNotifier.SubscribeSyncBlockEvent(bftConsensusService.syncBlockEventChan)
	bftConsensusService.quit = make(chan struct{})
	bftConsensusService.apis = []app.API{
		app.API{
			Namespace: "consensus",
			Version:   "1.0",
			Service: &ConsensusApi{
				consensusService: bftConsensusService,
			},
			Public: true,
		},
	}

	go bftConsensusService.handlerEvent()
	return nil
}

func (bftConsensusService *BftConsensusService) handlerEvent() {
	for {
		select {
		case e := <-bftConsensusService.syncBlockEventChan:
			if e.EventType == event.StartSyncBlock {
				bftConsensusService.pauseForSync = true
				//log.Trace("Start Sync Blcok")
			} else {
				bftConsensusService.pauseForSync = false
				//log.Trace("Stop Sync Blcok")
			}
		case <-bftConsensusService.quit:
			return
		}
	}
}

func (bftConsensusService *BftConsensusService) Start(executeContext *app.ExecuteContext) error {
	bftConsensusService.start = true

	go bftConsensusService.BftConsensus.processPeers()
	go bftConsensusService.BftConsensus.prepareForMining(bftConsensusService.P2pServer)
	go bftConsensusService.BftConsensus.bestHeight()

	go func() {
		for {
			select {
			case <-bftConsensusService.quit:
				return
			default:
				//consult privkey in wallet
				if bftConsensusService.Miner == nil {
					if bftConsensusService.Config.MyPk == nil {
						time.Sleep(time.Second * time.Duration(bftConsensusService.Config.BlockInterval))
						log.Trace("not set pubkey ,the node is listener")
						continue
					}

					accountNode, err := bftConsensusService.WalletService.Wallet.GetAccountByPubkey(bftConsensusService.Config.MyPk)
					if err != nil {
						log.WithField("err", err).WithField("addr", crypto.PubkeyToAddress(bftConsensusService.Config.MyPk).String()).Warn("privkey of MyPk in Config is not in local wallet or unlock address")
						time.Sleep(time.Second * time.Duration(bftConsensusService.Config.BlockInterval))
						continue
					}
					bftConsensusService.Miner = accountNode.PrivateKey
				}

				if bftConsensusService.pauseForSync {
					time.Sleep(time.Millisecond * 500)
					continue
				}
				log.WithField("Height", bftConsensusService.ChainService.BestChain().Height()).Trace("node start")
				block, err := bftConsensusService.BftConsensus.Run(bftConsensusService.Miner)
				if err != nil {
					log.WithField("Reason", err.Error()).Debug("Producer Block Fail")
				} else {
					_, _, err := bftConsensusService.ChainService.ProcessBlock(block)
					if err == nil {
						bftConsensusService.BroadCastor.BroadcastBlock(chainTypes.MsgTypeBlock, block, true)
						log.WithField("Height", block.Header.Height).WithField("txs:", block.Data.TxCount).Info("Process block successfully and broad case block message")
					} else {
						log.WithField("Height", block.Header.Height).WithField("txs:", block.Data.TxCount).WithField("err", err).Info("Process Block fail")
					}
				}
				nextBlockTime, waitSpan := bftConsensusService.getWaitTime()
				log.WithField("nextBlockTime", nextBlockTime).WithField("waitSpan", waitSpan).Debug("Sleep")
				time.Sleep(waitSpan)
			}
		}
	}()

	return nil
}

func (bftConsensusService *BftConsensusService) Stop(executeContext *app.ExecuteContext) error {
	if bftConsensusService.Config == nil { //|| !bftConsensusService.Config.StartMiner
		return nil
	}

	if bftConsensusService.quit != nil {
		close(bftConsensusService.quit)
	}

	if bftConsensusService.syncBlockEventSub != nil {
		bftConsensusService.syncBlockEventSub.Unsubscribe()
	}

	bftConsensusService.BftConsensus.Close()

	return nil
}

func (bftConsensusService *BftConsensusService) getWaitTime() (time.Time, time.Duration) {
	lastBlockTime := time.Unix(int64(bftConsensusService.ChainService.BestChain().Tip().TimeStamp), 0)
	targetTime := lastBlockTime.Add(time.Duration(int64(time.Second) * int64(bftConsensusService.Config.BlockInterval)))
	now := time.Now()
	if targetTime.Before(now) {
		interval := now.Sub(lastBlockTime)
		nextBlockInterval := int64(interval/(time.Second*time.Duration(bftConsensusService.Config.BlockInterval))) + 1
		nextBlockTime := lastBlockTime.Add(time.Second * time.Duration(nextBlockInterval*int64(bftConsensusService.Config.BlockInterval)))

		if nextBlockTime.Before(now) {
			return nextBlockTime, 0
		}

		return nextBlockTime, nextBlockTime.Sub(now)
	} else {
		return targetTime, targetTime.Sub(now)
	}
}

func (bftConsensusService *BftConsensusService) GetProducers(height uint64, topN int) ([]chainTypes.Producer, error) {
	block, err := bftConsensusService.ChainService.GetBlockByHeight(height)
	if err != nil {
		return nil, err
	}
	trie, err := store.TrieStoreFromStore(bftConsensusService.DatabaseService.LevelDb(), block.Header.StateRoot)
	if err != nil {
		return nil, err
	}
	return GetCandidates(trie, topN), nil
}

func (bftConsensusService *BftConsensusService) DefaultConfig(netType params.NetType) *BftConfig {
	switch bftConsensusService.NetType {
	case params.MainnetType:
		return &DefaultConfigMainnet
	case params.TestnetType:
		return &DefaultConfigTestnet
	default:
		return nil
	}

}
