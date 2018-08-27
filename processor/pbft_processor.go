package processor

import (
    "BlockChainTest/bean"
    "BlockChainTest/network"
    "BlockChainTest/crypto"
    "errors"
    "math/big"
)

type SetUpProcessor struct {
    PrvKey *bean.PrivateKey
    K []byte
    Leader *network.Peer
}

func (p *SetUpProcessor) Process(msg interface{}) {
    if setUp, ok := msg.(*bean.Setup); ok {
        testSig, err := crypto.Sign([]byte(announcement.Test))
        if err != nil {
            return err
        }
        k, q, err := crypto.GetRandomKQ()
        if err != nil {
            return nil
        }
        pubKey := p.PrvKey.PubKey
        copy(p.K, k)
        commitment := &bean.Commitment{PubKey: pubKey, Q: q, TestSig: testSig}
        peers := make([]*network.Peer, 1)
        peers[0] = p.Leader
        network.SendMessage(peers, commitment)
    }
}

type CommitmentProcessor struct {
    Test string
    MinorCommitmentMap map[string] *bean.Commitment
}

func (p *CommitmentProcessor) Process(msg interface{}) error {
    commitment, ok := msg.(*bean.Commitment)
    if !ok {
        return errors.New("wrong message type: not commitment")
    }
    if !crypto.Verify(commitment.TestSig, commitment.PubKey, []byte(p.Test)) {
        return errors.New("invalid test")
    }
    addr := commitment.PubKey.Addr()
    p.MinorCommitmentMap[addr] = commitment
    return nil
}

type ChallengeProcessor struct {
    PrvKey *bean.PrivateKey
    K []byte
    Leader *network.Peer
}

func (p *ChallengeProcessor) Process(msg interface{}) error {
    challenge, ok := msg.(*bean.Challenge)
    if !ok {
        return errors.New("wrong message type: not challenge")
    }
    curve := crypto.GetCurve()
    prvKey := p.PrvKey
    r := crypto.ConcatHash256(challenge.GroupQ.Bytes(), challenge.GroupPubKey.Bytes(), challenge.Object)
    r0 := new(big.Int).SetBytes(challenge.R)
    r1 := new(big.Int).SetBytes(r)
    if r0.Cmp(r1) != 0 {
        return errors.New("wrong hash value")
    }
    k := new(big.Int).SetBytes(p.K)
    prvInt := new(big.Int).SetBytes(prvKey.Prv)
    s := new(big.Int).Mul(r1, prvInt)
    s.Sub(k, s)
    s.Mod(s, curve.N)
    response := &bean.Response{PubKey: prvKey.PubKey, S: s.Bytes()}
    peers := make([]*network.Peer, 1)
    peers[0] = p.Leader
    network.SendMessage(peers, response)
    return nil
}

type ResponseProcessor struct {
    MinorResponseMap map[string] *bean.Response
}

func (p *ResponseProcessor) Process(msg interface{}) error {
    response, ok := msg.(*bean.Response)
    if !ok {
        return errors.New("wrong message type: not response")
    }
    addr := response.PubKey.Addr()
    p.MinorResponseMap[addr] = response
    return nil
}