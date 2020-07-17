package service

import (
	"fmt"
	"github.com/drep-project/DREP-Chain/app"
	"github.com/drep-project/DREP-Chain/params"
	"github.com/drep-project/DREP-Chain/pkgs/consensus/service/bft"
	"github.com/drep-project/DREP-Chain/pkgs/consensus/service/solo"

	"gopkg.in/urfave/cli.v1"
)

var (
	ConsensusModeFlag = cli.BoolFlag{
		Name:  "consensusmode",
		Usage: "specify consensus mod(solo bft...)",
	}
)

type ConsensusConfig struct {
	ConsensusMode params.NetType   `json:"consensusMode"`
	Solo          *solo.SoloConfig `json:"solo,omitempty"`
	Bft           *bft.BftConfig   `json:"bft,omitempty"`
}
type ConsensusService struct {
	SoloService *solo.SoloConsensusService
	BftService  *bft.BftConsensusService
	Config      *ConsensusConfig
}

func (consensusService *ConsensusService) Name() string {
	return "consensus"
}

func (consensusService *ConsensusService) Api() []app.API {
	return nil
}

func (consensusService *ConsensusService) CommandFlags() ([]cli.Command, []cli.Flag) {
	return nil, []cli.Flag{ConsensusModeFlag}
}

func (consensusService *ConsensusService) Init(executeContext *app.ExecuteContext) error {
	if executeContext.NetConfigType == params.MainnetType || executeContext.NetConfigType == params.TestnetType {
		consensusService.BftService = &bft.BftConsensusService{NetType: executeContext.NetConfigType}
	} else if executeContext.NetConfigType == params.SolonetType {
		consensusService.SoloService = &solo.SoloConsensusService{}
	} else {
		return fmt.Errorf("err param in func consensus service")
	}

	consensusService.Config.ConsensusMode = executeContext.NetConfigType
	return nil
}

func (consensusService *ConsensusService) Start(executeContext *app.ExecuteContext) error {
	//if executeContext.Cli.GlobalIsSet(ConsensusModeFlag.Name) {
	//	consensusService.Config.ConsensusMode = executeContext.Cli.GlobalString(ConsensusModeFlag.Name)
	//}
	return nil
}

func (consensusService *ConsensusService) Stop(executeContext *app.ExecuteContext) error {
	return nil
}

func (consensusService *ConsensusService) SelectService() app.Service {
	switch consensusService.Config.ConsensusMode {
	case params.SolonetType:
		return consensusService.SoloService
	case params.MainnetType, params.TestnetType:
		return consensusService.BftService
	}
	return nil
}

func (consensusService *ConsensusService) DefaultConfig(netType params.NetType) *ConsensusConfig {
	return &ConsensusConfig{
		ConsensusMode: netType,
		//Bft: &bft.DefaultConfig,
	}
}
