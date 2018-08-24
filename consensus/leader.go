package processor

import (
    "BlockChainTest/common"
    "BlockChainTest/store"
    "BlockChainTest/node"
    "fmt"
    "bytes"
    "BlockChainTest/network"
    "sync"
)

const (
    waiting              = 0
    setUp               = 1 // LS ->CHA M received then commit
    waitingForCommit     = 3 // MS  MS->RESPONSE
    CHALLENGE            = 4 // LS ->SETUP2
    WAITING_FOR_RESPONSE = 5 // MS -> COMMIT
)
type Leader struct {
    msg interface{}
    miners []*node.Miner
    sigFunc func(interface{}, []byte) []byte
    wrapFunc func(interface{}, []byte, []byte) interface{}
    state int
    commitWg sync.WaitGroup
    commitBitmap map[common.Address]bool
    responseWg sync.WaitGroup
    responseBitmap map[common.Address]bool
    sigs map[common.Address][]byte
}

func NewLeader() *Leader {
    l := &Leader{}
    l.msg = msg
    miners := store.GetMiners()
    l.miners = make([]*node.Miner, len(miners) - 1)
    last := 0
    pubKey := store.GetPubKey()
    for _, miner := range miners {
        if !bytes.Equal(miner.PubKey, pubKey) {
            l.miners[last] = miner
            last++
        }
    }
    l.state = waiting
    return l
}

func (l *Leader) processConsensus(msg interface{},
                                  sigFunc func(interface{}, []byte) []byte,
                                  wrapFunc func(interface{}, []byte, []byte) interface{}) map[common.Address][]byte {
    priKey := store.GetPriKey()
    sig := sigFunc(msg, priKey)
    msg
    w := sync.WaitGroup{}
    w.Add()
}
func (l *Leader) setUp() {
    miner

}
//WAITING = 0
//MSG_SETUP1 = 1 // LS ->CHA MR WAITING received then commit
//MSG_BLOCK1_COMMIT = 3 // MS  ->RESPONSE LR CHAN
//MSG_BLOCK1_CHALLENGE = 4 // LS ->SETUP2 MR RESP
//MSG_BLOCK1_RESPONSE = 5 // MS -> COMMIT LR SETUP2
//
//MSG_SETUP2 = 6 // LS -> CHA MR COMMIT
//MSG_BLOCK2_COMMIT = 7 // MS -> RESP LR CHA
//MSG_BLOCK2_CHALLENGE = 8 // LS -> BLOCK MR RESP
//MSG_BLOCK2_RESPONSE = 9 // MS -> BLOCK LR BLOCK
//
//MSG_BLOCK = 9 // LS -> WAITING M -> WAITING

type setup1Processor struct {

}

func (p *setup1Processor) process(msg interface{})  {
    if !store.CheckState(node.MINER, common.WAITING) {
        return
    }
    if setup1Msg, ok := msg.(common.SetUp1Message); ok {
        fmt.Println(setup1Msg)
        // TODO Check sig
        if !bytes.Equal(store.GetLeader().PubKey, setup1Msg.PubKey) {
            return
        }
        if setup1Msg.BlockHeight != store.GetBlockHeight() + 1 {
            return
        }
        store.SetBlock(setup1Msg.Block)
        store.MoveToState(common.MSG_BLOCK1_RESPONSE)
        // TODO clear block1CommitProcessor and Start countdown
        // TODO Get Qi
        //q := crypto.GetQ()
        peer := store.GetLeader().Peer
        // TODO Send Qi to the leader
        // TODO Generate the block
        //network.SendMessage(peer, block1CommitMsg{q, pubKey})
    }
}

type block1CommitProcessor struct {
    bitmap map[common.Address]bool
    count int
    //pubKey
    //q
}

func (p *block1CommitProcessor) clear() {
    p.bitmap = make(map[common.Address]bool)
    p.count = 0
}

func (p *block1CommitProcessor) process(msg interface{}) {
    if !store.CheckState(node.LEADER, common.MSG_BLOCK1_CHALLENGE) {
        return
    }
    if block1CommitMsg, ok := msg.(common.Block1CommitMessage); ok {
        fmt.Println(block1CommitMsg)
        miner := store.GetMiner(block1CommitMsg.PubKey)
        if miner == nil {
            return
        }
        // TODO p.pubKey += pubKey
        // TODO p.q += q
        p.bitmap[miner.Address] = true
        p.count++
        if p.count == len(store.GetMiners()) {
            store.MoveToState(common.MSG_SETUP2)
            block := store.GetBlock()
            // TODO calculate r
            miners := store.GetMiners()
            // TODO delete itself from miners
            // TODO Send r, q, pk to miners  common.Block1ChallengeMessage
        }
    }
}

type block1ChallengeProcessor struct {
    bitmap map[common.Address]bool
    count int
    //pubKey
    //q
}

func (p *block1ChallengeProcessor) clear() {
    p.bitmap = make(map[common.Address]bool)
    p.count = 0
}

func (p *block1ChallengeProcessor) process(msg interface{}) {
    if !store.CheckState(node.LEADER, common.MSG_BLOCK1_RESPONSE) {
        return
    }
    if block1ChallengeMsg, ok := msg.(common.Block1ChallengeMessage); ok {
        fmt.Println(block1ChallengeMsg)
        miner := store.GetMiner(block1ChallengeMsg.PubKey)
        if miner == nil {
            return
        }
        // TODO p.pubKey += pubKey
        // TODO p.q += q
        p.bitmap[miner.Address] = true
        p.count++
        if p.count == len(store.GetMiners()) {
            store.MoveToState(common.MSG_BLOCK2_COMMIT)
            // TODO calculate r
            miners := store.GetMiners()
            // TODO delete itself from miners
            // TODO Send r, q, pk to miners
        }
    }
}

type block1ResponseProcessor struct {
    bitmap map[common.Address]bool
    count int
    //pubKey
    //q
}

func (p *block1ResponseProcessor) process(msg interface{}) {
    if !store.CheckState(node.MINER, common.MSG_BLOCK1_RESPONSE) {
        return
    }
    if block1ResponseMsg, ok := msg.(common.Block1ResponseMessage); ok {
        fmt.Println(block1ResponseMsg)
        miner := store.GetMiner(block1ResponseMsg.PubKey)
        if miner == nil {
            return
        }
        // TODO calculate s
        // TODO send s to leader
    }
}