package component

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"github.com/drep-project/drep-chain/crypto/sha3"
	"golang.org/x/crypto/scrypt"
	"testing"
)


type DataAndPassword struct {
	Data string
	Password string
	Salt   string
	Iv    string
}
var (
    TestEncryptData = map[DataAndPassword]string{
		DataAndPassword{"aaaaaaaaa","123","5a832928a3c87eb118f1f0e5d9882a5cf41a385f51ae146620ab9ecb51c4f0f7","784f477c7feda7e442bff3823b5786c4"}: "febc5d22b68ce4a186",
		DataAndPassword{"s23565#12vgfdg3","123","eb6d3b5a80d01adcf5cd0d1198c17a70d2291c437449e61e477c59ace89f22e0","2a40304885f4a2837e43b7db62cc16b2"}: "9bfed037396f6c5264a8dce588b03d",
		DataAndPassword{"dvhryb qweqecryvtce","dfbv3v54v776b48c2","736af3950f05ca7671dfb3c23bed22b702248040c199e11ad20b9fad5c3fef43","c7ca8c9409a7d813079732dd68c9d042"}: "e897e49ff3544689e011537782aba8c46561cd",
	}
    TestDecryptData = map[*CryptedNode]DataAndPassword{	}
)
func init(){
	makeTestData()
}

func makeTestData() {
	for i:= 0 ; i< 5; i++ {
		data :=   []byte("dvhryb qweqecryvtce")
		password := []byte("dfbv3v54v776b48c2")
		cryptoNode := &CryptedNode{
			Version:      0,
			Data:      data ,
			Cipher:       "aes-128-ctr",
			CipherParams: CipherParams{},
			KDFParams:ScryptParams{
				N  	:StandardScryptN,
				R    :scryptR,
				P      :StandardScryptP,
				Dklen   :scryptDKLen,
			},
		}
		cryptoNode.EncryptData(password)
		rawdata, _ := json.Marshal(cryptoNode)
		newNode := &CryptedNode{ }
		json.Unmarshal(rawdata, newNode)
		newNode.Data = nil
		TestDecryptData[newNode] = DataAndPassword{
			Data :string(data),
			Password: string(password),
		}
		TestEncryptData[DataAndPassword{string(data),string(password),hex.EncodeToString(cryptoNode.KDFParams.Salt),hex.EncodeToString(cryptoNode.CipherParams.IV)}] = hex.EncodeToString(cryptoNode.CipherText)
	}
}

func TestEncrypt(t *testing.T) {
	for key, cipherText:= range  TestEncryptData {
		cryptoNode := &CryptedNode{
			Version:      0,
			Data:         []byte(key.Data),
			Cipher:       "aes-128-ctr",
			CipherParams: CipherParams{},
			KDFParams:ScryptParams{
				N  	:StandardScryptN,
				R    :scryptR,
				P      :StandardScryptP,
				Dklen   :scryptDKLen,
			},
		}
		salt,_ := hex.DecodeString(key.Salt)
		iv, _ := hex.DecodeString(key.Iv)
		encryptDataByGivenIvSalt(cryptoNode,[]byte(key.Password), salt, iv)
		mychipper := hex.EncodeToString(cryptoNode.CipherText)
		if mychipper != cipherText {
			t.Error("encrypt data not match given value", mychipper, "!=",cipherText )
		}
	}
}

func TestDecrypt(t *testing.T) {
	for cryptoNode, cipherText:= range  TestDecryptData {
		DecryptData(*cryptoNode, cipherText.Password)
		if bytes.Equal(cryptoNode.Data, []byte(cipherText.Data)) {
			t.Error("encrypt data not match given value",  []byte(cipherText.Data), "!=",cipherText )
		}
	}
}

func  encryptDataByGivenIvSalt(node *CryptedNode, auth []byte, salt []byte, iv []byte) error {
	derivedKey, err := scrypt.Key(auth, salt, node.KDFParams.N, node.KDFParams.R, node.KDFParams.P, node.KDFParams.Dklen)
	if err != nil {
		return err
	}
	encryptKey := derivedKey[:16]
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