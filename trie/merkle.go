package trie

import (
    "BlockChainTest/crypto"
    "math"
    "bytes"
    "fmt"
)

type MerkleNode struct {
    Parent     *MerkleNode
    LeftChild  *MerkleNode
    RightChild *MerkleNode
    Neighbour  *MerkleNode
    Hash       []byte
    Subscript  int
}

type MerkleLayer []*MerkleNode

type Merkle struct {
    Root   *MerkleNode
    //Trie   []MerkleLayer
    Leaves MerkleLayer
    Height int
}

func NewMerkle(hashes [][]byte) *Merkle {
    merkle := &Merkle{}
    height := getHeight(len(hashes))
    merkle.Height = height
    //merkle.Trie = make([]MerkleLayer, height)
    leaves := getLeaves(hashes)
    merkle.Leaves = leaves
    //merkle.Trie[0] = leaves
    layer := leaves
    fmt.Println("1: ", len(layer))
    for i := 0; i < height- 1; i++ {
        //merkle.Trie[i + 1] = getUpperLayer(merkle.Trie[i])
        layer = getUpperLayer(layer)
        fmt.Println(i + 2, ": ", len(layer))
    }
    //merkle.Root = merkle.Trie[height- 1][0]
    merkle.Root = layer[0]
    return merkle
}

func getHeight(n int) int {
    return int(math.Ceil(math.Log(float64(n)) / math.Log(2))) + 1
}

func getLeaves(hashes [][]byte) MerkleLayer {
    l := len(hashes)
    leaves := make([]*MerkleNode, l)
    for i := 0; i < l; i++ {
        node := &MerkleNode{}
        node.Hash = hashes[i]
        node.Subscript = i
        leaves[i] = node
    }
    return leaves
}

func getUpperLayer(layer MerkleLayer) MerkleLayer {
    l := len(layer)
    j := (l + 1) / 2
    upperLayer := make([]*MerkleNode, j)
    for i := 0; i < j; i++ {
        node := &MerkleNode{}
        lc := layer[2 * i]
        var rc *MerkleNode
        if 2 * i + 1 < l {
            rc = layer[2 * i + 1]
        }
        lc.Parent = node
        if rc != nil {
            lc.Neighbour = rc
            rc.Neighbour = lc
            rc.Parent = node
            node.Hash = crypto.ConcatHash256(lc.Hash, rc.Hash)
        } else {
            node.Hash = crypto.Hash256(lc.Hash)
        }
        node.LeftChild = lc
        node.RightChild = rc
        node.Subscript = i
        upperLayer[i] = node
    }
    return MerkleLayer(upperLayer)
}

func (m *Merkle) getAuthorizationPath(leaf *MerkleNode) []*MerkleNode {
    path := make([]*MerkleNode, m.Height)
    node := leaf
    for i := 0; i < m.Height; i++ {
        if node.Neighbour != nil {
            path[i] = node.Neighbour
        } else {
            path[i] = node
        }
        node = node.Parent
    }
    return path
}

func (m *Merkle) validate(leaf *MerkleNode, path []*MerkleNode) bool {
    h := leaf.Hash
    n := leaf
    for i, node := range path {
        if i == len(path) - 1 {
            return bytes.Equal(h, node.Hash)
        }
        if n != node {
            if n.Subscript < node.Subscript {
                h = crypto.ConcatHash256(n.Hash, node.Hash)
            } else {
                h = crypto.ConcatHash256(node.Hash, n.Hash)
            }
        } else {
            h = crypto.Hash256(n.Hash)
        }
        n = n.Parent
    }
    return false
}

func (m *Merkle) IsOnTrie(hash []byte) bool {
    for _, leaf := range m.Leaves {
        if bytes.Equal(hash, leaf.Hash) {
            path := m.getAuthorizationPath(leaf)
            return m.validate(leaf, path)
        }
    }
    return false
}