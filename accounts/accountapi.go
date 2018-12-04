package accounts 

import (
	"BlockChainTest/mycrypto"
)

type AccountApi struct {

}


func (accountapi * AccountApi) GetNewAddress() CommonAddress{
	prik,_ := mycrypto.GeneratePrivateKey()
	return PubKey2Address(prik.PubKey)
}