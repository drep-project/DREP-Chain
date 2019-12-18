package component

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"

	crypto2 "github.com/drep-project/DREP-Chain/crypto"
	"github.com/drep-project/DREP-Chain/crypto/secp256k1"
	"github.com/drep-project/DREP-Chain/crypto/sha3"
	"github.com/drep-project/DREP-Chain/types"
	"github.com/drep-project/binary"
	"github.com/ethereum/go-ethereum/crypto"
	"golang.org/x/crypto/scrypt"
	"io"
	"time"
)

type CryptedNode struct {
	Version int `json:"version"`

	Data []byte `json:"-"`

	CipherText []byte            `json:"cipherText"`
	ChainId    types.ChainIdType `json:"chainId"`
	ChainCode  []byte            `json:"chainCode"`

	Cipher       string       `json:"cipher"`
	CipherParams CipherParams `json:"cipherParams"`

	KDFParams ScryptParams `json:"KDFParams"`
	MAC       []byte       `json:"mac"`
}

type CipherParams struct {
	IV []byte `json:"iv"`
}

type ScryptParams struct {
	N     int    `json:"n"`
	R     int    `json:"r"`
	P     int    `json:"p"`
	Dklen int    `json:"dklen"`
	Salt  []byte `json:"salt"`
}

// Encryptdata encrypts the data given as 'data' with the password 'auth'.
func (node *CryptedNode) EncryptData(auth []byte) error {
	salt := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		panic("reading from crypto/rand failed: " + err.Error())
	}
	derivedKey, err := scrypt.Key(auth, salt, node.KDFParams.N, node.KDFParams.R, node.KDFParams.P, node.KDFParams.Dklen)
	if err != nil {
		return err
	}
	encryptKey := derivedKey[:16]

	iv := make([]byte, aes.BlockSize) // 16
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic("reading from crypto/rand failed: " + err.Error())
	}
	cipherText, err := aesCTRXOR(encryptKey, node.Data, iv)
	if err != nil {
		return err
	}
	mac := sha3.Keccak256(derivedKey[16:32], cipherText)

	node.CipherParams.IV = iv
	node.CipherText = cipherText
	node.MAC = mac
	node.KDFParams.Salt = salt
	return nil
}

func DecryptData(cryptoNode CryptedNode, auth string) ([]byte, error) {
	if cryptoNode.Cipher != "aes-128-ctr" {
		return nil, fmt.Errorf("Cipher not supported: %v", cryptoNode.Cipher)
	}

	fmt.Println("1:", time.Now().Unix(), time.Now().Nanosecond())

	derivedKey, err := getKDFKey(cryptoNode, auth)
	if err != nil {
		return nil, err
	}

	fmt.Println("2:", time.Now().Unix(), time.Now().Nanosecond())

	calculatedMAC := crypto.Keccak256(derivedKey[16:32], cryptoNode.CipherText)
	if !bytes.Equal(calculatedMAC, cryptoNode.MAC) {
		return nil, ErrDecrypt
	}

	fmt.Println("3:", time.Now().Unix(), time.Now().Nanosecond())

	plainText, err := aesCTRXOR(derivedKey[:16], cryptoNode.CipherText, cryptoNode.CipherParams.IV)
	if err != nil {
		return nil, err
	}
	fmt.Println("4:", time.Now().Unix(), time.Now().Nanosecond())
	return plainText, err
}

func getKDFKey(cryptoNode CryptedNode, auth string) ([]byte, error) {
	authArray := []byte(auth)
	return scrypt.Key(authArray, cryptoNode.KDFParams.Salt, cryptoNode.KDFParams.N, cryptoNode.KDFParams.R, cryptoNode.KDFParams.P, cryptoNode.KDFParams.Dklen)
}

func aesCTRXOR(key, inText, iv []byte) ([]byte, error) {
	// AES-128 is selected due to size of encryptKey.
	aesBlock, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	stream := cipher.NewCTR(aesBlock, iv)
	outText := make([]byte, len(inText))
	stream.XORKeyStream(outText, inText)
	return outText, err
}

// BytesToCryptoNode cocnvert given bytes and password to a node
func BytesToCryptoNode(data []byte, auth string) (node *types.Node, errRef error) {
	defer func() {
		if err := recover(); err != nil {
			errRef = ErrDecryptFail
		}
	}()

	fmt.Println("111:", time.Now().Unix(), time.Now().Nanosecond())
	cryptoNode := new(CryptedNode)

	err := binary.Unmarshal(data, cryptoNode)
	if err != nil {
		return nil, err
	}

	fmt.Println("222:", time.Now().Unix(), time.Now().Nanosecond())
	/*
		node2, errRef := EncryptData(data, []byte(auth),StandardScryptN, StandardScryptP)
		if errRef != nil {
			return
		}
	*/

	fmt.Println("333:", time.Now().Unix(), time.Now().Nanosecond())
	privD, errRef := DecryptData(*cryptoNode, auth)
	fmt.Println("444:", time.Now().Unix(), time.Now().Nanosecond())
	priv, pub := secp256k1.PrivKeyFromScalar(privD)
	fmt.Println("555:", time.Now().Unix(), time.Now().Nanosecond())
	addr := crypto2.PubkeyToAddress(pub)
	fmt.Println("666:", time.Now().Unix(), time.Now().Nanosecond())
	node = &types.Node{
		Address:    &addr,
		PrivateKey: priv,
		ChainId:    cryptoNode.ChainId,
		ChainCode:  cryptoNode.ChainCode,
	}
	return
}
