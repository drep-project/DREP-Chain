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