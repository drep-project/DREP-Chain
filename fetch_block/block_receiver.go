package fetch_block

import (
    "sync"
    "BlockChainTest/bean"
    "BlockChainTest/network"
    "fmt"
    "strconv"
    "errors"
)

type BlockReceiver struct {
    blockWg        sync.WaitGroup
    blockLock      sync.Mutex
    maxHeight      int64
    expectedHeight int64
    blockMap       map[string] *bean.Block
    lastBlock      *bean.Block
    senderPeer     *network.Peer
}

func NewBlockReceiver(peer *network.Peer) *BlockReceiver {
    receiver := &BlockReceiver{}
    receiver.senderPeer = peer
    return receiver
}

func (receiver *BlockReceiver) InitLocalBlock() {
    receiver.maxHeight = -1
    receiver.expectedHeight = -1
    receiver.blockMap = make(map[string] *bean.Block, 0)
}

func (receiver *BlockReceiver) FetchBlocks() {
    receiver.blockWg = sync.WaitGroup{}
    receiver.blockWg.Add(1)
    req := &bean.BlockReq{Req: "block", MinHeight: receiver.maxHeight + 1}
    peers := make([]*network.Peer, 1)
    peers[0] = receiver.senderPeer
    fmt.Println("m.leader: ", peers[0])
    network.SendMessage(peers, req)
    receiver.blockWg.Wait()
}

func (receiver *BlockReceiver) ProcessBlockResp(resp *bean.BlockResp) error {
    if resp.Resp != "block" {
        receiver.blockWg.Done()
        return errors.New("invalid resp")
    }
    receiver.expectedHeight = resp.MaxHeight
    block := resp.NewBlock
    if receiver.ValidateBlock(block) {
        receiver.AddBlock(block)
        fmt.Println("m.maxHeight: ", receiver.maxHeight)
        fmt.Println("m.expectedHeight: ", receiver.expectedHeight)
        if receiver.maxHeight == receiver.expectedHeight {
            receiver.blockWg.Done()
        }
        return nil
    } else {
        defer receiver.blockWg.Done()
        return errors.New("invalid block")
    }
}

func (receiver *BlockReceiver) ValidateBlock(block *bean.Block) bool {
    return true
}

func (receiver *BlockReceiver) AddBlock(block *bean.Block) {
    receiver.blockLock = sync.Mutex{}
    receiver.blockLock.Lock()
    key := "block_" + strconv.Itoa(int(receiver.maxHeight + 1))
    receiver.blockMap[key] = block
    receiver.maxHeight ++
    receiver.lastBlock = block
    receiver.blockLock.Unlock()
}

func (receiver *BlockReceiver) PrintBlocks() {
    fmt.Println("max height: ", receiver.maxHeight)
    fmt.Println("blocks:")
    for key, value := range receiver.blockMap {
        fmt.Println("id: ", key)
        fmt.Println("blk: ", value)
    }
}

func (receiver *BlockReceiver) DoFetch() {
    receiver.FetchBlocks()
    receiver.PrintBlocks()
}
