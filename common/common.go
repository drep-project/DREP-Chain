package common

import (
	"crypto/hmac"
	"crypto/sha512"
	"github.com/drep-project/DREP-Chain/crypto/sha3"
	"math/big"
	"math/rand"
	"net"
	"os"
	"time"
)

var Version int32 = 1

var (
	Big1   = big.NewInt(1)
	Big2   = big.NewInt(2)
	Big3   = big.NewInt(3)
	Big0   = big.NewInt(0)
	Big32  = big.NewInt(32)
	Big256 = big.NewInt(256)
	Big257 = big.NewInt(257)
)

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

	return sha3.Keccak256(append([]byte(uni), randBytes...)), nil
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
