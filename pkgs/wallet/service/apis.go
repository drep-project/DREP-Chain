package service

import (
	"errors"
	"fmt"
	"github.com/drep-project/drep-chain/common"
	"io/ioutil"
	"math/big"

	chainService "github.com/drep-project/drep-chain/chain/service"
	chainTypes "github.com/drep-project/drep-chain/chain/types"
	walletTypes "github.com/drep-project/drep-chain/pkgs/wallet/types"
	"github.com/drep-project/drep-chain/crypto"
	"github.com/drep-project/drep-chain/crypto/secp256k1"
	"github.com/drep-project/drep-chain/crypto/sha3"
	"github.com/drep-project/drep-chain/database"
)

type AccountApi struct {
	Wallet          *Wallet
	accountService  *AccountService
	chainService    *chainService.ChainService
	databaseService *database.DatabaseService
}

func (accountapi *AccountApi) AddressList() ([]*secp256k1.PublicKey, error) {
	if !accountapi.Wallet.IsOpen() {
		return nil, errors.New("wallet is not open")
	}
	return accountapi.Wallet.ListKeys()
}

func (accountapi *AccountApi) CreateWallet(password string) error {
	wallet, err := CreateWallet(accountapi.accountService.Config, accountapi.chainService.Config.ChainId,password)
	if err != nil {
		return err
	}
	accountapi.Wallet = wallet
	return nil
}

// Lock lock the wallet to protect private key
func (accountapi *AccountApi) LockWallet() error {
	if !accountapi.Wallet.IsOpen() {
		return errors.New("wallet is not open")
	}
	if !accountapi.Wallet.IsLock() {
		return accountapi.Wallet.Lock()
	}
	return errors.New("wallet is already locked")
}

// UnLock unlock the wallet
func (accountapi *AccountApi) UnLockWallet(password string) error {
	if !accountapi.Wallet.IsOpen() {
		return errors.New("wallet is not open")
	}
	if accountapi.Wallet.IsLock() {
		return accountapi.Wallet.UnLock(password)
	}
	return errors.New("wallet is already unlock")
}

func (accountapi *AccountApi) OpenWallet(password string) error {
	return accountapi.Wallet.Open(password)
}

func (accountapi *AccountApi) CloseWallet() {
	accountapi.Wallet.Close()
}

func (accountapi *AccountApi) SuggestKey() (interface{}, error)  {
	if !accountapi.Wallet.IsOpen() {
		return nil,  errors.New("wallet is not open")
	}
	if accountapi.Wallet.IsLock() {
		return nil,   errors.New("wallet has locked")
	}
	pri, chainCode := chainTypes.RandomAccount()
	key := &walletTypes.Key{pri.PubKey(), pri}
	accountapi.Wallet.cacheStore.StoreKey(key, accountapi.Wallet.password)

	return &struct{
		Pubkey 		*secp256k1.PublicKey
		PriKey 		*secp256k1.PrivateKey
		ChainCode 	[]byte
	}{pri.PubKey(),pri,  chainCode}, nil
}

func (accountapi *AccountApi) RegisterAccount(fromAccount, newAccountName string, pk *secp256k1.PublicKey, chainCode []byte) (string, error)  {
	if !accountapi.Wallet.IsOpen() {
		return "", errors.New("wallet is not open")
	}
	if accountapi.Wallet.IsLock() {
		return "", errors.New("wallet has locked")
	}

	nonce := accountapi.chainService.GetTransactionCount(fromAccount)
	action := chainTypes.NewRegisterAccountAction(newAccountName, chainTypes.NewAuthority(*pk), accountapi.Wallet.chainId, chainCode)
	//amount *big.Int, nonce int64, gasPrice, gasLimit *big.Int, action interface{}
	tx, err := chainTypes.NewTransaction(fromAccount, chainTypes.RegisterAccountType, new (big.Int), nonce, chainTypes.DefaultGasPrice, chainTypes.GasTable[chainTypes.RegisterAccountType], action)
	if err != nil{
		return "",err
	}
	err = accountapi.chainService.SendTransaction(tx)
	if err != nil{
		return "",err
	}
	return tx.TxHash().String(), nil
}

func (accountapi *AccountApi) Transfer(from, to string, amount *big.Int) (string, error) {
	nonce := accountapi.chainService.GetTransactionCount(from)

	action := chainTypes.NewTransferAction(to)
	//amount *big.Int, nonce int64, gasPrice, gasLimit *big.Int, action interface{}
	tx, err := chainTypes.NewTransaction(from, chainTypes.TransferType, amount, nonce, chainTypes.DefaultGasPrice, chainTypes.GasTable[chainTypes.TransferType], action)
	if err != nil{
		return "",err
	}
	err = accountapi.chainService.SendTransaction(tx)
	if err != nil{
		return "",err
	}
	return tx.TxHash().String(), nil
}

func (accountapi *AccountApi) Call(from, to string, input []byte, amount *big.Int, readOnly bool) (string, error) {
	nonce := accountapi.chainService.GetTransactionCount(from)
	action := chainTypes.NewCallContractAction(to,input, readOnly)
	tx, err := chainTypes.NewTransaction(from, chainTypes.CallContractType, amount, nonce, chainTypes.DefaultGasPrice, chainTypes.GasTable[chainTypes.CallContractType], action)
	if err != nil{
		return "",err
	}
	err = accountapi.chainService.SendTransaction(tx)
	if err != nil{
		return "",err
	}
	return tx.TxHash().String(), nil
}

func (accountapi *AccountApi) CreateCode(from, to string, byteCode []byte) (string, error) {
	nonce := accountapi.chainService.GetTransactionCount(from)
	action := chainTypes.NewCreateContractAction(to, byteCode)
	tx, err := chainTypes.NewTransaction(from, chainTypes.CallContractType, nil, nonce, chainTypes.DefaultGasPrice, chainTypes.GasTable[chainTypes.CallContractType], action)
	if err != nil{
		return "",err
	}
	err = accountapi.chainService.SendTransaction(tx)
	if err != nil{
		return "",err
	}
	return tx.TxHash().String(), nil
}
func (accountapi *AccountApi) Test(name string) (string, error) {
	nonce := accountapi.chainService.GetTransactionCount("dreptest1")
	fi, err := ioutil.ReadFile("C:\\Users\\Drep\\Desktop\\proj\\drep-wasm-cdt-rust\\target\\wasm32-unknown-unknown\\release\\token.wasm")
	if err != nil{
		return "",err
	}
	action := chainTypes.NewCreateContractAction(name, fi)
	tx, err := chainTypes.NewTransaction("dreptest1", chainTypes.CreateContractType, nil, nonce, chainTypes.DefaultGasPrice, chainTypes.GasTable[chainTypes.CreateContractType], action)
	if err != nil{
		return "",err
	}
	err = accountapi.chainService.SendTransaction(tx)
	if err != nil{
		return "",err
	}
	return tx.TxHash().String(), nil
}

func (accountapi *AccountApi) Test2(name string) (*chainTypes.Storage, error) {
	return accountapi.databaseService.GetStorage(name, false)
}
func (accountapi *AccountApi) Test3(name string) (string, error) {
	nonce := accountapi.chainService.GetTransactionCount("dreptest1")
	/*
	fi, err := ioutil.ReadFile("C:\\Users\\Drep\\Desktop\\proj\\drep-wasm-cdt-rust\\target\\wasm32-unknown-unknown\\release\\token.wasm")
	if err != nil{
		return "",err
	}
	*/
	sink := common.ZeroCopySink{}
	sink.WriteString("transfer")
	fmt.Println(sink.Bytes())
	sink.WriteVarBytes([]byte("gg"))
	fmt.Println(sink.Bytes())
	sink.WriteVarBytes([]byte("tt"))
	fmt.Println(sink.Bytes())
	bytes :=common.PaddedBigBytes(big.NewInt(1000000),32)
	for i := 0; i < len(bytes)/2; i++ {
		bytes[i], bytes[len(bytes)-i-1] = bytes[len(bytes)-i-1], bytes[i]
	}
	sink.WriteBytes(bytes[0:32])
	fmt.Println(sink.Bytes())

	/*
	xxx := common.NewZeroCopySource(sink.Bytes())
	str,_,_,_ := xxx.NextString()
	fmt.Println(str)
	d,_,_,_ := xxx.NextVarBytes()
	fmt.Println(string(d))
	d,_,_,_ = xxx.NextVarBytes()
	fmt.Println(string(d))
	d,_ = xxx.NextBytes(32)

	for i := 0; i < len(d)/2; i++ {
		d[i], d[len(d)-i-1] = d[len(d)-i-1], d[i]
	}
	fmt.Println(new (big.Int).SetBytes(d).Int64())
*/
	action := chainTypes.NewCallContractAction(name, sink.Bytes() ,false)
	tx, err := chainTypes.NewTransaction("dreptest1", chainTypes.CallContractType, nil, nonce, chainTypes.DefaultGasPrice, chainTypes.GasTable[chainTypes.CallContractType], action)
	if err != nil{
		return "",err
	}
	err = accountapi.chainService.SendTransaction(tx)
	if err != nil{
		return "",err
	}
	return tx.TxHash().String(), nil
}
// DumpPrikey dumpPrivate
func (accountapi *AccountApi) DumpPrivkey(address *secp256k1.PublicKey) (*secp256k1.PrivateKey, error) {
	if !accountapi.Wallet.IsOpen() {
		return nil, errors.New("wallet is not open")
	}
	if accountapi.Wallet.IsLock() {
		return nil, errors.New("wallet has locked")
	}

	key, err := accountapi.Wallet.DumpPrivateKey(address)
	if err != nil {
		return nil, err
	}
	return key, nil
}

func (accountapi *AccountApi) Sign(key *secp256k1.PublicKey, msg string) ([]byte, error) {
	prv, _ := accountapi.DumpPrivkey(key)
	bytes := sha3.Hash256([]byte(msg))
	return crypto.Sign(bytes, prv)
}

func (accountapi *AccountApi) GasPrice() *big.Int {
	return chainTypes.DefaultGasPrice
}

func (accountapi *AccountApi) GetCode(accountName string) []byte {
	return accountapi.databaseService.GetByteCode(accountName, false)
}
