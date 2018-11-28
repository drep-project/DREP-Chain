package accounts

import (
    "testing"
    "fmt"
    "encoding/hex"
)

func TestNewRootAccount(t *testing.T) {
    var parent *Node
    var chainId int64 = 0
    account, err := NewNormalAccount(parent, chainId)
    fmt.Println("err: ", err)
    fmt.Println("account: ", account)
}

func TestNewChildAccount(t *testing.T) {
    var parent *Node
    var chainId int64 = 0
    root, err := NewNormalAccount(parent, chainId)
    fmt.Println("root err: ", err)
    fmt.Println("root: ")
    fmt.Println("address: ", root.Address.Hex())
    fmt.Println("chainId: ", root.Node.ChainId)
    fmt.Println("balance: ", root.Storage.Balance)
    fmt.Println("nonce: ", root.Storage.Nonce)
    fmt.Println("byteCode: ", root.Storage.ByteCode)
    fmt.Println("codeHash: ", root.Storage.CodeHash)
    fmt.Println()
    var cid int64 = 567283
    child, err := NewNormalAccount(root.Node, cid)
    fmt.Println("child err: ", err)
    fmt.Println("child: ", child)
    fmt.Println("address: ", child.Address.Hex())
    fmt.Println("chainId: ", child.Node.ChainId)
    fmt.Println("balance: ", child.Storage.Balance)
    fmt.Println("nonce: ", child.Storage.Nonce)
    fmt.Println("byteCode: ", child.Storage.ByteCode)
    fmt.Println("codeHash: ", child.Storage.CodeHash)
}

func TestLoad(t *testing.T) {
    keyAddr := "82f01665d7d0d5e9e4aad84a9121d8bd228a1120"
    node, err := OpenKeystore(keyAddr)
    fmt.Println("err: ", err)
    fmt.Println("node: ", node)
    fmt.Println("prvkey: ", hex.EncodeToString(node.PrvKey.Prv))
    fmt.Println("chainCode: ", node.ChainCode)
}