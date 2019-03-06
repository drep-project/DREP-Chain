package service

import (
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/drep-project/dlog"
	chainTypes "github.com/drep-project/drep-chain/chain/types"
	"github.com/drep-project/drep-chain/common/event"
	"github.com/drep-project/drep-chain/crypto"
	p2pTypes "github.com/drep-project/drep-chain/network/types"
	"strings"
	"time"
)

func (chainService *ChainService) Receive(context actor.Context) {
	var msg interface{}
	msg = context.Message()
	routeMsg, ok := context.Message().(*p2pTypes.RouteIn)
	if ok {
		msg = routeMsg.Detail
	}
	switch msg := msg.(type) {
	case *chainTypes.BlockReq:
		chainService.HandleBlockReqMsg(routeMsg.Peer, msg)
	case *chainTypes.BlockResp:
		go chainService.HandleBlockRespMsg(routeMsg.Peer, msg)
	case *chainTypes.Transaction:
		transaction := msg
		id, _ := transaction.TxId()
		if ForwardedTransaction(id) {
			dlog.Debug("Forwarded this transaction ", "transaction", *transaction)
			return
		}
		// TODO backup nodes should not add
		err := chainService.transactionPool.AddTransaction(transaction)
		if err == nil {
			dlog.Debug("Succeed to add this transaction ", "transaction", *transaction)
			chainService.P2pServer.Broadcast(transaction)
			ForwardTransaction(id)
		} else {
			dlog.Debug("Fail to add this transaction ", "reason", err, "transaction", *transaction)
		}
	case *chainTypes.Block:
		_, isOrPhan, err := chainService.ProcessBlock(msg)
		if err != nil {
			return
		}
		if isOrPhan && routeMsg.Peer!= nil {
			if isOrPhan {
				hash := chainService.GetOrphanRoot(msg.Header.Hash())
				chainService.P2pServer.SendAsync(routeMsg.Peer, &chainTypes.BlockReq{
					StopHash: *hash,
				})
			}
		}
		if chainService.syncingMaxHeight != -1 {
			if chainService.BestChain.Tip().Height >= chainService.syncingMaxHeight {
				chainService.syncBlockEvent.Send(event.SyncBlockEvent{EventType: event.StopSyncBlock})
				chainService.syncingMaxHeight = -1
			}
		}
	case *chainTypes.PeerState:
		chainService.handlePeerState(routeMsg.Peer, msg)
	case *chainTypes.ReqPeerState:
		chainService.handleReqPeerState(routeMsg.Peer, msg)
	}
}

func (chainService *ChainService) HandleBlockReqMsg(peer *p2pTypes.Peer, req *chainTypes.BlockReq) {
	zero := crypto.Hash{}
	startHeight := int64(0)
	endHeight := chainService.BestChain.Tip().Height

	if req.StartHash != zero && chainService.blockExists(&req.StartHash) {
		startHeight = chainService.Index.LookupNode(&req.StartHash).Height
	}

	if req.StopHash != zero {
		block, err := chainService.DatabaseService.GetBlock(&req.StopHash)
		if err != nil {
			return
		}
		endHeight = block.Header.Height
	}

	blocks := []*chainTypes.Block{}
	count := 0
	for i:= startHeight; i <= endHeight; i++ {
		if count%64 == 0&&len(blocks)>0 {
			chainService.P2pServer.Send(peer, &chainTypes.BlockResp{
				Blocks:blocks,
			})
			blocks = []*chainTypes.Block{}
			count = 0
			time.Sleep(time.Second)
		}
		node := chainService.BestChain.NodeByHeight(i)
		block, err := chainService.DatabaseService.GetBlock(node.Hash)
		if err != nil {
			return
		}
		count = count +1
		blocks = append(blocks, block)
	}
	if len(blocks)>0 {
		chainService.P2pServer.Send(peer, &chainTypes.BlockResp{
			Blocks:blocks,
		})
		blocks = []*chainTypes.Block{}
	}

}

func (chainService *ChainService) HandleBlockRespMsg(peer *p2pTypes.Peer, req *chainTypes.BlockResp) {
	if len(req.Blocks) < 1 {
		return
	}
	chainService.syncMaxHeightMut.Lock()

	firstBlock := req.Blocks[0]
	_, isOrPhan, err := chainService.ProcessBlock(firstBlock)
	if err!=nil && !strings.HasPrefix(err.Error(),"already have block") {
		return
	}
	if isOrPhan && peer != nil {
		if isOrPhan {
			hash := chainService.GetOrphanRoot(firstBlock.Header.Hash())
			chainService.P2pServer.SendAsync(peer, &chainTypes.BlockReq{
				StopHash: *hash,
			})
		}
	}
	for i:=1; i < len(req.Blocks); i++ {
		_, isOrPhan, err = chainService.ProcessBlock(req.Blocks[i])
		if err!=nil && !strings.HasPrefix(err.Error(),"already have block") {
			return
		}
	}
	if chainService.syncingMaxHeight != -1 {
		if chainService.BestChain.Tip().Height >= chainService.syncingMaxHeight {
			chainService.syncBlockEvent.Send(event.SyncBlockEvent{EventType: event.StopSyncBlock})
			chainService.syncingMaxHeight = -1
		}
	}
	chainService.syncMaxHeightMut.Unlock()
}

