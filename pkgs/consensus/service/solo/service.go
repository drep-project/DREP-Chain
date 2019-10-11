package solo

import (
	"github.com/drep-project/drep-chain/crypto"
	"github.com/drep-project/drep-chain/crypto/secp256k1"
	"time"

	"github.com/drep-project/drep-chain/app"
	blockMgrService "github.com/drep-project/drep-chain/blockmgr"
	chainService "github.com/drep-project/drep-chain/chain"
	"github.com/drep-project/drep-chain/common/event"
	"github.com/drep-project/drep-chain/database"
	p2pService "github.com/drep-project/drep-chain/network/service"
	accountService "github.com/drep-project/drep-chain/pkgs/accounts/service"
	consensusTypes "github.com/drep-project/drep-chain/pkgs/consensus/types"
	chainTypes "github.com/drep-project/drep-chain/types"
	"gopkg.in/urfave/cli.v1"
)

var (
	EnableSoloConsensusFlag = cli.BoolFlag{
		Name:  "enable",
		Usage: "enable solo consensus",
	}
)

const (
	blockInterval = time.Second * 5
)

type ConsensusService struct {
	P2pServer        p2pService.P2P                       `service:"p2p"`
	ChainService     chainService.ChainServiceInterface   `service:"chain"`
	BroadCastor      blockMgrService.ISendMessage         `service:"blockmgr"`
	BlockMgrNotifier blockMgrService.IBlockNotify         `service:"blockmgr"`
	BlockGenerator   blockMgrService.IBlockBlockGenerator `service:"blockmgr"`
	DatabaseService  *database.DatabaseService            `service:"database"`
	WalletService    *accountService.AccountService       `service:"accounts"`

	Config           *SoloConfig
	syncBlockEventSub  event.Subscription
	syncBlockEventChan chan event.SyncBlockEvent
	ConsensusEngine    consensusTypes.IConsensusEngine
	Miner              *secp256k1.PrivateKey
	//During the process of synchronizing blocks, the miner stopped mining
	pauseForSync bool
	start        bool
	quit         chan struct{}
}

func (consensusService *ConsensusService) Name() string {
	return "solo"
}

func (consensusService *ConsensusService) Api() []app.API {
	return nil
}

func (consensusService *ConsensusService) CommandFlags() ([]cli.Command, []cli.Flag) {
	return nil, []cli.Flag{EnableSoloConsensusFlag}
}

func (consensusService *ConsensusService) Init(executeContext *app.ExecuteContext) error {
	if executeContext.Cli.GlobalIsSet(EnableSoloConsensusFlag.Name) {
		consensusService.Config.Miner = executeContext.Cli.GlobalBool(EnableSoloConsensusFlag.Name)
	}

	consensusService.ChainService.AddBlockValidator(NewSoloValidator(consensusService.Config.MyPk))
	if !consensusService.Config.Miner {
		return nil
	} else {
		if consensusService.WalletService.Wallet == nil {
			return ErrWalletNotOpen
		}
	}

	var engine consensusTypes.IConsensusEngine
	engine = NewSoloConsensus(
		consensusService.ChainService,
		consensusService.BlockGenerator,
		consensusService.Config.MyPk,
		consensusService.DatabaseService)
	//consult privkey in wallet
	accountNode, err := consensusService.WalletService.Wallet.GetAccountByPubkey(consensusService.Config.MyPk)
	if err != nil {
		log.WithField("init err", err).WithField("addr", crypto.PubkeyToAddress(consensusService.Config.MyPk).String()).Error("privkey of MyPk in Config is not in local wallet")
		return err
	}
	consensusService.Miner = accountNode.PrivateKey
	consensusService.ConsensusEngine = engine
	consensusService.syncBlockEventChan = make(chan event.SyncBlockEvent)
	consensusService.syncBlockEventSub = consensusService.BlockMgrNotifier.SubscribeSyncBlockEvent(consensusService.syncBlockEventChan)
	consensusService.quit = make(chan struct{})
	go consensusService.handlerEvent()

	return nil
}

func (consensusService *ConsensusService) handlerEvent() {
	for {
		select {
		case e := <-consensusService.syncBlockEventChan:
			if e.EventType == event.StartSyncBlock {
				consensusService.pauseForSync = true
				log.Info("Start Sync Blcok")
			} else {
				consensusService.pauseForSync = false
				log.Info("Stop Sync Blcok")
			}
		case <-consensusService.quit:
			return
		}
	}
}

func (consensusService *ConsensusService) Start(executeContext *app.ExecuteContext) error {
	if !consensusService.Config.Miner {
		return nil
	}
	consensusService.start = true
	go func() {
		select {
		case <-consensusService.quit:
			return
		default:
			for {
				if consensusService.pauseForSync {
					time.Sleep(time.Millisecond * 500)
					continue
				}
				log.WithField("Height", consensusService.ChainService.BestChain().Height()).Trace("node start")
				block, err := consensusService.ConsensusEngine.Run(consensusService.Miner)
				if err != nil {
					log.WithField("Reason", err.Error()).Debug("Producer Block Fail")
				} else {
					_, _, err := consensusService.ChainService.ProcessBlock(block)
					if err == nil {
						consensusService.BroadCastor.BroadcastBlock(chainTypes.MsgTypeBlock, block, true)
						log.WithField("Height", block.Header.Height).WithField("txs:", block.Data.TxCount).Info("Process block successfully and broad case block message")
					} else {
						log.WithField("Height", block.Header.Height).WithField("txs:", block.Data.TxCount).WithField("err", err).Info("Process Block fail")
					}
				}
				nextBlockTime, waitSpan := consensusService.getWaitTime()
				log.WithField("nextBlockTime", nextBlockTime).WithField("waitSpan", waitSpan).Debug("Sleep")
				time.Sleep(waitSpan)
			}
		}
	}()

	return nil
}

func (consensusService *ConsensusService) Stop(executeContext *app.ExecuteContext) error {
	if consensusService.Config == nil || !consensusService.Config.Miner {
		return nil
	}

	if consensusService.quit != nil {
		close(consensusService.quit)
	}

	if consensusService.syncBlockEventSub != nil {
		consensusService.syncBlockEventSub.Unsubscribe()
	}

	return nil
}

func (consensusService *ConsensusService) getWaitTime() (time.Time, time.Duration) {
	// max_delay_time +(min_block_interval)*windows = expected_block_interval*windows
	// 6h + 5s*windows = 10s*windows
	// windows = 4320

	lastBlockTime := time.Unix(int64(consensusService.ChainService.BestChain().Tip().TimeStamp), 0)
	targetTime := lastBlockTime.Add(blockInterval)
	now := time.Now()
	if targetTime.Before(now) {
		return now.Add(time.Millisecond * 500), time.Millisecond * 500
	} else {
		return targetTime, targetTime.Sub(now)
	}
	/*
		     window := int64(4320)
		     endBlock := consensusService.DatabaseService.GetHighestBlock().Header
		     if endBlock.Height < window {
				 lastBlockTime := time.Unix(consensusService.DatabaseService.GetHighestBlock().Header.Timestamp, 0)
				 span := time.Now().Sub(lastBlockTime)
				 if span > blockInterval {
					 span = 0
				 } else {
					 span = blockInterval - span
				 }
				 return span
			 }else{
			 	//wait for test
				 startHeight := endBlock.Height - window
				 if startHeight <0 {
					 startHeight = int64(0)
				 }
				 startBlock :=consensusService.DatabaseService.GetBlock(startHeight).Header

				 xx := window * 10 -(time.Unix(startBlock.Timestamp,0).Sub(time.Unix(endBlock.Timestamp,0))).Seconds()

				 span := time.Unix(startBlock.Timestamp,0).Sub(time.Unix(endBlock.Timestamp,0))  //window time
				 avgSpan := span.Nanoseconds()/window
				 return time.Duration(avgSpan) * time.Nanosecond
			 }
	*/
}
