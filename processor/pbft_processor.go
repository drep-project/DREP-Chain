package processor

import (
    "BlockChainTest/bean"
    "BlockChainTest/network"
    "BlockChainTest/crypto"
)

type AnnouncementProcessor struct {
    Ann *bean.Announcement
    K []byte
    Leader *network.Peer
}

func (p *AnnouncementProcessor) Process() error {
    testSig, err := crypto.Sign([]byte(p.Ann.Test))
    if err != nil {
        return err
    }
    k, q, err := crypto.GetRandomKQ()
    if err != nil {
        return nil
    }
    pubKey, err := crypto.GetPubKey()
    if err != nil {
        return nil
    }
    copy(p.K, k)
    commitment := &bean.Commitment{PubKey:pubKey, Q: q, TestSig: testSig}
    peers := make([]*network.Peer, 1)
    peers[0] = p.Leader
    network.SendMessage(peers, commitment)
    return nil
}

type CommitmentProcessor struct {

}
