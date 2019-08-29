package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/drep-project/drep-chain/crypto"
	"github.com/drep-project/drep-chain/crypto/secp256k1"
)

func FromECDSAPub(pub *ecdsa.PublicKey) []byte {
	if pub == nil || pub.X == nil || pub.Y == nil {
		return nil
	}
	return elliptic.Marshal(secp256k1.S256(), pub.X, pub.Y)
}

func main() {
	type A interface {
		A()
	}

	type B interface {
		A
		B()
	}
	type C interface {
		A
		C()
	}
	type D interface {
		B
		C
	}
	bytexxxs, _ := hex.DecodeString("0002cf77d347ff7bdcd336c0b2375b1b2e3c8a56a3a8c9c18afaab427c4ee34da009040aba9500008c78cefbfeea05201f54a940a2c2eb36e4fbb6168e9f2627e30218b5d8a317fe4bca3491d5f3c65900000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")

	print16Array := func(xxxx []byte) {
		for _, val := range xxxx {
			fmt.Print(fmt.Sprintf("0x%x ,", val))
		}
		fmt.Println()
	}

	/*	ChainId        ChainIdType
		Version        int32
		PreviousHash   crypto.Hash
		GasLimit       big.Int
		GasUsed        big.Int
		Height         uint64
		Timestamp      uint64
		StateRoot      []byte
		TxRoot         []byte
		ReceiptRoot    crypto.Hash
		Bloom          Bloom*/
	prePos := 0
	pos := 0
	reader := bytes.NewReader(bytexxxs)
	binary.ReadUvarint(reader) //ChainId
	pos = int(reader.Size()) - reader.Len()
	print16Array(bytexxxs[prePos:pos])
	prePos = pos

	binary.ReadVarint(reader) //Version
	pos = int(reader.Size()) - reader.Len()
	print16Array(bytexxxs[prePos:pos])
	prePos = pos

	hash := make([]byte, 32)
	reader.Read(hash) //PreviousHash
	pos = int(reader.Size()) - reader.Len()
	print16Array(bytexxxs[prePos:pos])
	prePos = pos

	len, _ := binary.ReadUvarint(reader) //GasLimit
	gasLimit := make([]byte, len)
	reader.Read(gasLimit)
	pos = int(reader.Size()) - reader.Len()
	print16Array(bytexxxs[prePos:pos])
	prePos = pos

	len, _ = binary.ReadUvarint(reader) //gasUsed
	gasUsed := make([]byte, len)
	reader.Read(gasUsed)
	pos = int(reader.Size()) - reader.Len()
	print16Array(bytexxxs[prePos:pos])
	prePos = pos

	fmt.Println(binary.ReadUvarint(reader)) //Height
	pos = int(reader.Size()) - reader.Len()
	print16Array(bytexxxs[prePos:pos])
	prePos = pos

	binary.ReadUvarint(reader) //Timestamp
	pos = int(reader.Size()) - reader.Len()
	print16Array(bytexxxs[prePos:pos])
	prePos = pos

	len, _ = binary.ReadUvarint(reader) //txRoot
	txRoot := make([]byte, len)
	reader.Read(txRoot)
	pos = int(reader.Size()) - reader.Len()
	print16Array(bytexxxs[prePos:pos])
	prePos = pos

	len, _ = binary.ReadUvarint(reader) //StateRoot
	stateRoot := make([]byte, len)
	reader.Read(stateRoot)
	pos = int(reader.Size()) - reader.Len()
	print16Array(bytexxxs[prePos:pos])
	prePos = pos

	receiptRoot := make([]byte, 32) //ReceiptRoot
	reader.Read(receiptRoot)
	pos = int(reader.Size()) - reader.Len()
	print16Array(bytexxxs[prePos:pos])
	prePos = pos

	bloom := make([]byte, 256) //Bloom
	reader.Read(bloom)
	pos = int(reader.Size()) - reader.Len()
	print16Array(bytexxxs[prePos:pos])
	prePos = pos

	crypto.HexToAddress("0xaD3dC2D8aedef155eabA42Ab72C1FE480699336c")
}
