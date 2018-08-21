package processor

import (
    "BlockChainTest/common"
    "BlockChainTest/storage"
    "BlockChainTest/util"
)

type setUpMessageProcessor struct {
    setup map[string]bool
}

func (p * setUpMessageProcessor) finishSetup() bool {
    for _, v := range storage.GetState().GetMiners() {
        if !p.setup[v.Address] {
            return false
        }
    }
    return true
}
func (p *setUpMessageProcessor) process(msg interface{}) {
    if setUpMessage, ok := msg.(common.SetUpMessage); ok {
        if !storage.GetState().ContainsMiner(setUpMessage.PubKey) {
            return
        }
        // TODO Verify
        p.setup[util.GetAddress(setUpMessage.PubKey)] = true
    }
}