package service

import (
	"github.com/drep-project/drep-chain/app"
	"github.com/drep-project/drep-chain/pkgs/consensus/service/bft"
	"github.com/drep-project/drep-chain/pkgs/consensus/service/solo"
	"gopkg.in/urfave/cli.v1"
)

var (
	ConsensusModeFlag = cli.BoolFlag{
		Name:  "consensusmode",
		Usage: "specify consensus mod(solo bft...)",
	}
)

type ConsensusConfig struct {
	ConsensusMode string          `json:"consensusMode"`
	Solo          solo.SoloConfig `json:"solo"`
	Bft           bft.BftConfig   `json:"bft"`
}
type ConsensusService struct {
	Config   *ConsensusConfig
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
	return nil
}

func (consensusService *ConsensusService) Start(executeContext *app.ExecuteContext) error {
	if executeContext.Cli.GlobalIsSet(ConsensusModeFlag.Name) {
		consensusService.Config.ConsensusMode = executeContext.Cli.GlobalString(ConsensusModeFlag.Name)
	}
	return nil
}

func (consensusService *ConsensusService) Stop(executeContext *app.ExecuteContext) error {
	return nil
}

func (consensusService *ConsensusService) SelectService() app.Service {
	switch consensusService.Config.ConsensusMode {
	case "solo":
		return &solo.ConsensusService{}
	case "bft":
		return &bft.ConsensusService{}
	}
	return nil
}
