package store

import (
    "sync"
    "BlockChainTest/network"
    "BlockChainTest/bean"
    "BlockChainTest/consensus"
    "BlockChainTest/crypto"
    "BlockChainTest/log"
    "math/big"
    role2 "BlockChainTest/role"
)

var (
    role int
    leader *consensus.Leader
    member *consensus.Member

    user   *role2.User
    newcomer *role2.Newcomer

    //miningState int
    miners []*network.Peer
    //minerIndex map[bean.Address]int
    peers []*network.Peer
    blockHeight int
    block *bean.Block
    lock sync.Locker
    prvKey *crypto.PrivateKey
    pubKey *crypto.Point
    address bean.Address
    remainingSetUp *bean.Setup

    currentMinerIndex int
    myIndex = 2
)

func init()  {
    lock = &sync.Mutex{}
    currentMinerIndex = -1
    //prvKey, _ = crypto.GetPrivateKey()
    //pubKey = GetPubKey()
    curve := crypto.GetCurve()
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
    prv0 := &crypto.PrivateKey{Prv: k0, PubKey: pub0}
    prv1 := &crypto.PrivateKey{Prv: k1, PubKey: pub1}
    prv2 := &crypto.PrivateKey{Prv: k2, PubKey: pub2}
    //prv3 := &crypto.PrivateKey{Prv: k3, PubKey: pub3}
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
    miners = []*network.Peer{peer0, peer1, peer2}
    peers = []*network.Peer{peer0, peer1, peer2}//, peer3}
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

func ChangeRole() {
    lock.Lock()
    miners = GetMiners()
    currentMinerIndex = (currentMinerIndex + 1) % len(miners)
    log.Println("Current miner index:", currentMinerIndex, " my index: ", myIndex)
    if currentMinerIndex == myIndex {
        role = bean.LEADER
    } else {
        role = bean.MEMBER
    }
    leader = nil
    member = nil
    lock.Unlock()
}

func NewcomerRole()  {
    switch myIndex {
    case 2:
        role = bean.NEWCOMER
    default:
        role = bean.OTHER
    }
    leader = nil
    member = nil
}

func SetLeader(l *consensus.Leader) {
    leader = l
}

func SetMember(m *consensus.Member) {
    member = m
}

func SetUser(u *role2.User) {
    user = u
}

func SetNewComer(n *role2.Newcomer)  {
    newcomer = n
}

func GetRole() int {
    return role
}

func GetMiners() []*network.Peer {
    return miners
}

func ContainsMiner(pubKey *crypto.Point) bool {
    for _, v:= range miners {
        if v.PubKey.Equal(pubKey) {
            return true
        }
    }
    return false
}

func GetMiner(pubKey *crypto.Point) *network.Peer {
    for _, v:= range miners {
        if v.PubKey.Equal(pubKey) {
            return v
        }
    }
    return nil
}

func GetPeer(pubKey *crypto.Point) *network.Peer {
    for _, v:= range peers {
        if v.PubKey.Equal(pubKey) {
            return v
        }
    }
    return nil
}

func GetPeers() []*network.Peer {
    result := make([]*network.Peer, 0)
    for _, v := range peers {
        if !v.PubKey.Equal(pubKey) {
            result = append(result, v)
        }
    }
    return result
}

func GetBlockHeight() int {
    return blockHeight
}

func CheckRole(r int) bool {
    return role == r
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

func SetBlock(b *bean.Block) {
    block = b
}

func GetPubKey() *crypto.Point {
    return pubKey
}

func GetAddress() bean.Address {
    return address
}

func GetPrvKey() *crypto.PrivateKey {
    return prvKey
}

func GetItSelfOnLeader() *consensus.Leader {
    return leader
}

func GetItSelfOnMember() *consensus.Member {
    return member
}

func GetItSelfOnUser() *role2.User {
    return user
}

func GetItSelfOnNewcomer() *role2.Newcomer {
    return newcomer
}

func SetRemainingSetup(setup *bean.Setup)  {
    remainingSetUp = setup
}

func GetRemainingSetup() *bean.Setup {
    return remainingSetUp
}