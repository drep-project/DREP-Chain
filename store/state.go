package store

import (
    "sync"
    "BlockChainTest/network"
    "BlockChainTest/bean"
    "BlockChainTest/mycrypto"
    "math/big"
)

var (

    minerNum = 3
    lock     sync.Locker
    prvKey   *mycrypto.PrivateKey
    pubKey   *mycrypto.Point
    address  bean.Address

    port network.Port

    myIndex = 0
)

func init()  {
    lock = &sync.Mutex{}
    curMiner = -1
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
    var ip0, ip1, ip2 network.IP
    var port0, port1, port2 network.Port
    if LocalTest {
        ip0 = network.IP("127.0.0.1")
        ip1 = network.IP("127.0.0.1")
        port0 = network.Port(55555)
        port1 = network.Port(55556)
        port2 = network.Port(55557)
    } else {
        ip0 = network.IP("192.168.3.231")
        ip1 = network.IP("192.168.3.197")
        ip2 = network.IP("192.168.3.43")
        port0 = network.Port(55555)
        port1 = network.Port(55555)
        port2 = network.Port(55555)
    }
    //port2 := network.Port(55555)
    //port3 := network.Port(55555)
    peer0 := &network.Peer{IP: ip0, Port: port0, PubKey: pub0}
    peer1 := &network.Peer{IP: ip1, Port: port1, PubKey: pub1}
    peer2 := &network.Peer{IP: ip2, Port: port2, PubKey: pub2}
    //peer3 := &network.Peer{IP: ip3, Port: port3, PubKey: pub3}
    AddPeer(peer0)
    AddPeer(peer1)
    if minerNum > 2 {
        AddPeer(peer2)
        curMiners = []*network.Peer{peer0, peer1, peer2}
        miners = []*network.Peer{peer0, peer1, peer2}
    } else {
        curMiners = []*network.Peer{peer0, peer1} //, peer2}
        miners = []*network.Peer{peer0, peer1}
    }
    minerIndex = minerNum - 1
    switch myIndex {
    case 0:
        pubKey = pub0
        prvKey = prv0
        address = bean.Addr(pub0)
        adminPubKey = pub0
        port = port0
        //leader = consensus.NewLeader(pub0, peers)
        //member = nil
    case 1:
        pubKey = pub1
        prvKey = prv1
        address = bean.Addr(pub1)
        port = port1
        //leader = nil
        //member = consensus.NewMember(peer0, prvKey)
    case 2:
        pubKey = pub2
        prvKey = prv2
        address = bean.Addr(pub2)
        port = port2
        //leader = nil
        //member = consensus.NewMember(peer0, prvKey)
    }
    IsStart = myIndex < minerNum
}

func GenerateBlock() *bean.Block {
    height := GetCurrentBlockHeight() + 1
    //currentBlockHeight = height
    ts := PickTransactions(BlockGasLimit)
    return &bean.Block{Header: &bean.BlockHeader{Height: height, LeaderPubKey:GetPubKey()},Data:&bean.BlockData{TxCount:int32(len(ts)), TxList:ts}}
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

func GetPort() network.Port {
    return port
}