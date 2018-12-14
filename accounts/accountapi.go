package accounts 

import (
	"os"
	"errors"
	"BlockChainTest/config"
)
type AccountApi struct {
   KeyStoreDir string
   ChainId config.ChainIdType
}

func (accountapi * AccountApi) AccountList() (string, error){
	if _, err := os.Stat(accountapi.KeyStoreDir); os.IsNotExist(err) {
		return "", err
	}
	node, err := OpenKeystore(accountapi.KeyStoreDir)
	if err != nil {
		return "", err
	}
	return node.Address().Hex(), nil
}

// create a new account
func (accountapi * AccountApi) CreateAccount() (string, error){
	if _, err := os.Stat(accountapi.KeyStoreDir); !os.IsNotExist(err) {
		return "", errors.New("keystore has exist")
	}
	  
	account, err := NewNormalAccount(nil,accountapi.ChainId)
	if err != nil {
		return "", err
	}
	SaveKeystore(account.Node, accountapi.KeyStoreDir)
	return account.Address.Hex(), nil
}