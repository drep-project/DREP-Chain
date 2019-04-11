package common

import (
	"crypto/hmac"
	"crypto/sha512"
	"github.com/drep-project/drep-chain/crypto/sha3"
	"math/rand"
	"net"
	"os"
	"time"
)

var Version uint64 = 1

func HmAC(message, key []byte) []byte {
	h := hmac.New(sha512.New, key)
	h.Write(message)
	return h.Sum(nil)
}

func GenUnique() ([]byte, error) {
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

	randBytes := make([]byte, 64)
	_, err = rand.Read(randBytes)
	if err != nil {
		panic("key generation: could not read from random source: " + err.Error())
	}

	return sha3.Hash256(append([]byte(uni), randBytes...)), nil
}

// FileExist checks if a file exists at filePath.
func FileExist(filePath string) bool {
	_, err := os.Stat(filePath)
	if err != nil && os.IsNotExist(err) {
		return false
	}

	return true
}


type PrettyDuration time.Duration
