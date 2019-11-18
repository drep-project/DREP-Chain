package addrgenerator

import (
	"fmt"
	"github.com/drep-project/DREP-Chain/crypto/secp256k1"
	"testing"
)

func Test_Generate(t *testing.T) {
	pri, _ := secp256k1.GeneratePrivateKey(nil)
	generator := &AddrGenerate{
		PrivateKey: pri,
	}
	fmt.Println("瑞波：", generator.ToRipple())
	fmt.Println("比特币：", generator.ToBtc())
	fmt.Println("以太坊：", generator.ToEth())
	fmt.Println("小蚁：", generator.ToNeo())
	fmt.Println("达世：", generator.ToDash())
	fmt.Println("狗狗币：", generator.ToDogecoin())
	fmt.Println("莱特币：", generator.ToLiteCoin())
	fmt.Println("Cosmos：", generator.ToAtom())
	fmt.Println("Tron：", generator.ToTron())
}
