package store

import (
    "sync"
    "BlockChainTest/network"
    "BlockChainTest/bean"
    "BlockChainTest/mycrypto"
    "math/big"
    "BlockChainTest/database"
    "encoding/json"
    "errors"
    "time"
    "BlockChainTest/trie"
    "BlockChainTest/accounts"
)

var (
    minerNum = 3
    lock     sync.Locker
    chainId  accounts.ChainID
    prvKey   *mycrypto.PrivateKey
    pubKey   *mycrypto.Point
    address  bean.Address

    port network.Port

    myIndex = 0
    nodes map[string] *accounts.Node
)

func init()  {
    lock = &sync.Mutex{}
    curMiner = -1
    //prvKey, _ = mycrypto.GetPrivateKey()
    //pubKey = GetPubKey()
    curve := mycrypto.GetCurve()
    var id0 int64 = 0
    var id1 int64 = 0
    var id2 int64 = 0
    k0 := []byte{0x22, 0x11}
    k1 := []byte{0x14, 0x44}
    k2 := []byte{0x11, 0x55}
    //k3 := []byte{0x12, 0x55}
    pub0 := curve.ScalarBaseMultiply(k0)
    pub1 := curve.ScalarBaseMultiply(k1)
    pub2 := curve.ScalarBaseMultiply(k2)
    database.PutBalance(accounts.Hex2Address(bean.Addr(pub0).String()), big.NewInt(10000))
    database.PutBalance(accounts.Hex2Address(bean.Addr(pub1).String()), big.NewInt(10000))
    database.PutBalance(accounts.Hex2Address(bean.Addr(pub2).String()), big.NewInt(10000))
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
        ip2 = network.IP("192.168.3.236")
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
        chainId = accounts.ChainID(id0)
        pubKey = pub0
        prvKey = prv0
        address = bean.Addr(pub0)
        adminPubKey = pub0
        port = port0
        //leader = consensus.NewLeader(pub0, peers)
        //member = nil
    case 1:
        chainId = accounts.ChainID(id1)
        pubKey = pub1
        prvKey = prv1
        address = bean.Addr(pub1)
        port = port1
        //leader = nil
        //member = consensus.NewMember(peer0, prvKey)
    case 2:
        chainId = accounts.ChainID(id2)
        pubKey = pub2
        prvKey = prv2
        address = bean.Addr(pub2)
        port = port2
        //leader = nil
        //member = consensus.NewMember(peer0, prvKey)
    }
    acc, _ := accounts.NewAccountInDebug(prvKey.Prv)
    database.PutAccount(acc)
    database.AddNode(acc.GetNode())
    nodes = database.GetNodes()

    IsStart = myIndex < minerNum
}

func GenerateBlock() (*bean.Block, error) {
    maxHeight := database.GetMaxHeight()
    height := maxHeight + 1
    ts := PickTransactions(BlockGasLimit)
    previousBlock := database.GetHighestBlock()
    var b, previousHash []byte
    var err error
    if previousBlock != nil {
        b, err = json.Marshal(previousBlock.Header)
        if err != nil {
            return nil, err
        }
        previousHash = mycrypto.Hash256(b)
    } else {
        previousHash = []byte{}
    }
    gasLimit := new(big.Int).SetInt64(int64(10000000)).Bytes()
    gasUsed := GetGasSum(ts).Bytes()
    if ExceedGasLimit(gasUsed, gasLimit) {
        return nil, errors.New("gas used exceeds gas limit")
    }
    timestamp := time.Now().Unix()
    stateRoot := GetStateRoot(ts)
    txHashes, err := GetTxHashes(ts)
    if err != nil {
        return nil, err
    }
    merkle := trie.NewMerkle(txHashes)
    merkleRoot := merkle.Root.Hash
    return &bean.Block{
        Header: &bean.BlockHeader{
            Version: Version,
            PreviousHash: previousHash,
            GasLimit: gasLimit,
            GasUsed: gasUsed,
            Timestamp: timestamp,
            StateRoot: stateRoot,
            MerkleRoot: merkleRoot,
            TxHashes: txHashes,
            Height: height,
            LeaderPubKey:GetPubKey(),
        },
        Data:&bean.BlockData{
            TxCount:int32(len(ts)),
            TxList:ts,
        },
    }, nil
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

func CreateAccount(addr string, id int64) (string, error) {
    isMain := id == int64(accounts.RootChainID)
    var (
        acc accounts.Account
        err error
    )
    if isMain {
        acc, err = accounts.NewMainAccount(nil)
        if err != nil {
            return "", err
        }
    } else {
        m := database.GetAccount(accounts.Hex2Address(addr))
        if acc == nil {
            return "", errors.New("main account: " + addr + " not found")
        }
        if _, ok := m.(*accounts.MainAccount); !ok {
            return "", errors.New(addr + " is not a main account address")
        }
        acc, err = accounts.NewSubAccount(m.(*accounts.MainAccount), accounts.ChainID(id), nil)
        if err != nil {
            return "", nil
        }
    }
    err = database.PutAccount(acc)
    if err != nil {
        return "", err
    }
    database.AddNode(acc.GetNode())
    return acc.GetAddress().Hex(), nil
}

func SwitchAccount(addr string) error {
    if node, ok := nodes[addr]; ok {
        chainId = node.ChainId
        prvKey = node.PrvKey
        pubKey = node.PrvKey.PubKey
        return nil
    } else {
        return errors.New("fail to switch accounts: " + addr + " not found")
    }
}

func CurrentAccount() string {
    return bean.PubKey2Address(GetPubKey()).Hex()
}

func GetAccounts() []string {
    acc := make([]string, len(nodes))
    i := 0
    for addr, _ := range nodes {
        acc[i] = addr
        i++
    }
    return acc
}

func GetPort() network.Port {
    return port
}

func GetGasSum(ts []*bean.Transaction) *big.Int {
    gasSum := new(big.Int)
    for _, tx := range ts {
        gasSum = gasSum.Add(gasSum, tx.GetGas())
    }
    return gasSum
}

func ExceedGasLimit(used, limit []byte) bool {
    if new(big.Int).SetBytes(used).Cmp(new(big.Int).SetBytes(limit)) > 0 {
        return true
    }
    return false
}

func GetStateRoot(ts []*bean.Transaction) []byte {
    //for _, tx := range ts {
    //    from := bean.PubKey2Address(tx.Data.PubKey)
    //    to := bean.Hex2Address(tx.Data.To)
    //    gasUsed := tx.GetGas()
    //    nonce := tx.Data.Nonce
    //    amount := new(big.Int).SetBytes(tx.Data.Amount)
    //    prevSenderBalance := database.GetBalance(from)
    //    prevReceiverBalance := database.GetBalance(to)
    //    newSenderBalance := new(big.Int).Sub(prevSenderBalance, amount)
    //    newSenderBalance = newSenderBalance.Sub(newSenderBalance, gasUsed)
    //    newReceiverBalance := new(big.Int).Add(prevReceiverBalance, amount)
    //    database.PutBalance(from, newSenderBalance)
    //    database.PutBalance(to, newReceiverBalance)
    //    database.PutNonce(from, nonce)
    //}
    return database.GetStateRoot()
}

func GetTxHashes(ts []*bean.Transaction) ([][]byte, error) {
    txHashes := make([][]byte, len(ts))
    for i, tx := range ts {
        b, err := json.Marshal(tx.Data)
        if err != nil {
            return nil, err
        }
        txHashes[i] = mycrypto.Hash256(b)
    }
    return txHashes, nil
}
