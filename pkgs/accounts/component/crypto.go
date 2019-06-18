package component

import (
	"bytes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"github.com/drep-project/drep-chain/app"
	crypto2 "github.com/drep-project/drep-chain/crypto"
	"github.com/drep-project/drep-chain/crypto/sha3"
	"crypto/aes"
	"github.com/drep-project/drep-chain/crypto/secp256k1"
	chainTypes "github.com/drep-project/drep-chain/chain/types"
	"github.com/ethereum/go-ethereum/crypto"
	"golang.org/x/crypto/scrypt"
	"io"
)

type CryptedNode struct {
	Version   			int 				  `json:"version"`

	Data 				[]byte 				  `json:"-"`

	CipherText 			[]byte          		`json:"cipherText"`
	ChainId    			app.ChainIdType 		`json:"chainId"`
	ChainCode  			[]byte          		`json:"chainCode"`

	Cipher 				string				  `json:"cipher"`
	CipherParams   		CipherParams		  `json:"cipherParams"`

	KDFParams       	ScryptParams		  `json:"KDFParams"`
	MAC  				[]byte				  `json:"mac"`
}

type CipherParams struct {
	IV []byte `json:"iv"`
}

type ScryptParams struct {
    N  		int		`json:"n"`
	R       int		`json:"r"`
	P       int		`json:"p"`
	Dklen   int		`json:"dklen"`
	Salt 	[]byte 	`json:"salt"`
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
	derivedKey, err := getKDFKey(cryptoNode, auth)
	if err != nil {
		return nil, err
	}

	calculatedMAC := crypto.Keccak256(derivedKey[16:32], cryptoNode.CipherText)
	if !bytes.Equal(calculatedMAC, cryptoNode.MAC) {
		return nil, ErrDecrypt
	}

	plainText, err := aesCTRXOR(derivedKey[:16], cryptoNode.CipherText, cryptoNode.CipherParams.IV)
	if err != nil {
		return nil, err
	}
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
func BytesToCryptoNode(data []byte, auth string) (node *chainTypes.Node, errRef error) {
	defer func() {
		if err := recover(); err != nil {
			errRef = ErrDecryptFail
		}
	}()
	cryptoNode := new(CryptedNode)
	if err := json.Unmarshal(data, cryptoNode); err != nil {
		return nil, err
	}

	/*
	node2, errRef := EncryptData(data, []byte(auth),StandardScryptN, StandardScryptP)
	if errRef != nil {
		return
	}
	*/
	privD, errRef := DecryptData(*cryptoNode, auth)
	priv,pub := secp256k1.PrivKeyFromScalar(privD)
	addr := crypto2.PubKey2Address(pub)
	node =  &chainTypes.Node{
		Address  : &addr,
		PrivateKey :priv,
		ChainId    :cryptoNode.ChainId,
		ChainCode  :cryptoNode.ChainCode,
	}
	return
}
