package filter

import (
	"context"

	"gopkg.in/urfave/cli.v1"

	"github.com/drep-project/drep-chain/app"
	"github.com/drep-project/drep-chain/common"
	"github.com/drep-project/drep-chain/common/bloombits"
	"github.com/drep-project/drep-chain/crypto"
	"github.com/drep-project/drep-chain/database"
	"github.com/drep-project/drep-chain/types"
	"github.com/drep-project/drep-chain/pkgs/evm/event"
	"github.com/drep-project/drep-chain/pkgs/evm/vm"
)

type ServiceDatabase interface {
	Get(key []byte) ([]byte, error)
	Put(key []byte, value []byte) error
	Delete(key []byte) error
}

type Backend interface {
	EventMux() *event.TypeMux
	HeaderByNumber(ctx context.Context, blockNr common.BlockNumber) (*types.BlockHeader, error)
	HeaderByHash(ctx context.Context, blockHash crypto.Hash) (*types.BlockHeader, error)
	GetReceipts(ctx context.Context, blockHash crypto.Hash) (types.Receipts, error)
	GetLogs(ctx context.Context, blockHash crypto.Hash) ([][]*types.Log, error)

	SubscribeNewTxsEvent(chan<- vm.NewTxsEvent) event.Subscription
	SubscribeChainEvent(ch chan<- vm.ChainEvent) event.Subscription
	SubscribeRemovedLogsEvent(ch chan<- vm.RemovedLogsEvent) event.Subscription
	SubscribeLogsEvent(ch chan<- []*types.Log) event.Subscription

	BloomStatus() (uint64, uint64)
	ServiceFilter(ctx context.Context, session *bloombits.MatcherSession)
}

type FilterServiceInterface interface {
	app.Service
	ServiceDatabase
	Backend
}

var _ FilterServiceInterface = &FilterService{}

type FilterService struct {
	DatabaseService *database.DatabaseService		`service:"database"`
	apis            []app.API

	chainId types.ChainIdType
}


// implement Service interface
func (service *FilterService) Name() string {
	return MODULENAME
}

func (service *FilterService) Api() []app.API {
	return service.apis
}

func (service *FilterService) CommandFlags() ([]cli.Command, []cli.Flag) {
	return nil, []cli.Flag{EnableFilterFlag}
}

func (service *FilterService) Init(executeContext *app.ExecuteContext) error {

	return nil
}

func (service *FilterService) Start(executeContext *app.ExecuteContext) error {

	return nil
}

func (service *FilterService) Stop(executeContext *app.ExecuteContext) error {

	return nil
}