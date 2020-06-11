package solo

import (
	"bytes"
	"encoding/binary"
	"github.com/drep-project/DREP-Chain/chain/store"
	"github.com/drep-project/DREP-Chain/crypto"
	"github.com/drep-project/DREP-Chain/crypto/secp256k1"
	"time"

	"github.com/drep-project/DREP-Chain/app"
	blockMgrService "github.com/drep-project/DREP-Chain/blockmgr"
	chainService "github.com/drep-project/DREP-Chain/chain"
	"github.com/drep-project/DREP-Chain/common/event"
	"github.com/drep-project/DREP-Chain/database"
	p2pService "github.com/drep-project/DREP-Chain/network/service"
	accountService "github.com/drep-project/DREP-Chain/pkgs/accounts/service"
	consensusTypes "github.com/drep-project/DREP-Chain/pkgs/consensus/types"
	chainTypes "github.com/drep-project/DREP-Chain/types"
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

type SoloConsensusService struct {
	P2pServer        p2pService.P2P                       `service:"p2p"`
	ChainService     chainService.ChainServiceInterface   `service:"chain"`
	BroadCastor      blockMgrService.ISendMessage         `service:"blockmgr"`
	BlockMgrNotifier blockMgrService.IBlockNotify         `service:"blockmgr"`
	BlockGenerator   blockMgrService.IBlockBlockGenerator `service:"blockmgr"`
	DatabaseService  *database.DatabaseService            `service:"database"`
	WalletService    *accountService.AccountService       `service:"accounts"`

	Config             *SoloConfig
	syncBlockEventSub  event.Subscription
	syncBlockEventChan chan event.SyncBlockEvent
	ConsensusEngine    consensusTypes.IConsensusEngine
	Miner              *secp256k1.PrivateKey
	//During the process of synchronizing blocks, the miner stopped mining
	pauseForSync bool
	start        bool
	quit         chan struct{}
}

func (soloConsensusService *SoloConsensusService) Name() string {
	return "solo"
}

func (soloConsensusService *SoloConsensusService) Api() []app.API {
	return nil
}

func (soloConsensusService *SoloConsensusService) CommandFlags() ([]cli.Command, []cli.Flag) {
	return nil, []cli.Flag{EnableSoloConsensusFlag}
}

func (soloConsensusService *SoloConsensusService) Init(executeContext *app.ExecuteContext) error {
	if executeContext.Cli.GlobalIsSet(EnableSoloConsensusFlag.Name) {
		soloConsensusService.Config.StartMiner = executeContext.Cli.GlobalBool(EnableSoloConsensusFlag.Name)
	}

	soloConsensusService.ChainService.AddBlockValidator(NewSoloValidator(soloConsensusService.Config.MyPk))
	if !soloConsensusService.Config.StartMiner {
		return nil
	} else {
		if soloConsensusService.WalletService.Wallet == nil {
			return ErrWalletNotOpen
		}
	}

	value := make([]byte, 0)
	buffer := bytes.NewBuffer(value)
	binary.Write(buffer, binary.BigEndian, uint64(soloConsensusService.Config.ChangeInterval))
	soloConsensusService.DatabaseService.LevelDb().Put([]byte(store.ChangeInterval), buffer.Bytes())

	var engine consensusTypes.IConsensusEngine
	engine = NewSoloConsensus(
		soloConsensusService.ChainService,
		soloConsensusService.BlockGenerator,
		soloConsensusService.Config.MyPk,
		soloConsensusService.DatabaseService,
		soloConsensusService.Config)

	soloConsensusService.ConsensusEngine = engine
	soloConsensusService.syncBlockEventChan = make(chan event.SyncBlockEvent)
	soloConsensusService.syncBlockEventSub = soloConsensusService.BlockMgrNotifier.SubscribeSyncBlockEvent(soloConsensusService.syncBlockEventChan)
	soloConsensusService.quit = make(chan struct{})
	go soloConsensusService.handlerEvent()

	return nil
}

func (soloConsensusService *SoloConsensusService) handlerEvent() {
	for {
		select {
		case e := <-soloConsensusService.syncBlockEventChan:
			if e.EventType == event.StartSyncBlock {
				soloConsensusService.pauseForSync = true
				log.Info("Start Sync Blcok")
			} else {
				soloConsensusService.pauseForSync = false
				log.Info("Stop Sync Blcok")
			}
		case <-soloConsensusService.quit:
			return
		}
	}
}

func (soloConsensusService *SoloConsensusService) Start(executeContext *app.ExecuteContext) error {
	if !soloConsensusService.Config.StartMiner {
		return nil
	}
	soloConsensusService.start = true
	go func() {
		select {
		case <-soloConsensusService.quit:
			return
		default:
			for {
				//consult privkey in wallet
				accountNode, err := soloConsensusService.WalletService.Wallet.GetAccountByPubkey(soloConsensusService.Config.MyPk)
				if err != nil {

					log.WithField("init err", err).WithField("addr", crypto.PubkeyToAddress(soloConsensusService.Config.MyPk).String()).Error("privkey of MyPk in Config is not in local wallet")

					time.Sleep(time.Second * 3)
					continue
				}
				soloConsensusService.Miner = accountNode.PrivateKey
				if soloConsensusService.pauseForSync {
					time.Sleep(time.Millisecond * 500)
					continue
				}
				log.WithField("Height", soloConsensusService.ChainService.BestChain().Height()).Trace("node start")
				block, err := soloConsensusService.ConsensusEngine.Run(soloConsensusService.Miner)
				if err != nil {
					log.WithField("Reason", err.Error()).Debug("Producer Block Fail")
				} else {
					_, _, err := soloConsensusService.ChainService.ProcessBlock(block)
					if err == nil {
						soloConsensusService.BroadCastor.BroadcastBlock(chainTypes.MsgTypeBlock, block, true)
						log.WithField("Height", block.Header.Height).WithField("txs:", block.Data.TxCount).Info("Process block successfully and broad case block message")
					} else {
						log.WithField("Height", block.Header.Height).WithField("txs:", block.Data.TxCount).WithField("err", err).Info("Process Block fail")
					}

				}
				nextBlockTime, waitSpan := soloConsensusService.getWaitTime()
				log.WithField("nextBlockTime", nextBlockTime).WithField("waitSpan", waitSpan).Debug("Sleep")
				time.Sleep(waitSpan)
			}
		}
	}()

	return nil
}

func (soloConsensusService *SoloConsensusService) Stop(executeContext *app.ExecuteContext) error {
	if soloConsensusService.Config == nil || !soloConsensusService.Config.StartMiner {
		return nil
	}

	if soloConsensusService.quit != nil {
		close(soloConsensusService.quit)
	}

	if soloConsensusService.syncBlockEventSub != nil {
		soloConsensusService.syncBlockEventSub.Unsubscribe()
	}

	return nil
}

func (soloConsensusService *SoloConsensusService) getWaitTime() (time.Time, time.Duration) {
	// max_delay_time +(min_block_interval)*windows = expected_block_interval*windows
	// 6h + 5s*windows = 10s*windows
	// windows = 4320

	lastBlockTime := time.Unix(int64(soloConsensusService.ChainService.BestChain().Tip().TimeStamp), 0)
	targetTime := lastBlockTime.Add(time.Duration(int(time.Second) * soloConsensusService.Config.BlockInterval))
	now := time.Now()
	if targetTime.Before(now) {
		return now.Add(time.Millisecond * 20), time.Millisecond * 20
	} else {
		return targetTime, targetTime.Sub(now)
	}
}
func (soloConsensusService *SoloConsensusService) DefaultConfig() *SoloConfig {
	return &SoloConfig{
		BlockInterval: 7,
	}
}
