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
)

func padding(b []byte) []byte {
    if len(b) < bitSize {
        zero := make([]byte, bitSize - len(b))
        b = append(zero, b...)
    }
    return b
}

func bytes2Hex(b []byte) string {
    //return "1234"
    var key string = string( hex.EncodeToString(padding(b)))
    return key
}

func hex2Bytes(s string) []byte {
    b, _ := hex.DecodeString(s)
    return padding(b)
}

//export NewRootChainAccount
func NewRootChainAccount() (*C.char, *C.char, *C.char, *C.char) {
    uni, _ := genUnique()
    h := hmAC(uni, mark)
    sk := genPrvKey(h[:bitSize])
    cc := h[bitSize:]
    prvKey := make([]byte, bitSize)
    copy(prvKey, padding(sk.Prv))
    pubKey := make([]byte, 2 * bitSize)
    copy(pubKey[:bitSize], padding(sk.PubKey.X))
    copy(pubKey[bitSize:], padding(sk.PubKey.Y))
    chainCode := make([]byte, bitSize)
    copy(chainCode, padding(cc))
    address := PubKey2Address(sk.PubKey).Hex()
    return C.CString(bytes2Hex(prvKey)), C.CString(bytes2Hex(pubKey)), C.CString(bytes2Hex(chainCode)), C.CString(address)
}

//export NewSubChainAccount
func NewSubChainAccount(chainId, parentPrvKey, parentChainCode string) (*C.char, *C.char, *C.char, *C.char) {
    pid := new(big.Int).SetBytes(hex2Bytes(parentChainCode))
    cid := new(big.Int).SetBytes(hex2Bytes(chainId))
    msg := new(big.Int).Xor(pid, cid).Bytes()
    h := hmAC(msg, hex2Bytes(parentPrvKey))
    sk := genPrvKey(h[:bitSize])
    cc := h[bitSize:]
    prvKey := make([]byte, bitSize)
    copy(prvKey, padding(sk.Prv))
    pubKey := make([]byte, 2 * bitSize)
    copy(pubKey[:bitSize], padding(sk.PubKey.X))
    copy(pubKey[bitSize:], padding(sk.PubKey.Y))
    chainCode := make([]byte, bitSize)
    copy(chainCode, padding(cc))
    address := PubKey2Address(sk.PubKey).Hex()
    return C.CString(bytes2Hex(prvKey)), C.CString(bytes2Hex(pubKey)), C.CString(bytes2Hex(chainCode)), C.CString(address)
}

//export ImportKeystore
func ImportKeystore(prvKey string) (*C.char, *C.char) {
    curve := mycrypto.GetCurve()
    prv := hex2Bytes(prvKey)
    pub := curve.ScalarBaseMultiply(prv)
    pubKey := make([]byte, 2 * bitSize)
    copy(pubKey[:bitSize], padding(pub.X))
    copy(pubKey[bitSize:], padding(pub.Y))
    addr := PubKey2Address(pub).Hex()
    return C.CString(bytes2Hex(pubKey)), C.CString(addr)
}

//export Sign
func Sign(prvKey, pubKey, msg string) *C.char {
    prv := hex2Bytes(prvKey)
    pub := hex2Bytes(pubKey)
    sk := &mycrypto.PrivateKey{
        Prv: prv,
        PubKey: &mycrypto.Point{
            X: pub[:bitSize],
            Y: pub[bitSize:],
        },
    }
    sig, _ := mycrypto.Sign(sk, hex2Bytes(msg))
    signature := make([]byte, 2 * bitSize)
    copy(signature[:bitSize], padding(sig.R))
    copy(signature[bitSize:], padding(sig.S))
    return C.CString(bytes2Hex(signature))
}

//export Verify
func Verify(pubKey, msg, signature string) bool {
    pub := hex2Bytes(pubKey)
    sign := hex2Bytes(signature)
    pk := &mycrypto.Point{
        X: pub[:bitSize],
        Y: pub[bitSize:],
    }
    sig := &mycrypto.Signature{
        R: sign[:bitSize],
        S: sign[bitSize:],
    }
    return mycrypto.Verify(sig, pk, hex2Bytes(msg))
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