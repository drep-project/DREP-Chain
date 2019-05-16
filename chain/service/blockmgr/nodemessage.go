package service

import (
	"time"

	"github.com/drep-project/dlog"
	chainTypes "github.com/drep-project/drep-chain/chain/types"
	"github.com/drep-project/drep-chain/crypto"
	"github.com/drep-project/drep-chain/network/p2p"
	"github.com/pkg/errors"
)

func (blockMgr *BlockMgr) receiveMsg(peer *chainTypes.PeerInfo, rw p2p.MsgReadWriter) error {
	//1 与peer同步一下状态
	timeout := time.After(time.Second * maxNetworkTimeout)
	errCh := make(chan error)
	msgCh := make(chan p2p.Msg)
	go func() {
		blockMgr.P2pServer.Send(peer.GetMsgRW(), chainTypes.MsgTypePeerStateReq, &chainTypes.PeerState{Height: uint64(blockMgr.ChainService.BestChain.Height())})
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
		return ErrReqStateTimeout
	case msg := <-msgCh:
		switch msg.Code {
		case chainTypes.MsgTypePeerStateReq:
			var req chainTypes.PeerStateReq
			if err := msg.Decode(&req); err != nil {
				return errors.Wrapf(ErrDecodeMsg, "PeerStateReq msg:%v err:%v", msg, err)
			}
			blockMgr.handleReqPeerState(peer, &req)
		default:
			return errors.Wrapf(ErrMsgType, "expected type:%d, receive type:%d", chainTypes.MsgTypePeerStateReq, msg.Code)
		}
	}

	//通知给同步协程
	blockMgr.newPeerCh <- peer

	//2 处理所有消息
	return blockMgr.dealMsg(peer, rw)
}

func (blockMgr *BlockMgr) dealMsg(peer *chainTypes.PeerInfo, rw p2p.MsgReadWriter) error {
	dlog.Info("new peer", "peer addr:", peer.GetAddr())
	for {
		msg, err := rw.ReadMsg()
		if err != nil {
			dlog.Info("blockMgr receive msg", "err", err)
			return err
		}

		if msg.Size > chainTypes.MaxMsgSize {
			return ErrOverFlowMaxMsgSize
		}

		switch msg.Code {
		case chainTypes.MsgTypeBlockReq:
			var req chainTypes.BlockReq
			if err := msg.Decode(&req); err != nil {
				return errors.Wrapf(ErrDecodeMsg, "MsgTypeBlockReq msg:%v err:%v", msg, err)
			}
			go blockMgr.HandleBlockReqMsg(peer, &req)
		case chainTypes.MsgTypeBlockResp:
			var resp chainTypes.BlockResp
			if err := msg.Decode(&resp); err != nil {
				return errors.Wrapf(ErrDecodeMsg, "BlockResp msg:%v err:%v", msg, err)
			}
			go blockMgr.HandleBlockRespMsg(&resp)
		case chainTypes.MsgTypeTransaction:
			var txs []*chainTypes.Transaction
			if err := msg.Decode(&txs); err != nil {
				return errors.Wrapf(ErrDecodeMsg, "Transactions msg:%v err:%v", msg, err)
			}

			// TODO backup nodes should not add
			for _, tx := range txs {
				dlog.Trace("comming transaction", "transaction", tx.Nonce())
				tx := tx
				peer.MarkTx(tx)
				blockMgr.SendTransaction(tx, false)
			}

		case chainTypes.MsgTypeBlock:
			var newBlock chainTypes.Block
			if err := msg.Decode(&newBlock); err != nil {
				return errors.Wrapf(ErrDecodeMsg, "Block msg:%v err:%v", msg, err)
			}

			_, isOrPhan, err := blockMgr.ChainService.ProcessBlock(&newBlock)
			if err == nil {
				//return err
			}

			peer.MarkBlock(&newBlock)
			blockMgr.BroadcastBlock(chainTypes.MsgTypeBlock, &newBlock, false)

			if isOrPhan {
				// todo 触发同步
				//blockMgr.synchronise()
			}
		case chainTypes.MsgTypePeerState:
			var resp chainTypes.PeerState
			if err := msg.Decode(&resp); err != nil {
				return errors.Wrapf(ErrDecodeMsg, "PeerState msg:%v err:%v", msg, err)
			}
			go blockMgr.handlePeerState(peer, &resp)
		case chainTypes.MsgTypePeerStateReq:
			var req chainTypes.PeerStateReq
			if err := msg.Decode(&req); err != nil {
				return errors.Wrapf(ErrDecodeMsg, "PeerStateReq msg:%v err:%v", msg, err)
			}
			go blockMgr.handleReqPeerState(peer, &req)
		case chainTypes.MsgTypeHeaderReq:
			var req chainTypes.HeaderReq
			if err := msg.Decode(&req); err != nil {
				return errors.Wrapf(ErrDecodeMsg, "HeaderReq msg:%v err:%v", msg, err)
			}
			go blockMgr.handleHeaderReq(peer, &req)
		case chainTypes.MsgTypeHeaderRsp:
			var resp chainTypes.HeaderRsp
			if err := msg.Decode(&resp); err != nil {
				return errors.Wrapf(ErrDecodeMsg, "HeaderRsp msg:%v err:%v", msg, err)
			}
			go blockMgr.handleHeaderRsp(peer, &resp)
		}
	}

	return nil
}

func (blockMgr *BlockMgr) handleHeaderReq(peer *chainTypes.PeerInfo, req *chainTypes.HeaderReq) {
	headers := make([]chainTypes.BlockHeader, 0, req.ToHeight-req.FromHeight+1)
	for i := req.FromHeight; i <= req.ToHeight; i++ {
		node := blockMgr.ChainService.BestChain.NodeByHeight(uint64(i))
		if node != nil {
			headers = append(headers, node.Header())
		}
	}

	dlog.Info("header req len", "total header", len(headers), "from", req.FromHeight, "to", req.ToHeight)
	blockMgr.P2pServer.Send(peer.GetMsgRW(), uint64(chainTypes.MsgTypeHeaderRsp), chainTypes.HeaderRsp{Headers: headers})
}

func (blockMgr *BlockMgr) handleHeaderRsp(peer *chainTypes.PeerInfo, rsp *chainTypes.HeaderRsp) {
	headerHashs := make([]*syncHeaderHash, 0, len(rsp.Headers))
	for _, h := range rsp.Headers {
		headerHashs = append(headerHashs, &syncHeaderHash{headerHash: h.Hash(), height: h.Height})
	}

	//请求的相关协程要关闭。
	err := blockMgr.checkHeaderChain(rsp.Headers)
	if err != nil {
		dlog.Info("handleHeaderRsp", "err", err)
		return
	}

	if len(headerHashs) >= 1 {
		dlog.Info("handleHeaderRsp ", "total len:", len(headerHashs), "from height:", headerHashs[0].height, "end height:", headerHashs[len(headerHashs)-1].height)
	} else {
		dlog.Error("handleHeaderRsp rsp nil")
		return
	}

	blockMgr.headerHashCh <- headerHashs
}

func (blockMgr *BlockMgr) HandleBlockReqMsg(peer *chainTypes.PeerInfo, req *chainTypes.BlockReq) {
	dlog.Info("sync req block", "num:", len(req.BlockHashs))
	zero := crypto.Hash{}
	startHeight := uint64(0)
	endHeight := blockMgr.ChainService.BestChain.Tip().Height

	if len(req.BlockHashs) == 0 {
		dlog.Warn("handle block req", "block hash num", len(req.BlockHashs))
		return
	}
	startHash := req.BlockHashs[0]
	endHash := req.BlockHashs[len(req.BlockHashs)-1]
	if startHash != zero && blockMgr.ChainService.BlockExists(&startHash) {
		startHeight = blockMgr.ChainService.Index.LookupNode(&startHash).Height
	}

	if endHash != zero {
		block, err := blockMgr.DatabaseService.GetBlock(&endHash)
		if err != nil {
			return
		}
		endHeight = block.Header.Height
	}

	blocks := []*chainTypes.Block{}
	count := 0
	for i := startHeight; i <= endHeight; i++ {
		if count%maxBlockCountReq == 0 && len(blocks) > 0 {
			blockMgr.P2pServer.Send(peer.GetMsgRW(), chainTypes.MsgTypeBlockResp, &chainTypes.BlockResp{
				Blocks: blocks,
			})
			blocks = []*chainTypes.Block{}
			count = 0
		}

		node := blockMgr.ChainService.BestChain.NodeByHeight(i)
		block, err := blockMgr.DatabaseService.GetBlock(node.Hash)
		if err != nil {
			return
		}
		count = count + 1
		blocks = append(blocks, block)
	}
	if len(blocks) > 0 {
		blockMgr.P2pServer.Send(peer.GetMsgRW(), chainTypes.MsgTypeBlockResp, &chainTypes.BlockResp{
			Blocks: blocks,
		})
		dlog.Info("req blocks and rsp:", "num", len(blocks))
		//blocks = []*chainTypes.Block{}
	}
}

func (blockMgr *BlockMgr) HandleBlockRespMsg(rsp *chainTypes.BlockResp) {
	blockMgr.blocksCh <- rsp.Blocks
}
