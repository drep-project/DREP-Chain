package fetch_block

import (
    "sync"
    "BlockChainTest/bean"
    "fmt"
    "BlockChainTest/network"
    "strconv"
    "time"
    "errors"
)

type BlockSender struct {
    blockWg      sync.WaitGroup
    maxHeight    int64
    blockMap     map[string] *bean.Block
    receiverPeer *network.Peer
}

func NewBlockSender(peer *network.Peer) *BlockSender {
    sender := &BlockSender{}
    sender.receiverPeer = peer
    return sender
}

func (sender *BlockSender) InitLocalBlock() {
    sender.blockWg = sync.WaitGroup{}
    sender.blockWg.Add(1)
    sender.maxHeight = 4
    sender.blockMap = make(map[string] *bean.Block, sender.maxHeight + 1)
    var i int64
    for i = 0; i <= sender.maxHeight; i++ {
        key := "block_" + strconv.Itoa(int(i))
        value := &bean.Block{Header: &bean.BlockHeader{Timestamp: time.Now().Unix(), Height: i}}
        sender.blockMap[key] = value
    }
}

func (sender *BlockSender) GetBlock(height int64) *bean.Block {
    key := "block_" + strconv.Itoa(int(height))
    return sender.blockMap[key]
}

func (sender *BlockSender) ProcessBlockReq(req *bean.BlockReq) error {
    if req.Req != "block" {
        return errors.New("invalid req")
    }
    minHeight := req.MinHeight
    maxHeight := sender.maxHeight
    fmt.Println()
    fmt.Println("minHeight: ", minHeight)
    fmt.Println("maxHeight: ", maxHeight)
    fmt.Println()
    peers := make([]*network.Peer, 1)
    peers[0] = sender.receiverPeer
    for height := minHeight; height <= maxHeight; height ++ {
        block := sender.GetBlock(height)
        if block == nil {
            fmt.Println("no such block exists")
            break
        }
        resp := &bean.BlockResp{Resp: "block", MaxHeight: maxHeight, NewBlock: block}
        network.SendMessage(peers, resp)
    }
    //server.RespWg.Done()
    return nil
}

func (sender *BlockSender) WaitForBlockReq() {
    sender.blockWg = sync.WaitGroup{}
    sender.blockWg.Add(1)
    sender.blockWg.Wait()
}
