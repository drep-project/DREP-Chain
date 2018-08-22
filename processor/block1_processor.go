package processor

import (
    "fmt"
    "BlockChainTest/common"
    "BlockChainTest/pool"
    "BlockChainTest/util"
    "BlockChainTest/storage"
    "BlockChainTest/network"
    "bytes"
)

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

func (p *setup1Processor) process(msg interface{}) {
    if !storage.GetState().CheckState(storage.MINER, common.WAITING) {
        return
    }
    if setup1Msg, ok := msg.(common.SetUp1Message); ok {
        fmt.Println(setup1Msg)
        // TODO Check sig
        if !util.SliceEqual(storage.GetState().GetLeader().PubKey, setup1Msg.PubKey) {
            return
        }
        if setup1Msg.BlockHeight != storage.GetState().GetBlockHeight() + 1 {
            return
        }
        storage.GetState().SetBlock(setup1Msg.Block)
        storage.GetState().MoveToState(common.MSG_BLOCK1_RESPONSE)
        // TODO clear block1CommitProcessor and Start countdown
        // TODO Get Qi
        //q := crypto.GetQ()
        peer := storage.GetState().GetLeader().Peer
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
    if !storage.GetState().CheckState(storage.LEADER, common.MSG_BLOCK1_CHALLENGE) {
        return
    }
    if block1CommitMsg, ok := msg.(common.Block1CommitMessage); ok {
        fmt.Println(block1CommitMsg)
        miner := storage.GetState().GetMiner(block1CommitMsg.PubKey)
        if miner == nil {
            return
        }
        // TODO p.pubKey += pubKey
        // TODO p.q += q
        p.bitmap[miner.Address] = true
        p.count++
        if p.count == len(storage.GetState().GetMiners()) {
            storage.GetState().MoveToState(common.MSG_SETUP2)
            block := storage.GetState().GetBlock()
            // TODO calculate r
            miners := storage.GetState().GetMiners()
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
    if !storage.GetState().CheckState(storage.LEADER, common.MSG_BLOCK1_RESPONSE) {
        return
    }
    if block1ChallengeMsg, ok := msg.(common.Block1ChallengeMessage); ok {
        fmt.Println(block1ChallengeMsg)
        miner := storage.GetState().GetMiner(block1ChallengeMsg.PubKey)
        if miner == nil {
            return
        }
        // TODO p.pubKey += pubKey
        // TODO p.q += q
        p.bitmap[miner.Address] = true
        p.count++
        if p.count == len(storage.GetState().GetMiners()) {
            storage.GetState().MoveToState(common.MSG_BLOCK2_COMMIT)
            // TODO calculate r
            miners := storage.GetState().GetMiners()
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
    if !storage.GetState().CheckState(storage.MINER, common.MSG_BLOCK1_RESPONSE) {
        return
    }
    if block1ResponseMsg, ok := msg.(common.Block1ResponseMessage); ok {
        fmt.Println(block1ResponseMsg)
        miner := storage.GetState().GetMiner(block1ResponseMsg.PubKey)
        if miner == nil {
            return
        }
        // TODO calculate s
        // TODO send s to leader
    }
}