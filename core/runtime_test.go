package _our

import (
	"math/big"
	"BlockChainTest/core/vm"
	"testing"
	"fmt"
	"BlockChainTest/database"
	"BlockChainTest/core/abi"
	"strings"
	"encoding/hex"
	"BlockChainTest/accounts"
	"encoding/json"
	"BlockChainTest/bean"
)

// Execute executes the code using the input as call data during the execution.
// It returns the EVM's return value, the new state and an error if it failed.
//
// Executes sets up a in memory, temporarily, environment for the execution of
// the given code. It makes sure that it's restored to it's original state afterwards.
func ExecuteCreate(code []byte) {
	evm := vm.NewEVM()
	s1 := "111111"
	s2 := "222222"
	var chainId int64 = 0
	callerAddr1 := accounts.Hex2Address(s1)
	callerAddr2 := accounts.Hex2Address(s2)
	caller1 := &accounts.Account{Address: callerAddr1, Storage: &accounts.Storage{Balance: new(big.Int).SetInt64(100)}}
	caller2 := &accounts.Account{Address: callerAddr2, Storage: &accounts.Storage{Balance: new(big.Int).SetInt64(200)}}
	errPut1 := database.PutStorage(callerAddr1, chainId, caller1.Storage)
	errPut2 := database.PutStorage(callerAddr2, chainId, caller2.Storage)
	fmt.Println("errPut1: ", errPut1)
	fmt.Println("errPut2: ", errPut2)
	gas := uint64(1000000)
	value := new(big.Int).SetInt64(0)
	ret1, _, _, err1 := evm.CreateContractCode(callerAddr1, chainId, code, gas, value)
	ret2, _, _, err2 := evm.CreateContractCode(callerAddr1, chainId, code, gas, value)
	fmt.Println("err1: ", err1)
	fmt.Println("err2: ", err2)
	fmt.Println("ret1: ", ret1)
	fmt.Println("ret2: ", ret2)
}

func ExecuteCall(input []byte) {
	evm := vm.NewEVM()
	s1 := "111111"
	callerAddr := accounts.Hex2Address(s1)
	s2 := "bf101a61d5cc3d5f0b3e2c44c967c3c725b29017"
	gas := uint64(1000000)
	value := new(big.Int).SetInt64(0)
	contractAddr := accounts.Hex2Address(s2)
	var chainId int64 = 0
	evm.CallContractCode(callerAddr, contractAddr, chainId, input, gas, value)
}

func TestCreate(t *testing.T) {
	code, err := hex.DecodeString("608060405260043610610078576000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff168063095ea7b31461007d57806323b872dd146100e2578063343d2f1a1461016757806370a0823114610194578063a9059cbb146101eb578063e0b1cccb14610250575b600080fd5b34801561008957600080fd5b506100c8600480360381019080803573ffffffffffffffffffffffffffffffffffffffff1690602001909291908035906020019092919050505061029d565b604051808215151515815260200191505060405180910390f35b3480156100ee57600080fd5b5061014d600480360381019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190803573ffffffffffffffffffffffffffffffffffffffff1690602001909291908035906020019092919050505061038f565b604051808215151515815260200191505060405180910390f35b34801561017357600080fd5b5061019260048036038101908080359060200190929190505050610606565b005b3480156101a057600080fd5b506101d5600480360381019080803573ffffffffffffffffffffffffffffffffffffffff16906020019092919050505061064c565b6040518082815260200191505060405180910390f35b3480156101f757600080fd5b50610236600480360381019080803573ffffffffffffffffffffffffffffffffffffffff16906020019092919080359060200190929190505050610694565b604051808215151515815260200191505060405180910390f35b34801561025c57600080fd5b5061029b600480360381019080803573ffffffffffffffffffffffffffffffffffffffff1690602001909291908035906020019092919050505061080b565b005b600081600160003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060008573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020819055508273ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff167f8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925846040518082815260200191505060405180910390a36001905092915050565b6000816000808673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002054036000808673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000208190555081600160008673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000205403600160008673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002081905550816000808573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002054016000808573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020819055508273ffffffffffffffffffffffffffffffffffffffff168473ffffffffffffffffffffffffffffffffffffffff167fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef846040518082815260200191505060405180910390a3600190509392505050565b806000803373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000208190555050565b60008060008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020549050919050565b6000816000803373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002054036000803373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002081905550816000808573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002054016000808573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020819055508273ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff167fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef846040518082815260200191505060405180910390a36001905092915050565b806000808473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000208190555050505600a165627a7a72305820b0ebd8ff75aaf318488cf838557b60d2a4cb689b7882a240db91d4c949be445e0029")
	if err != nil {
		fmt.Println("bad byte code")
	}
	ExecuteCreate(code)
}

func TestCallMyCode(t *testing.T) {
	var mystr = `[
	{
		"constant": false,
		"inputs": [
			{
				"name": "spender",
				"type": "address"
			},
			{
				"name": "tokens",
				"type": "uint256"
			}
		],
		"name": "approve",
		"outputs": [
			{
				"name": "success",
				"type": "bool"
			}
		],
		"payable": false,
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"constant": false,
		"inputs": [
			{
				"name": "from",
				"type": "address"
			},
			{
				"name": "to",
				"type": "address"
			},
			{
				"name": "tokens",
				"type": "uint256"
			}
		],
		"name": "transferFrom",
		"outputs": [
			{
				"name": "success",
				"type": "bool"
			}
		],
		"payable": false,
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"constant": false,
		"inputs": [
			{
				"name": "bal",
				"type": "uint256"
			}
		],
		"name": "updateCurrentBalance",
		"outputs": [],
		"payable": false,
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"constant": true,
		"inputs": [
			{
				"name": "tokenOwner",
				"type": "address"
			}
		],
		"name": "balanceOf",
		"outputs": [
			{
				"name": "balance",
				"type": "uint256"
			}
		],
		"payable": false,
		"stateMutability": "view",
		"type": "function"
	},
	{
		"constant": false,
		"inputs": [
			{
				"name": "to",
				"type": "address"
			},
			{
				"name": "tokens",
				"type": "uint256"
			}
		],
		"name": "transfer",
		"outputs": [
			{
				"name": "success",
				"type": "bool"
			}
		],
		"payable": false,
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"constant": false,
		"inputs": [
			{
				"name": "addr",
				"type": "address"
			},
			{
				"name": "bal",
				"type": "uint256"
			}
		],
		"name": "updateBalance",
		"outputs": [],
		"payable": false,
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"anonymous": false,
		"inputs": [
			{
				"indexed": true,
				"name": "from",
				"type": "address"
			},
			{
				"indexed": true,
				"name": "to",
				"type": "address"
			},
			{
				"indexed": false,
				"name": "tokens",
				"type": "uint256"
			}
		],
		"name": "Transfer",
		"type": "event"
	},
	{
		"anonymous": false,
		"inputs": [
			{
				"indexed": true,
				"name": "tokenOwner",
				"type": "address"
			},
			{
				"indexed": true,
				"name": "spender",
				"type": "address"
			},
			{
				"indexed": false,
				"name": "tokens",
				"type": "uint256"
			}
		],
		"name": "Approval",
		"type": "event"
	}
]`

	myabi, err := abi.JSON(strings.NewReader(mystr))
	if err != nil {
		fmt.Println("abi json error: ", err)
	}

	//updateCurrentBalance, err := myabi.Pack("updateCurrentBalance", new(big.Int).SetUint64(uint64(666)))
	//if err != nil {
	//	fmt.Println("abi pack error: ", err)
	//} else {
	//	fmt.Println("abi: ", updateCurrentBalance)
	//}
	//ExecuteCall(updateCurrentBalance)

	//s2 := "222222"
	//addr := bean.Hex2Address(s2)
	//updateBalance, err := myabi.Pack("updateBalance", addr, new(big.Int).SetUint64(uint64(1888)))
	//if err != nil {
	//	fmt.Println("abi pack error: ", err)
	//} else {
	//	fmt.Println("abi: ", updateBalance)
	//}
	//ExecuteCall(updateBalance)

	//s1 := "111111"
	//addr := bean.Hex2Address(s1)
	//
	//balanceOf, err := myabi.Pack("balanceOf", addr)
	//if err != nil {
	//	fmt.Println("abi pack error: ", err)
	//} else {
	//	fmt.Println("abi: ", balanceOf)
	//}
	//ExecuteCall(balanceOf)

	//s2 := "222222"
	//addr := bean.Hex2Address(s2)
	//
	//balanceOf, err := myabi.Pack("balanceOf", addr)
	//if err != nil {
	//	fmt.Println("abi pack error: ", err)
	//} else {
	//	fmt.Println("abi: ", balanceOf)
	//}
	//ExecuteCall(balanceOf)

	//s2 := "222222"
	//addr := bean.Hex2Address(s2)
	//
	//transfer, err := myabi.Pack("transfer", addr, new(big.Int).SetUint64(266))
	//if err != nil {
	//	fmt.Println("abi pack error: ", err)
	//} else {
	//	fmt.Println("abi: ", transfer)
	//}
	//ExecuteCall(transfer)

	s1 := "111111"
	from := accounts.Hex2Address(s1)
	s2 := "222222"
	to := accounts.Hex2Address(s2)
	transferFrom, err := myabi.Pack("transferFrom", from, to, new(big.Int).SetUint64(10))
	if err != nil {
		fmt.Println("abi pack error: ", err)
	} else {
		fmt.Println("abi: ", transferFrom)
	}
	ExecuteCall(transferFrom)
}

func TestThis(t *testing.T) {

}

func TestMain123(t *testing.T) {
	s := vm.GetState()
	db := s.GetDB()
	itr := db.NewIterator()
	for itr.Next() {
		key := itr.Key()
		value := itr.Value()
		fmt.Println("key: ", key)
		fmt.Println()
		fmt.Println("value: ", value)
		fmt.Println()
		account := &accounts.Account{}
		err := json.Unmarshal(value, account)
		if err == nil {
			fmt.Println("v addr: ", account.Address.Hex())
			fmt.Println("v: ", account)
			fmt.Println(account.Storage.Balance)
			continue
		}
		log := &bean.Log{}
		err = json.Unmarshal(value, log)
		if err == nil {
			fmt.Println("log: ", log)
		}
		fmt.Println()
	}
}

func TestHaha(t *testing.T) {
	bb := []byte{230, 136, 145, 101, 114, 116}
	fmt.Println(string(bb))
	ss := "ed1e338910836644d88868b3f7326fad9262abff7bc13dd4d1d7eb51cc42f29a"
	b, _ := hex.DecodeString(ss)
	fmt.Println("result: ", new(big.Int).SetBytes(b))
}

func TestCreate1(t *testing.T) {
	code, err := hex.DecodeString("608060405260043610603f576000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff1680639b22c05d146044575b600080fd5b348015604f57600080fd5b50606c60048036038101908080359060200190929190505050606e565b005b80600081905550505600a165627a7a72305820639f67b470196bb950d918a7bb6f9eca9d635235ea324b7388bad23fabdd01cf0029")
	if err != nil {
		fmt.Println("bad byte code")
	}
	ExecuteCreate(code)
}

func TestCall1(t *testing.T) {
	var mystr = `[
	{
		"constant": false,
		"inputs": [
			{
				"name": "_a",
				"type": "int256"
			}
		],
		"name": "test",
		"outputs": [],
		"payable": false,
		"stateMutability": "nonpayable",
		"type": "function"
	}
]`

	myabi, err := abi.JSON(strings.NewReader(mystr))
	if err != nil {
		fmt.Println("abi json error: ", err)
	}

	test, err := myabi.Pack("test", new(big.Int).SetInt64(9898))
	if err != nil {
		fmt.Println("pack err: ", err)
	}
	fmt.Println("test: ", test)
	//ExecuteCall(test)
}

func TestCreate2(t *testing.T) {
	code, err := hex.DecodeString("608060405260043610610041576000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff1680636291679114610046575b600080fd5b34801561005257600080fd5b506100ad600480360381019080803590602001908201803590602001908080601f01602080910402602001604051908101604052809392919081815260200183838082843782019150505050505091929192905050506100af565b005b6000736705ff829df94abdb22f5dbc7bf8c1637f58d8e190508073ffffffffffffffffffffffffffffffffffffffff168260405180828051906020019080838360005b8381101561010d5780820151818401526020810190506100f2565b50505050905090810190601f16801561013a5780820380516001836020036101000a031916815260200191505b50915050600060405180830381855af49150505050505600a165627a7a72305820761ccd19d982b80dd404c28942b9436e0462acc7696542323acd58c6454d18270029")
	if err != nil {
		fmt.Println("bad byte code")
	}
	ExecuteCreate(code)
}

func TestCall2(t *testing.T) {
	var mystr = `[
	{
		"constant": false,
		"inputs": [
			{
				"name": "_b",
				"type": "bytes"
			}
		],
		"name": "demo",
		"outputs": [],
		"payable": false,
		"stateMutability": "nonpayable",
		"type": "function"
	}
]`

	myabi, err := abi.JSON(strings.NewReader(mystr))
	if err != nil {
		fmt.Println("abi json error: ", err)
	}

	_b := []byte{155, 34, 192, 93, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 38, 170}
	demo, err := myabi.Pack("demo", _b)
	if err != nil {
		fmt.Println("pack err: ", err)
	}
	ExecuteCall(demo)
}