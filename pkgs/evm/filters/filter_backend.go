package filters

import (
    "github.com/drep-project/drep-chain/chain/types"
    "github.com/drep-project/drep-chain/common/event"
    "context"
    "github.com/drep-project/drep-chain/rpc"
    "github.com/drep-project/drep-chain/chain/service/chainservice"
    "github.com/drep-project/drep-chain/crypto"
    "github.com/drep-project/drep-chain/pkgs/evm/bloombits"
    "github.com/drep-project/drep-chain/pkgs/evm/vm"
)

type FilterBackend struct {
    BlockChain chainservice.ChainService
}

func (fb *FilterBackend) Chain() chainservice.ChainService { return fb.BlockChain }
func (fb *FilterBackend) EventMux() *event.TypeMux         { panic("not supported") }

func (fb *FilterBackend) HeaderByNumber(ctx context.Context, block rpc.BlockNumber) (*types.BlockHeader, error) {
    if block == rpc.LatestBlockNumber {
        return fb.BlockChain.GetCurrentHeader()
    }
    return fb.BlockChain.GetBlockHeaderByHeight(uint64(block.Int64()))
}

func (fb *FilterBackend) HeaderByHash(ctx context.Context, hash crypto.Hash) (*types.BlockHeader, error) {
    return fb.BlockChain.GetBlockHeaderByHash(&hash)
}

func (fb *FilterBackend) GetReceipts(ctx context.Context, hash crypto.Hash) (types.Receipts, error) {
    number := fb.BlockChain.GetHeight(hash)
    if number == nil {
        return nil, nil
    }
    return fb.BlockChain.DatabaseService.GetReceipts(hash), nil
}

func (fb *FilterBackend) GetLogs(ctx context.Context, hash crypto.Hash) ([][]*types.Log, error) {
    number := fb.BlockChain.GetHeight(hash)
    if number == nil {
        return nil, nil
    }
    receipts := fb.BlockChain.DatabaseService.GetReceipts(hash)
    if receipts == nil {
        return nil, nil
    }
    logs := make([][]*types.Log, len(receipts))
    for i, receipt := range receipts {
        logs[i] = receipt.Logs
    }
    return logs, nil
}

func (fb *FilterBackend) SubscribeNewTxsEvent(ch chan<- vm.NewTxsEvent) event.Subscription {
    return event.NewSubscription(func(quit <-chan struct{}) error {
        <-quit
        return nil
    })
}
func (fb *FilterBackend) SubscribeChainEvent(ch chan<- vm.ChainEvent) event.Subscription {
    return fb.BlockChain.SubscribeChainEvent(ch)
}
func (fb *FilterBackend) SubscribeRemovedLogsEvent(ch chan<- vm.RemovedLogsEvent) event.Subscription {
    return fb.BlockChain.SubscribeRemovedLogsEvent(ch)
}
func (fb *FilterBackend) SubscribeLogsEvent(ch chan<- []*types.Log) event.Subscription {
    return fb.BlockChain.SubscribeLogsEvent(ch)
}

func (fb *FilterBackend) BloomStatus() (uint64, uint64) { return 4096, 0 }
func (fb *FilterBackend) ServiceFilter(ctx context.Context, ms *bloombits.MatcherSession) {
    panic("not supported")
}
