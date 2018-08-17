package network

import (
	"crypto/sha256"
)

func HashEnc(plainText []byte) (cipherText []byte) {
	h := sha256.New()
	h.Write(plainText)
	cipherText = h.Sum(nil)
	return
}