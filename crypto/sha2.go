package crypto

import (
	"crypto/sha256"
	"math/big"
)

func Hash256(plainText []byte) ([]byte) {
	h := sha256.New()
	h.Write(plainText)
	ret := h.Sum(nil)
	hash := make([]byte, 32)
	copy(hash[32 - len(ret):], ret)
	return hash
}

func KDF(plainText []byte) ([]byte) {
	pLen := len(plainText)
	k := pLen / 32
	if pLen - k * 32 > 0 {
		k += 1
	}
	hash := make([]byte, k * 32)
	p := new(big.Int).SetBytes(plainText)
	p.Lsh(p, 8)
	count := 0
	for count < k {
		c := new(big.Int).SetInt64(int64(count))
		b := new(big.Int).Add(p, c).Bytes()
		h := Hash256(b)
		copy(hash[count * 32: (count + 1) * 32], h)
		count += 1
	}
	return hash
}