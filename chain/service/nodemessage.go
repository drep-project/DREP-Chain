package service

import (
	"errors"
	"fmt"
	"github.com/drep-project/dlog"
	chainTypes "github.com/drep-project/drep-chain/chain/types"
	"github.com/drep-project/drep-chain/crypto"
	"github.com/drep-project/drep-chain/network/p2p"
	"time"
)

func (chainService *ChainService) receiveMsg(peer *chainTypes.PeerInfo, rw p2p.MsgReadWriter) error {

	//1 与peer同步一下状态
	timeout := time.After(time.Second * maxNetworkTimeout)
	errCh := make(chan error)
	msgCh := make(chan p2p.Msg)
	go func() {
		chainService.P2pServer.Send(peer.GetMsgRW(), chainTypes.MsgTypePeerStateReq, &chainTypes.PeerState{Height: uint64(chainService.BestChain.Height())})
		msg, err := rw.ReadMsg()
		if err != nil {
			errCh <- err
			return
		}
		msgCh <- msg
	}()

	select {
	case err := <-errCh:
		return err
	case <-timeout:
		return fmt.Errorf("req state timeout")
	case msg := <-msgCh:
		switch msg.Code {
		case chainTypes.MsgTypePeerStateReq:
			var req chainTypes.PeerStateReq
			if err := msg.Decode(&req); err != nil {
				return fmt.Errorf("peerstatereq msg:%v err:%v", msg, err)
			}
			chainService.handleReqPeerState(peer, &req)
		default:
			return fmt.Errorf("err msg type:%d", msg.Code)
		}
	}


	//通知给同步协程
	chainService.newPeerCh <- peer

	//2 处理所有消息
	return chainService.dealMsg(peer, rw)
}

func (chainService *ChainService) dealMsg(peer *chainTypes.PeerInfo, rw p2p.MsgReadWriter) error {
	dlog.Info("new peer", "peer addr:" ,peer.GetAddr())
	for {
		msg, err := rw.ReadMsg()
		if err != nil {
			dlog.Info("chainService receive msg", "err", err)
			return err
		}

		if msg.Size > chainTypes.MaxMsgSize {
			return errors.New("err msg size")
		}

		switch msg.Code {
		case chainTypes.MsgTypeBlockReq:
			var req chainTypes.BlockReq
			if err := msg.Decode(&req); err != nil {
				return fmt.Errorf("blockreq msg:%v err:%v", msg, err)
			}
			go chainService.HandleBlockReqMsg(peer, &req)
		case chainTypes.MsgTypeBlockResp:
			var resp chainTypes.BlockResp
			if err := msg.Decode(&resp); err != nil {
				return fmt.Errorf("block resp msg:%v err:%v", msg, err)
			}
			go chainService.HandleBlockRespMsg(&resp)
		case chainTypes.MsgTypeTransaction:
			var txs chainTypes.Transactions
			if err := msg.Decode(&txs); err != nil {
				return fmt.Errorf("tx msg:%v err:%v", msg, err)
			}

			// TODO backup nodes should not add
			for _,tx := range txs {
				dlog.Trace("comming transaction", "transaction", tx.Nonce())
				tx := tx
				peer.MarkTx(&tx)
				err := chainService.transactionPool.AddTransaction(&tx)
				if err == nil {
					dlog.Debug("Succeed to add this transaction ", "transaction", tx)
					chainService.BroadcastTx(chainTypes.MsgTypeTransaction, &tx, false)
				} else {
					dlog.Debug("Fail to add this transaction ", "reason", err, "transaction", tx)
				}
			}

		case chainTypes.MsgTypeBlock:
			var newBlock chainTypes.Block
			if err := msg.Decode(&newBlock); err != nil {
				return fmt.Errorf("new block msg:%v err:%v", msg, err)
			}

			_, isOrPhan, err := chainService.ProcessBlock(&newBlock)
			if err != nil {
				return err
			}

			peer.MarkBlock(&newBlock)
			chainService.BroadcastBlock(chainTypes.MsgTypeBlock, &newBlock, false)

			if isOrPhan {
				// todo 触发同步
				//chainService.synchronise()
			}
		case chainTypes.MsgTypePeerState:
			var resp chainTypes.PeerState
			if err := msg.Decode(&resp); err != nil {
				return fmt.Errorf("peerstate rsp msg:%v err:%v", msg, err)
			}
			go chainService.handlePeerState(peer, &resp)
		case chainTypes.MsgTypePeerStateReq:
			var req chainTypes.PeerStateReq
			if err := msg.Decode(&req); err != nil {
				return fmt.Errorf("peerstatereq msg:%v err:%v", msg, err)
			}
			go chainService.handleReqPeerState(peer, &req)
		case chainTypes.MsgTypeHeaderReq:
			var req chainTypes.HeaderReq
			if err := msg.Decode(&req); err != nil {
				return fmt.Errorf("peerstate rsp msg:%v err:%v", msg, err)
			}
			go chainService.handleHeaderReq(peer, &req)
		case chainTypes.MsgTypeHeaderRsp:
			var resp chainTypes.HeaderRsp
			if err := msg.Decode(&resp); err != nil {
				return fmt.Errorf("peerstate rsp msg:%v err:%v", msg, err)
			}
			go chainService.handleHeaderRsp(peer, &resp)
		}
	}

	return nil
}

func (chainService *ChainService) handleHeaderReq(peer *chainTypes.PeerInfo, req *chainTypes.HeaderReq) {
	headers := make([]chainTypes.BlockHeader, 0, req.ToHeight-req.FromHeight+1)
	for i := req.FromHeight; i <= req.ToHeight; i++ {
		node := chainService.BestChain.NodeByHeight(uint64(i))
		if node != nil {
			headers = append(headers, node.Header())
		}
	}

	dlog.Info("header req len", "total header", len(headers), "from", req.FromHeight, "to", req.ToHeight)
	chainService.P2pServer.Send(peer.GetMsgRW(), uint64(chainTypes.MsgTypeHeaderRsp), chainTypes.HeaderRsp{Headers: headers})
}

func (chainService *ChainService) handleHeaderRsp(peer *chainTypes.PeerInfo, rsp *chainTypes.HeaderRsp) {
	headerHashs := make([]*syncHeaderHash, 0, len(rsp.Headers))
	for _, h := range rsp.Headers {
		headerHashs = append(headerHashs, &syncHeaderHash{headerHash: h.Hash(), height: h.Height})
	}

	//请求的相关协程要关闭。
	err := chainService.checkHeaderChain(rsp.Headers)
	if err != nil {
		dlog.Info("handleHeaderRsp", "err", err)
		return
	}

	if len(headerHashs) >= 1 {
		dlog.Info("handleHeaderRsp ", "total len:", len(headerHashs), "from height:", headerHashs[0].height, "end height:", headerHashs[len(headerHashs)-1].height)
	} else {
		dlog.Info("handleHeaderRsp rsp nil")
	}

	chainService.headerHashCh <- headerHashs
}

func (chainService *ChainService) HandleBlockReqMsg(peer *chainTypes.PeerInfo, req *chainTypes.BlockReq) {
	dlog.Info("sync req block", "num:", len(req.BlockHashs))
	zero := crypto.Hash{}
	startHeight := uint64(0)
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
			chainService.P2pServer.Send(peer.GetMsgRW(), chainTypes.MsgTypeBlockResp, &chainTypes.BlockResp{
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
		chainService.P2pServer.Send(peer.GetMsgRW(), chainTypes.MsgTypeBlockResp, &chainTypes.BlockResp{
			Blocks: blocks,
		})
		dlog.Info("req blocks and rsp:", "num", len(blocks))
		//blocks = []*chainTypes.Block{}
	}
}

func (chainService *ChainService) HandleBlockRespMsg(rsp *chainTypes.BlockResp) {
	chainService.blocksCh <- rsp.Blocks
}
