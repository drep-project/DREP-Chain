package accounts

import (
    "crypto/hmac"
    "crypto/sha512"
    "BlockChainTest/mycrypto"
    "crypto/rand"
    "BlockChainTest/log"
    "math/big"
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

func NewAccountInDebug(prv []byte) (Account, error) {
    account := &MainAccount{
        Node: NewNode(prv, MainChainID),
        Storage: NewStorage(nil),
        ChainCode: new(big.Int).SetInt64(int64(1)).Bytes(),
        SubAccounts: make(map[ChainID] *SubAccount),
    }
    err := store(account.Node)
    if err != nil {
        return nil, err
    }
    return account, nil
}