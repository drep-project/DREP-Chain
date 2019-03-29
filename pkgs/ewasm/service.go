package dwasm

import (
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/drep-project/drep-chain/app"
	"github.com/drep-project/drep-chain/chain/types"
	"github.com/drep-project/drep-chain/database"
	"gopkg.in/urfave/cli.v1"
)
var (
	DefaultvmConfig = &VMConfig{}
)
type VmService struct {
	DatabaseApi *database.DatabaseService `service:"database"`
	Config *VMConfig
}

func (vmService *VmService) Name() string {
	return "vm"
}

func (vmService *VmService) Api() []app.API {
	return []app.API{}
}

func (vmService *VmService) CommandFlags() ([]cli.Command, []cli.Flag) {
	return nil, []cli.Flag{}
}

func (vmService *VmService) P2pMessages() map[int]interface{} {
	return map[int]interface{}{}
}

func (vmService *VmService) Init(executeContext *app.ExecuteContext) error {
	vmService.Config = DefaultvmConfig
	err := executeContext.UnmashalConfig(vmService.Name(), vmService.Config)
	if err != nil {
		return err
	}
	return nil
}

func (vmService *VmService)  Start(executeContext *app.ExecuteContext) error {
	return nil
}

func (vmService *VmService)  Stop(executeContext *app.ExecuteContext) error{
	return nil
}

func (vmService *VmService)  Receive(context actor.Context) { }

func  (vmService *VmService) ApplyMessage(message *types.Message) (uint64, error) {
    vm := &WasmVm{
   		databaseApi:vmService.DatabaseApi,
	}
    return vm.RunMessage(message)
}