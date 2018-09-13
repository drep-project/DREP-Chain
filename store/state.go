package store

import (
    "sync"
    "BlockChainTest/network"
    "BlockChainTest/bean"
    "BlockChainTest/consensus"
    "BlockChainTest/mycrypto"
    "BlockChainTest/log"
    "math/big"
)

const maxMinerNumber = 2

var (
    role = bean.OTHER
    leader *consensus.Leader
    member *consensus.Member

    //miningState int
    miners []*network.Peer
    //minerIndex map[bean.Address]int

    blockHeight int
    lock sync.Locker
    prvKey *mycrypto.PrivateKey
    pubKey *mycrypto.Point
    address bean.Address
    remainingSetUp *bean.Setup

    currentMinerIndex int
    myIndex = 0
)

func init()  {
    lock = &sync.Mutex{}
    currentMinerIndex = -1
    //prvKey, _ = mycrypto.GetPrivateKey()
    //pubKey = GetPubKey()
    curve := mycrypto.GetCurve()
    k0 := []byte{0x22, 0x11}
    k1 := []byte{0x14, 0x44}
    k2 := []byte{0x11, 0x55}
    //k3 := []byte{0x12, 0x55}
    pub0 := curve.ScalarBaseMultiply(k0)
    pub1 := curve.ScalarBaseMultiply(k1)
    pub2 := curve.ScalarBaseMultiply(k2)
    balances[bean.Addr(pub0)] = big.NewInt(10000)
    balances[bean.Addr(pub1)] = big.NewInt(10000)
    balances[bean.Addr(pub2)] = big.NewInt(10000)
    //pub3 := curve.ScalarBaseMultiply(k3)
    prv0 := &mycrypto.PrivateKey{Prv: k0, PubKey: pub0}
    prv1 := &mycrypto.PrivateKey{Prv: k1, PubKey: pub1}
    prv2 := &mycrypto.PrivateKey{Prv: k2, PubKey: pub2}
    //prv3 := &mycrypto.PrivateKey{Prv: k3, PubKey: pub3}
    ip0 := network.IP("192.168.3.147")
    ip1 := network.IP("192.168.3.43")
    ip2 := network.IP("192.168.3.113")
    //ip3 := network.IP("192.168.3.79")
    port0 := network.Port(55555)
    port1 := network.Port(55555)
    port2 := network.Port(55555)
    //port3 := network.Port(55555)
    peer0 := &network.Peer{IP: ip0, Port: port0, PubKey: pub0}
    peer1 := &network.Peer{IP: ip1, Port: port1, PubKey: pub1}
    peer2 := &network.Peer{IP: ip2, Port: port2, PubKey: pub2}
    //peer3 := &network.Peer{IP: ip3, Port: port3, PubKey: pub3}
    AddPeer(peer0)
    AddPeer(peer1)
    AddPeer(peer2)
    miners = []*network.Peer{peer0, peer1, peer2}

    switch myIndex {
    case 0:
        pubKey = pub0
        prvKey = prv0
        address = bean.Addr(pub0)
        //leader = consensus.NewLeader(pub0, peers)
        //member = nil
    case 1:
        pubKey = pub1
        prvKey = prv1
        address = bean.Addr(pub1)
        //leader = nil
        //member = consensus.NewMember(peer0, prvKey)
    case 2:
       pubKey = pub2
       prvKey = prv2
       address = bean.Addr(pub2)
        //leader = nil
        //member = consensus.NewMember(peer0, prvKey)
    }
}

func MoveToNextMiner() bool {
    lock.Lock()
    currentMinerIndex = (currentMinerIndex + 1) % len(miners)
    log.Println("Current miner index:", currentMinerIndex, " my index: ", myIndex)
    r := currentMinerIndex == myIndex
    leader = nil
    member = nil
    lock.Unlock()
    return r
}

func SetLeader(l *consensus.Leader) {
    leader = l
}

func SetMember(m *consensus.Member) {
    member = m
}

func GetRole() int {
    return role
}

func GetMiners() []*network.Peer {
    return miners
}

func AddMiner(miner *network.Peer) {
    if miner != nil {
        if len(miners) < maxMinerNumber {
            miners = append(miners, miner)
        } else {
            miners = append(miners[1:], miner)
        }
    }
    for _, m := range miners {
        if m.PubKey.Equal(pubKey) {
            role = bean.MINER
            return
        }
    }
    role = bean.OTHER
}

func ContainsMiner(pubKey *mycrypto.Point) bool {
    for _, v:= range miners {
        if v.PubKey.Equal(pubKey) {
            return true
        }
    }
    return false
}

func GetMiner(pubKey *mycrypto.Point) *network.Peer {
    for _, v:= range miners {
        if v.PubKey.Equal(pubKey) {
            return v
        }
    }
    return nil
}


func GetBlockHeight() int {
    return blockHeight
}

func GetLeader() *network.Peer {
    return miners[currentMinerIndex]
}

func GenerateBlock() *bean.Block {
    height := currentBlockHeight + 1
    currentBlockHeight = height
    ts := PickTransactions(BlockGasLimit)
    return &bean.Block{Header: &bean.BlockHeader{Height: height},Data:&bean.BlockData{TxCount:int32(len(ts)), TxList:ts}}
}

func GetPubKey() *mycrypto.Point {
    return pubKey
}

func GetAddress() bean.Address {
    return address
}

func GetPrvKey() *mycrypto.PrivateKey {
    return prvKey
}

func GetItSelfOnLeader() *consensus.Leader {
    return leader
}

func GetItSelfOnMember() *consensus.Member {
    return member
}

func SetRemainingSetup(setup *bean.Setup)  {
    remainingSetUp = setup
}

func GetRemainingSetup() *bean.Setup {
    return remainingSetUp
}