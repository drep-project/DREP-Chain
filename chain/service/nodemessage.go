package service

import (
	"github.com/AsynkronIT/protoactor-go/actor"
	chainTypes "github.com/drep-project/drep-chain/chain/types"
	"github.com/drep-project/dlog"
	p2pTypes "github.com/drep-project/drep-chain/network/types"
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
			chainService.ProcessBlockReq(routeMsg.Peer, msg)
		case *chainTypes.BlockResp:
			go func() {
				for _, block := range msg.Blocks {
					chainService.ProcessBlock(block)
				}
			}()

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
			block := msg
			if block.Header.Height <= chainService.DatabaseService.GetMaxHeight() {
				return
			}
			id, _ := block.BlockHashHex()
			if ForwardedBlock(id) { // if forwarded, then processed. later this will be read from db
				dlog.Debug("Forwarded this block ", "block" ,*block)
				return
			}
			ForwardBlock(id)
			chainService.ProcessBlock(block)
		case *chainTypes.PeerState:
			chainService.handlePeerState(routeMsg.Peer, msg)
		case *chainTypes.ReqPeerState:
			chainService.handleReqPeerState(routeMsg.Peer, msg)
		}
}