package blockmgr

import (
	"time"

	"github.com/drep-project/DREP-Chain/crypto"
	"github.com/drep-project/DREP-Chain/network/p2p"
	"github.com/drep-project/DREP-Chain/types"
	"github.com/pkg/errors"
)

func (blockMgr *BlockMgr) receiveMsg(peer *types.PeerInfo, rw p2p.MsgReadWriter) error {
	//1 Synchronize the state with the peer
	timeout := time.After(time.Second * maxNetworkTimeout)
	errCh := make(chan error)
	msgCh := make(chan p2p.Msg)
	go func() {
		blockMgr.P2pServer.Send(peer.GetMsgRW(), types.MsgTypePeerStateReq, &types.PeerState{Height: uint64(blockMgr.ChainService.BestChain().Height())})
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
		case types.MsgTypePeerStateReq:
			var req types.PeerStateReq
			if err := msg.Decode(&req); err != nil {
				return errors.Wrapf(ErrDecodeMsg, "PeerStateReq msg:%v err:%v", msg, err)
			}
			blockMgr.handleReqPeerState(peer, &req)
		default:
			return errors.Wrapf(ErrMsgType, "expected type:%d, receive type:%d", types.MsgTypePeerStateReq, msg.Code)
		}
	}

	//Notify the synchronization coroutine
	blockMgr.newPeerCh <- peer

	//2 Process all messages
	return blockMgr.dealMsg(peer, rw)
}

func (blockMgr *BlockMgr) dealMsg(peer *types.PeerInfo, rw p2p.MsgReadWriter) error {

	log.WithField("peer addr:", peer.GetAddr()).Info("new peer receive")

	for {
		msg, err := rw.ReadMsg()
		if err != nil {
			log.WithField("Reason", err).WithField("ip", peer.GetAddr()).Info("receive blockMgr msg fail")
			return err
		}

		if msg.Size > types.MaxMsgSize {
			return ErrOverFlowMaxMsgSize
		}

		switch msg.Code {
		case types.MsgTypeBlockReq:
			var req types.BlockReq
			if err := msg.Decode(&req); err != nil {
				return errors.Wrapf(ErrDecodeMsg, "MsgTypeBlockReq msg:%v err:%v", msg, err)
			}
			go blockMgr.HandleBlockReqMsg(peer, &req)
		case types.MsgTypeBlockResp:
			var resp types.BlockResp
			if err := msg.Decode(&resp); err != nil {
				return errors.Wrapf(ErrDecodeMsg, "BlockResp msg:%v err:%v", msg, err)
			}
			go blockMgr.HandleBlockRespMsg(peer, &resp)
		case types.MsgTypeTransaction:
			var txs []*types.Transaction
			if err := msg.Decode(&txs); err != nil {
				return errors.Wrapf(ErrDecodeMsg, "Transactions msg:%v err:%v", msg, err)
			}

			// TODO backup nodes should not add
			for _, tx := range txs {
				from, _ := tx.From()
				log.WithField("transaction", tx.Nonce()).WithField("from", from.String()).Trace("comming transaction")
				tx := tx
				peer.MarkTx(tx)
				blockMgr.SendTransaction(tx, false)
			}

		case types.MsgTypeBlock:
			var newBlock types.Block
			if err := msg.Decode(&newBlock); err != nil {
				return errors.Wrapf(ErrDecodeMsg, "Block msg:%v err:%v", msg, err)
			}

			_, isOrPhan, err := blockMgr.ChainService.ProcessBlock(&newBlock)
			if err == nil {
				//return err
			}

			peer.MarkBlock(&newBlock)
			blockMgr.BroadcastBlock(types.MsgTypeBlock, &newBlock, false)

			if isOrPhan {
				//blockMgr.synchronise()
			}
		case types.MsgTypePeerState:
			var resp types.PeerState
			if err := msg.Decode(&resp); err != nil {
				return errors.Wrapf(ErrDecodeMsg, "PeerState msg:%v err:%v", msg, err)
			}
			go blockMgr.handlePeerState(peer, &resp)
		case types.MsgTypePeerStateReq:
			var req types.PeerStateReq
			if err := msg.Decode(&req); err != nil {
				return errors.Wrapf(ErrDecodeMsg, "PeerStateReq msg:%v err:%v", msg, err)
			}
			go blockMgr.handleReqPeerState(peer, &req)
		case types.MsgTypeHeaderReq:
			var req types.HeaderReq
			if err := msg.Decode(&req); err != nil {
				return errors.Wrapf(ErrDecodeMsg, "HeaderReq msg:%v err:%v", msg, err)
			}
			go blockMgr.handleHeaderReq(peer, &req)
		case types.MsgTypeHeaderRsp:
			var resp types.HeaderRsp
			if err := msg.Decode(&resp); err != nil {
				return errors.Wrapf(ErrDecodeMsg, "HeaderRsp msg:%v err:%v", msg, err)
			}
			go blockMgr.handleHeaderRsp(peer, &resp)
		}
	}

	return nil
}

func (blockMgr *BlockMgr) handleHeaderReq(peer types.PeerInfoInterface, req *types.HeaderReq) {
	headers := make([]types.BlockHeader, 0, req.ToHeight-req.FromHeight+1)
	for i := req.FromHeight; i <= req.ToHeight; i++ {
		node := blockMgr.ChainService.BestChain().NodeByHeight(uint64(i))
		if node != nil {
			headers = append(headers, node.Header())
		}
	}

	log.WithField("total header", len(headers)).WithField("from", req.FromHeight).WithField("to", req.ToHeight).Info("header req len")
	blockMgr.P2pServer.Send(peer.GetMsgRW(), uint64(types.MsgTypeHeaderRsp), types.HeaderRsp{Headers: headers})
}

func (blockMgr *BlockMgr) handleHeaderRsp(peer types.PeerInfoInterface, rsp *types.HeaderRsp) {
	peer.CalcAverageRtt()
	headerHashs := make([]*syncHeaderHash, 0, len(rsp.Headers))
	for _, h := range rsp.Headers {
		headerHashs = append(headerHashs, &syncHeaderHash{headerHash: h.Hash(), height: h.Height})
	}

	//The requested associated coroutine is closed
	err := blockMgr.checkHeaderChain(rsp.Headers)
	if err != nil {
		log.WithField("Reason", err).Info("checkHeaderChain fail")
		return
	}

	if len(headerHashs) >= 1 {
		log.WithField("total len:", len(headerHashs)).WithField("from height:", headerHashs[0].height).WithField("end height:", headerHashs[len(headerHashs)-1].height).Info("handleHeaderRsp ")
	} else {
		log.Error("handleHeaderRsp rsp nil")
		return
	}

	blockMgr.headerHashCh <- headerHashs
}

func (blockMgr *BlockMgr) HandleBlockReqMsg(peer types.PeerInfoInterface, req *types.BlockReq) {
	log.WithField("num:", len(req.BlockHashs)).Info("sync req block")
	zero := crypto.Hash{}
	startHeight := uint64(0)
	endHeight := blockMgr.ChainService.BestChain().Tip().Height

	if len(req.BlockHashs) == 0 {
		log.WithField("block hash num", len(req.BlockHashs)).Warn("handle block req")
		return
	}
	startHash := req.BlockHashs[0]
	endHash := req.BlockHashs[len(req.BlockHashs)-1]
	if startHash != zero && blockMgr.ChainService.BlockExists(&startHash) {
		startHeight = blockMgr.ChainService.Index().LookupNode(&startHash).Height
	}

	if endHash != zero {
		block, err := blockMgr.chainStore.GetBlock(&endHash)
		if err != nil {
			return
		}
		endHeight = block.Header.Height
	}

	blocks := []*types.Block{}
	count := 0
	for i := startHeight; i <= endHeight; i++ {
		if count%maxBlockCountReq == 0 && len(blocks) > 0 {
			blockMgr.P2pServer.Send(peer.GetMsgRW(), types.MsgTypeBlockResp, &types.BlockResp{
				Blocks: blocks,
			})
			blocks = []*types.Block{}
			count = 0
		}

		node := blockMgr.ChainService.BestChain().NodeByHeight(i)
		block, err := blockMgr.chainStore.GetBlock(node.Hash)
		if err != nil {
			return
		}
		count = count + 1
		blocks = append(blocks, block)
	}
	if len(blocks) > 0 {
		blockMgr.P2pServer.Send(peer.GetMsgRW(), types.MsgTypeBlockResp, &types.BlockResp{
			Blocks: blocks,
		})
		log.WithField("num", len(blocks)).Info("req blocks and rsp")
		//blocks = []*types.Block{}
	}
}

func (blockMgr *BlockMgr) HandleBlockRespMsg(peer types.PeerInfoInterface, rsp *types.BlockResp) {
	peer.CalcAverageRtt()
	blockMgr.blocksCh <- rsp.Blocks
}
