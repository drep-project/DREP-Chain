package main

import (
    "math/big"
    "encoding/hex"
    "BlockChainTest/mycrypto"
    "crypto/hmac"
    "crypto/sha512"
    "net"
    "time"
    "errors"
    "C"
)

var (
    mark    = []byte("Drep Coin Seed")
    bitSize = 32
    hexSize = 64
)

func padding(b []byte) []byte {
    if len(b) < bitSize {
        zero := make([]byte, bitSize - len(b))
        b = append(zero, b...)
    }
    return b
}

func bytes2Hex(b []byte) string {
    return hex.EncodeToString(padding(b))
}

func hex2Bytes(s string) []byte {
    b, _ := hex.DecodeString(s)
    return padding(b)
}

func genPrivateKey() (prvKey string) {
    uni, _ := genUnique()
    h := hmAC(uni, mark)
    sk := genPrvKey(h[:bitSize])
    prvKey = bytes2Hex(sk.Prv)
    return
}

func NewMainAccountKey() (prvKey, pubKey, chainCode, address string) {
    uni, _ := genUnique()
    h := hmAC(uni, mark)
    sk := genPrvKey(h[:bitSize])
    cc := h[bitSize:]
    prvKey = bytes2Hex(sk.Prv)
    pubKey = bytes2Hex(sk.PubKey.X) + bytes2Hex(sk.PubKey.Y)
    chainCode = bytes2Hex(cc)
    address = PubKey2Address(sk.PubKey).Hex()
    return
}

func NewSubAccountKey(chainID, parentPrvKey, parentChainCode string) (prvKey, pubKey, address string) {
    pid := new(big.Int).SetBytes(hex2Bytes(parentChainCode))
    cid := new(big.Int).SetBytes(hex2Bytes(chainID))
    msg := new(big.Int).Xor(pid, cid).Bytes()
    h := hmAC(msg, hex2Bytes(parentPrvKey))
    sk := genPrvKey(h[:bitSize])
    prvKey = bytes2Hex(sk.Prv)
    pubKey = bytes2Hex(sk.PubKey.X) + bytes2Hex(sk.PubKey.Y)
    address = PubKey2Address(sk.PubKey).Hex()
    return
}

func Sign(prvKey, pubKey, msg string) (signature string) {
    sk := &mycrypto.PrivateKey{
        Prv: hex2Bytes(prvKey),
        PubKey: &mycrypto.Point{
            X: hex2Bytes(pubKey[:hexSize]),
            Y: hex2Bytes(pubKey[hexSize:]),
        },
    }
    b := hex2Bytes(msg)
    sig, _ := mycrypto.Sign(sk, b)
    signature = bytes2Hex(sig.R) + bytes2Hex(sig.S)
    return
}

func Verify(pubKey, msg, signature string) bool {
    pk := &mycrypto.Point{
        X: hex2Bytes(pubKey[:hexSize]),
        Y: hex2Bytes(pubKey[hexSize:]),
    }
    sig := &mycrypto.Signature{
        R: hex2Bytes(signature[:hexSize]),
        S: hex2Bytes(signature[hexSize:]),
    }
    b := hex2Bytes(msg)
    return mycrypto.Verify(sig, pk, b)
}

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

const (
    HashLength    = 32
    AddressLength = 20
    RootChainID   = 0
)

var (
    ErrExceedHashLength = errors.New("bytes length exceed maximum hash length of 32")
)

type CommonAddress [AddressLength]byte

func (addr CommonAddress) IsEmpty() bool {
    return addr == CommonAddress{}
}

func Bytes2Address(b []byte) CommonAddress {
    if b == nil {
        return CommonAddress{}
    }
    var addr CommonAddress
    addr.SetBytes(b)
    return addr
}

func (addr *CommonAddress) SetBytes(b []byte) {
    if len(b) > len(addr) {
        copy(addr[:], b[len(b) - AddressLength:])
    } else {
        copy(addr[AddressLength-len(b):], b)
    }
}

func (addr CommonAddress) Bytes() []byte {
    return addr[:]
}

func Hex2Address(s string) CommonAddress {
    if s == "" {
        return CommonAddress{}
    }
    b, err := hex.DecodeString(s)
    if err != nil {
        return CommonAddress{}
    }
    return Bytes2Address(b)
}

func (addr CommonAddress) Hex() string {
    return hex.EncodeToString(addr.Bytes())
}

func Big2Address(x *big.Int) CommonAddress {
    return Bytes2Address(x.Bytes())
}

func (addr CommonAddress) Big() *big.Int {
    return new(big.Int).SetBytes(addr.Bytes())
}

func PubKey2Address(pubKey *mycrypto.Point) CommonAddress {
    return Bytes2Address(mycrypto.Hash256(pubKey.Bytes()))
}

func main() {}