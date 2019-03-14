package service

import (
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/drep-project/dlog"
	chainTypes "github.com/drep-project/drep-chain/chain/types"
	"github.com/drep-project/drep-chain/crypto"
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
		go chainService.HandleBlockReqMsg(routeMsg.Peer, msg)
	case *chainTypes.BlockResp:
		go chainService.HandleBlockRespMsg(msg)
	case *chainTypes.Transaction:
		transaction := msg
		//id, _ := transaction.TxId()
		//if ForwardedTransaction(id) {
		//	dlog.Debug("Forwarded this transaction ", "transaction", *transaction)
		//	return
		//}
		// TODO backup nodes should not add
		err := chainService.transactionPool.AddTransaction(transaction)
		if err == nil {
			dlog.Debug("Succeed to add this transaction ", "transaction", *transaction)
			chainService.P2pServer.Broadcast(transaction)
			//ForwardTransaction(id)
		} else {
			dlog.Debug("Fail to add this transaction ", "reason", err, "transaction", *transaction)
		}
	case *chainTypes.Block:
		_, isOrPhan, err := chainService.ProcessBlock(msg)
		if err != nil {
			return
		}
		if isOrPhan {
			// todo ��ʼ����ͬ��
			//chainService.synchronise()
		}
	case *chainTypes.PeerState:
		go chainService.handlePeerState(routeMsg.Peer, msg)
	case *chainTypes.ReqPeerState:
		go chainService.handleReqPeerState(routeMsg.Peer, msg)
	case *chainTypes.HeaderReq:
		go chainService.handleHeaderReq(routeMsg.Peer, msg)
	case *chainTypes.HeaderRsp:
		go chainService.handleHeaderRsp(routeMsg.Peer, msg)
	}
}

func (chainService *ChainService) handleHeaderReq(peer *p2pTypes.Peer, req *chainTypes.HeaderReq) {
	headers := make([]chainTypes.BlockHeader, 0, req.ToHeight-req.FromHeight+1)
	for i := req.FromHeight; i <= req.ToHeight; i++ {
		node := chainService.BestChain.NodeByHeight(i)
		if node != nil {
			headers = append(headers, node.Header())
		}
	}

	dlog.Info("header req len", "total header", len(headers), "from", req.FromHeight, "to", req.ToHeight)
	chainService.P2pServer.Send(peer, chainTypes.HeaderRsp{Headers: headers})
}

func (chainService *ChainService) handleHeaderRsp(peer *p2pTypes.Peer, rsp *chainTypes.HeaderRsp) {
	headerHashs := make([]*syncHeaderHash, 0, len(rsp.Headers))
	for _, h := range rsp.Headers {
		headerHashs = append(headerHashs, &syncHeaderHash{headerHash: h.Hash(), height: h.Height})
	}

	err := chainService.checkHeaderChain(rsp.Headers)
	if err != nil {
		dlog.Info("handleHeaderRsp", "err", err)
		return
	}

	dlog.Info("handleHeaderRsp ", "total len:", len(headerHashs), "from height:", headerHashs[0].height, "end height:", headerHashs[len(headerHashs)-1].height)
	chainService.headerHashCh <- headerHashs
}

func (chainService *ChainService) HandleBlockReqMsg(peer *p2pTypes.Peer, req *chainTypes.BlockReq) {
	dlog.Info("sync req block", "num:", len(req.BlockHashs))
	zero := crypto.Hash{}
	startHeight := int64(0)
	endHeight := chainService.BestChain.Tip().Height

	if len(req.BlockHashs) == 0 {
		dlog.Warn("handle block req", "block hash num", len(req.BlockHashs))
		return
	}
	startHash := req.BlockHashs[0]
	endHash := req.BlockHashs[len(req.BlockHashs)-1]
	if startHash != zero && chainService.blockExists(&startHash) {
		startHeight = chainService.Index.LookupNode(&startHash).Height
	}

	if endHash != zero {
		block, err := chainService.DatabaseService.GetBlock(&endHash)
		if err != nil {
			return
		}
		endHeight = block.Header.Height
	}

	blocks := []*chainTypes.Block{}
	count := 0
	for i := startHeight; i <= endHeight; i++ {
		if count%maxBlockCountReq == 0 && len(blocks) > 0 {
			chainService.P2pServer.Send(peer, &chainTypes.BlockResp{
				Blocks: blocks,
			})
			blocks = []*chainTypes.Block{}
			count = 0
		}

		node := chainService.BestChain.NodeByHeight(i)
		block, err := chainService.DatabaseService.GetBlock(node.Hash)
		if err != nil {
			return
		}
		count = count + 1
		blocks = append(blocks, block)
	}
	if len(blocks) > 0 {
		chainService.P2pServer.Send(peer, &chainTypes.BlockResp{
			Blocks: blocks,
		})
		dlog.Info("req blocks and rsp:", "num", len(blocks))
		blocks = []*chainTypes.Block{}
	}
}

func (chainService *ChainService) HandleBlockRespMsg(rsp *chainTypes.BlockResp) {
	chainService.blocksCh <- rsp.Blocks
}
