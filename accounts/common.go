package accounts

import (
    "crypto/hmac"
    "crypto/sha512"
    "BlockChainTest/mycrypto"
    "crypto/rand"
    "BlockChainTest/log"
)

func hmAC(message, key []byte) []byte {
    h := hmac.New(sha512.New, key)
    h.Write(message)
    return h.Sum(nil)
}

func genSeed() ([]byte, error) {
    seed := make([]byte, SeedSize)
    _, err := rand.Read(seed)
    if err != nil {
        log.Println("Error in genSeed().")
    }
    return seed, err
}

func genPrvKey(prv []byte) *mycrypto.PrivateKey {
    cur := mycrypto.GetCurve()
    pubKey := cur.ScalarBaseMultiply(prv)
    prvKey := &mycrypto.PrivateKey{Prv: prv, PubKey: pubKey}
    return prvKey
}

func NewNodeInDebug(prv []byte) (*Node, error) {
    seed, err := genSeed()
    if err != nil {
        return nil, err
    }
    h := hmAC(seed, SeedMark)
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
        Storage: NewStorage(nil),
    }
    return account, nil
}