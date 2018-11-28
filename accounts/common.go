package accounts

import (
    "crypto/hmac"
    "crypto/sha512"
    "BlockChainTest/mycrypto"
    "net"
    "time"
)

func hmAC(message, key []byte) []byte {
    h := hmac.New(sha512.New, key)
    h.Write(message)
    return h.Sum(nil)
}

func genUnique() ([]byte, error) {
    interfaces, err := net.Interfaces()
    if err != nil {
        return nil, err
    }
    uni := ""
    for _, inter := range interfaces {
        mac := inter.HardwareAddr
        uni += mac.String()
    }
    uni += time.Now().String()
    return mycrypto.Hash256([]byte(uni)), nil
}

func genPrvKey(prv []byte) *mycrypto.PrivateKey {
    cur := mycrypto.GetCurve()
    pubKey := cur.ScalarBaseMultiply(prv)
    prvKey := &mycrypto.PrivateKey{Prv: prv, PubKey: pubKey}
    return prvKey
}

func NewNodeInDebug(prv []byte) (*Node, error) {
    uni, err := genUnique()
    if err != nil {
        return nil, err
    }
    h := hmAC(uni, DrepMark)
    prvKey := genPrvKey(prv)
    chainCode := h[KeyBitSize:]
    node := &Node{
        PrvKey: prvKey,
        ChainId: RootChainID,
        ChainCode: chainCode,
    }
    return node, nil
}

func NewAccountInDebug(prv []byte) (*Account, error) {
    node, _ := NewNodeInDebug(prv)
    err := store(node)
    if err != nil {
        return nil, err
    }
    address := PubKey2Address(node.PrvKey.PubKey)
    account := &Account{
        Address: address,
        Node: node,
        Storage: NewStorage(),
    }
    return account, nil
}