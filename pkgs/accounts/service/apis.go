package service

import (
	"errors"
	"github.com/drep-project/drep-chain/pkgs/accounts/addrgenerator"
	"math/big"

	"github.com/drep-project/drep-chain/chain/service/blockmgr"

	chainTypes "github.com/drep-project/drep-chain/chain/types"
	"github.com/drep-project/drep-chain/common"
	"github.com/drep-project/drep-chain/crypto"
	"github.com/drep-project/drep-chain/crypto/secp256k1"
	"github.com/drep-project/drep-chain/database"
)

/*
name: 账号rpc接口
usage: 地址管理及发起简单交易
prefix:account
*/
type AccountApi struct {
	Wallet          *Wallet
	accountService  *AccountService
	blockmgr        *blockmgr.BlockMgr
	databaseService *database.DatabaseService
}

/*
 name: listAddress
 usage: 列出所有本地地址
 return: 地址数组
 example:   curl http://localhost:15645 -X POST --data '{"jsonrpc":"2.0","method":"account_listAddress","params":[], "id": 3}' -H "Content-Type:application/json"
 response:
  {"jsonrpc":"2.0","id":3,"result":["0x3296d3336895b5baaa0eca3df911741bd0681c3f","0x3ebcbe7cb440dd8c52940a2963472380afbb56c5"]}
*/
func (accountapi *AccountApi) ListAddress() ([]*crypto.CommonAddress, error) {
	if !accountapi.Wallet.IsOpen() {
		return nil, ErrClosedWallet
	}
	return accountapi.Wallet.ListAddress()
}

/*
 name: createAccount
 usage: 创建本地账号
 return: 新账号地址信息
 example:   curl http://localhost:15645 -X POST --data '{"jsonrpc":"2.0","method":"account_createAccount","params":[], "id": 3}' -H "Content-Type:application/json"
 response:
	  {"jsonrpc":"2.0","id":3,"result":"0x2944c15c466fad03ec1282bab579dec5a0cf0fa3"}
*/
func (accountapi *AccountApi) CreateAccount() (*crypto.CommonAddress, error) {
	if !accountapi.Wallet.IsOpen() {
		return nil, ErrClosedWallet
	}
	newAaccount, err := accountapi.Wallet.NewAccount()
	if err != nil {
		return nil, err
	}
	return newAaccount.Address, nil
}

/*
 name: createWallet
 usage: 创建本地钱包
 params:
	1. 钱包密码
 return: 无
 example:   curl http://localhost:15645 -X POST --data '{"jsonrpc":"2.0","method":"account_createWallet","params":["123"], "id": 3}' -H "Content-Type:application/json"
 response:
	  {"jsonrpc":"2.0","id":3,"result":null}
*/
func (accountapi *AccountApi) CreateWallet(password string) error {
	err := accountapi.accountService.CreateWallet(password)
	if err != nil {
		return err
	}
	return accountapi.OpenWallet(password)
}

/*
 name: lockWallet
 usage: 锁定钱包（无法发起需要私钥的相关工作）
 params:
 return: 无
 example:   curl http://localhost:15645 -X POST --data '{"jsonrpc":"2.0","method":"account_lockWallet","params":[], "id": 3}' -H "Content-Type:application/json"
 response:
	 {"jsonrpc":"2.0","id":3,"result":null}
*/
func (accountapi *AccountApi) LockWallet() error {
	if !accountapi.Wallet.IsOpen() {
		return ErrClosedWallet
	}
	if !accountapi.Wallet.IsLock() {
		return accountapi.Wallet.Lock()
	}
	return ErrLockedWallet
}

/*
 name: lockWallet
 usage: 解锁钱包
 params:
	1. 钱包密码
 return: 无
 example:   curl http://localhost:15645 -X POST --data '{"jsonrpc":"2.0","method":"account_openWallet","params":["123"], "id": 3}' -H "Content-Type:application/json"
 response:
	 {"jsonrpc":"2.0","id":3,"result":null}
*/
func (accountapi *AccountApi) UnLockWallet(password string) error {
	if !accountapi.Wallet.IsOpen() {
		return ErrClosedWallet
	}
	if accountapi.Wallet.IsLock() {
		return accountapi.Wallet.UnLock(password)
	}
	return ErrAlreadyUnLocked
}

/*
 name: openWallet
 usage: 打开钱包
 params:
	1. 钱包密码
 return: 无
 example:   curl http://localhost:15645 -X POST --data '{"jsonrpc":"2.0","method":"account_openWallet","params":["123"], "id": 3}' -H "Content-Type:application/json"
 response:
	 {"jsonrpc":"2.0","id":3,"result":null}
*/
func (accountapi *AccountApi) OpenWallet(password string) error {
	return accountapi.Wallet.Open(password)
}

/*
 name: closeWallet
 usage: 关闭钱包
 params:
 return: 无
 example:   curl http://localhost:15645 -X POST --data '{"jsonrpc":"2.0","method":"account_closeWallet","params":[], "id": 3}' -H "Content-Type:application/json"
 response:
	 {"jsonrpc":"2.0","id":3,"result":null}
*/
func (accountapi *AccountApi) CloseWallet() {
	accountapi.Wallet.Close()
}

/*
 name: transfer
 usage: 转账
 params:
	1. 发起转账的地址
	2. 接受者的地址
	3. 金额
	4. gas价格
	5. gas上线
	6. 备注
 return: 交易地址
 example:   curl -H "Content-Type: application/json" -X post --data '{"jsonrpc":"2.0","method":"account_transfer","params":["0x3ebcbe7cb440dd8c52940a2963472380afbb56c5","0x3ebcbe7cb440dd8c52940a2963472380afbb56c5","0x111","0x110","0x30000",""],"id":1}' http://127.0.0.1:15645
 response:
	 {"jsonrpc":"2.0","id":1,"result":"0x3a3b59f90a21c2fd1b690aa3a2bc06dc2d40eb5bdc26fdd7ecb7e1105af2638e"}
*/
func (accountapi *AccountApi) Transfer(from crypto.CommonAddress, to crypto.CommonAddress, amount, gasprice, gaslimit *common.Big, data common.Bytes) (string, error) {
	nonce := accountapi.blockmgr.GetTransactionCount(&from)
	tx := chainTypes.NewTransaction(to, (*big.Int)(amount), (*big.Int)(gasprice), (*big.Int)(gaslimit), nonce)
	sig, err := accountapi.Wallet.Sign(&from, tx.TxHash().Bytes())
	if err != nil {
		return "", err
	}
	tx.Sig = sig
	err = accountapi.blockmgr.SendTransaction(tx, true)
	if err != nil {
		return "", err
	}
	return tx.TxHash().String(), nil
}

/*
 name: ReplaceTx
 usage: 替换老的交易
 params:
	1. 发起转账的地址
	2. 接受者的地址
	3. 金额
	4. gas价格
	5. gas上线
	6. 备注
	7. 被代替交易的nonce
 return: 新交易地址
 example: curl -H "Content-Type: application/json" -X post --data '{"jsonrpc":"2.0","method":"account_replaceTx","params":["0x3ebcbe7cb440dd8c52940a2963472380afbb56c5","0x3ebcbe7cb440dd8c52940a2963472380afbb56c5","0x111","0x110","0x30000","",1000],"id":1}' http://127.0.0.1:15645
 response:
	 {"jsonrpc":"2.0","id":1,"result":"0x3a3b59f90a21c2fd1b690aa3a2bc06dc2d40eb5bdc26fdd7ecb7e1105af2638e"}
*/
func (accountapi *AccountApi) ReplaceTx(from crypto.CommonAddress, to crypto.CommonAddress, amount, gasprice, gaslimit *common.Big, data common.Bytes, nonce *uint64) (string, error) {
	if nonce == nil {
		return "", errors.New("nonce is nil")
	}

	tx := chainTypes.NewTransaction(to, (*big.Int)(amount), (*big.Int)(gasprice), (*big.Int)(gaslimit), *nonce)
	sig, err := accountapi.Wallet.Sign(&from, tx.TxHash().Bytes())
	if err != nil {
		return "", err
	}
	tx.Sig = sig
	err = accountapi.blockmgr.SendTransaction(tx, true)
	if err != nil {
		return "", err
	}
	return tx.TxHash().String(), nil
}

/*
 name: GetTxInPool
 usage: 查询交易是否在交易池，如果在，返回交易
 params:
	1. 发起转账的地址

 return: 交易完整信息
 example: curl -H "Content-Type: application/json" -X post --data '{"jsonrpc":"2.0","method":"account_getTxInPool","params":["0x3ebcbe7cb440dd8c52940a2963472380afbb56c5"],"id":1}' http://127.0.0.1:15645
 response:
	 {"jsonrpc":"2.0","id":1,"result":transaction}
*/
func (accountapi *AccountApi) GetTxInPool(hash string) (*chainTypes.Transaction, error) {
	return accountapi.blockmgr.GetTxInPool(hash)
}

/*
 name: setAlias
 usage: 设置别名
 params:
	1. 带设置别名的地址
	2. 别名
	3. gas价格
	4. gas上限
 return: 交易地址
 example:
	curl -H "Content-Type: application/json" -X post --data '{"jsonrpc":"2.0","method":"account_setAlias","params":["0x3ebcbe7cb440dd8c52940a2963472380afbb56c5","AAAAA","0x110","0x30000"],"id":1}' http://127.0.0.1:15645
response:
	{"jsonrpc":"2.0","id":1,"result":"0x5adb248f2943e12fb91c140bd3d0df6237712061e9abae97345b0869c3daa749"}
*/
func (accountapi *AccountApi) SetAlias(srcAddr crypto.CommonAddress, alias string, gasprice, gaslimit *common.Big) (string, error) {
	nonce := accountapi.blockmgr.GetTransactionCount(&srcAddr)
	t := chainTypes.NewAliasTransaction(alias, (*big.Int)(gasprice), (*big.Int)(gaslimit), nonce)
	sig, err := accountapi.Wallet.Sign(&srcAddr, t.TxHash().Bytes())
	if err != nil {
		return "", err
	}
	t.Sig = sig
	err = accountapi.blockmgr.SendTransaction(t, true)
	if err != nil {
		return "", err
	}
	return t.TxHash().String(), nil
}

/*
 name: call
 usage: 调用合约
 params:
	1. 调用者的地址
	2. 合约地址
	3. 代码
	4. 金额
	4. gas价格
	5. gas上限
 return: 合约地址
 example:
 	curl -H "Content-Type: application/json" -X post --data '{"jsonrpc":"2.0","method":"account_createCode","params":["0x3ebcbe7cb440dd8c52940a2963472380afbb56c5","0x6d4ce63c","0x111","0x110","0x30000"],"id":1}' http://127.0.0.1:15645
 response:
	 {"jsonrpc":"2.0","id":1,"result":"0x5d74aba54ace5f01a5f0057f37bfddbbe646ea6de7265b368e2e7d17d9cdeb9c"}
*/
func (accountapi *AccountApi) Call(from crypto.CommonAddress, to crypto.CommonAddress, input common.Bytes, amount, gasprice, gaslimit *common.Big) (string, error) {
	nonce := accountapi.blockmgr.GetTransactionCount(&from)
	t := chainTypes.NewCallContractTransaction(to, input, (*big.Int)(amount), (*big.Int)(gasprice), (*big.Int)(gaslimit), nonce)
	sig, err := accountapi.Wallet.Sign(&from, t.TxHash().Bytes())
	if err != nil {
		return "", err
	}
	t.Sig = sig
	accountapi.blockmgr.SendTransaction(t, true)
	return t.TxHash().String(), nil
}

/*
 name: createCode
 usage: 部署合约
 params:
	1. 部署合约的地址
	2. 合约内容
	3. 金额
	4. gas价格
	5. gas上线
 return: 合约地址
 example:
 	curl -H "Content-Type: application/json" -X post --data '{"jsonrpc":"2.0","method":"account_createCode","params":["0x3ebcbe7cb440dd8c52940a2963472380afbb56c5","0x608060405234801561001057600080fd5b5061018c806100206000396000f3fe608060405260043610610051576000357c0100000000000000000000000000000000000000000000000000000000900480634f2be91f146100565780636d4ce63c1461006d578063db7208e31461009e575b600080fd5b34801561006257600080fd5b5061006b6100dc565b005b34801561007957600080fd5b5061008261011c565b604051808260070b60070b815260200191505060405180910390f35b3480156100aa57600080fd5b506100da600480360360208110156100c157600080fd5b81019080803560070b9060200190929190505050610132565b005b60016000808282829054906101000a900460070b0192506101000a81548167ffffffffffffffff021916908360070b67ffffffffffffffff160217905550565b60008060009054906101000a900460070b905090565b806000806101000a81548167ffffffffffffffff021916908360070b67ffffffffffffffff1602179055505056fea165627a7a723058204b651e4313ab6bc4eda61084cac1f805699cefbb979ddfd3a2d7f970903307cd0029","0x111","0x110","0x30000"],"id":1}' http://127.0.0.1:15645
 response:
	 {"jsonrpc":"2.0","id":1,"result":"0x9a8d8d5d7d00bbe0eb1b9431a13a7219008e352241b751b177bfb29e4e75b0d1"}
*/
func (accountapi *AccountApi) CreateCode(from crypto.CommonAddress, byteCode common.Bytes, amount, gasprice, gaslimit *common.Big) (string, error) {
	nonce := accountapi.blockmgr.GetTransactionCount(&from)
	t := chainTypes.NewContractTransaction(byteCode, (*big.Int)(gasprice), (*big.Int)(gaslimit), nonce)
	sig, err := accountapi.Wallet.Sign(&from, t.TxHash().Bytes())
	if err != nil {
		return "", err
	}
	t.Sig = sig
	accountapi.blockmgr.SendTransaction(t, true)
	return t.TxHash().String(), nil
}

/*
	 name: dumpPrivkey
	 usage: 关闭钱包
	 params:
		1.地址
	 return: 私钥
	 example:   curl http://localhost:15645 -X POST --data '{"jsonrpc":"2.0","method":"account_dumpPrivkey","params":["0x3ebcbe7cb440dd8c52940a2963472380afbb56c5"], "id": 3}' -H "Content-Type:application/json"
	 response:
		 {"jsonrpc":"2.0","id":3,"result":"0x270f4b122603999d1c07aec97e972a2ddf7bd8b5bfe3543c10814e6a19f13aaf"}
*/
func (accountapi *AccountApi) DumpPrivkey(address *crypto.CommonAddress) (*secp256k1.PrivateKey, error) {
	if !accountapi.Wallet.IsOpen() {
		return nil, ErrClosedWallet
	}
	if accountapi.Wallet.IsLock() {
		return nil, ErrLockedWallet
	}

	node, err := accountapi.Wallet.GetAccountByAddress(address)
	if err != nil {
		return nil, err
	}
	return node.PrivateKey, nil
}

/*
	 name: sign
	 usage: 关闭钱包
	 params:
		1.地址
		2.消息hash
	 return: 私钥
	 example:
		curl http://localhost:15645 -X POST --data '{"jsonrpc":"2.0","method":"account_sign","params":["0x3ebcbe7cb440dd8c52940a2963472380afbb56c5", "0x00001c9b8c8fdb1f53faf02321f76253704123e2b56cce065852bab93e526ae2"], "id": 3}' -H "Content-Type:application/json"

	response:
		 {"jsonrpc":"2.0","id":3,"result":"0x1f1d16412468dd9b67b568d31839ac608bdfddf2580666db4d364eefbe285fdaed569a3c8fa1decfebbfa0ed18b636059dbbf4c2106c45fc8846909833ef2cb1de"}
*/
func (accountapi *AccountApi) Sign(address crypto.CommonAddress, hash common.Bytes) (common.Bytes, error) {
	sig, err := accountapi.Wallet.Sign(&address, hash)
	if err != nil {
		return nil, err
	}
	return sig, nil
}

/*
	 name: generateAddresses
	 usage: 生成其他链的地址
	 params:
		1.drep地址
	 return: {BTCaddress, ethAddress, neoAddress}
	 example:
		curl http://localhost:15645 -X POST --data '{"jsonrpc":"2.0","method":"account_generateAddresses","params":["0x3ebcbe7cb440dd8c52940a2963472380afbb56c5"], "id": 3}' -H "Content-Type:application/json"

	response:
		 {"jsonrpc":"2.0","id":3,"result":"0x1f1d16412468dd9b67b568d31839ac608bdfddf2580666db4d364eefbe285fdaed569a3c8fa1decfebbfa0ed18b636059dbbf4c2106c45fc8846909833ef2cb1de"}
*/
func (accountapi *AccountApi) GenerateAddresses(address crypto.CommonAddress) (*RpcAddresses, error) {
	privkey, err := accountapi.Wallet.DumpPrivateKey(&address)
	if err != nil {
		return nil, err
	}
	generator := &addrgenerator.AddrGenerate{
		PrivateKey: privkey,
	}
	return &RpcAddresses {
		BtcAddress:generator.ToBtc(),
		EthAddress:generator.ToEth(),
		NeoAddress:generator.ToNeo(),
		RippleAddress:generator.ToRipple(),
		DashAddress:generator.ToDash(),
		DogeCoinAddress:generator.ToDogecoin(),
		LiteCoiAddress:generator.ToLiteCoin(),
	}, nil
}

type RpcAddresses struct {
	BtcAddress string
	EthAddress string
	NeoAddress string
	RippleAddress string
	DashAddress string
	DogeCoinAddress string
	LiteCoiAddress string
}

type RpcAccount struct {
	Addr   *crypto.CommonAddress
	Pubkey string
}
