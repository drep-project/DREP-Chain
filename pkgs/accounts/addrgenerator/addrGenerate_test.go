package addrgenerator

import (
	"fmt"
	"github.com/drep-project/drep-chain/crypto/secp256k1"
	"testing"
)

func Test_Generate(t *testing.T) {
	pri, _ := secp256k1.GeneratePrivateKey(nil)
	generator := &AddrGenerate{
		PrivateKey: pri,
	}
	fmt.Println("Ripple：", generator.ToRipple())
	fmt.Println("Btc：", generator.ToBtc())
	fmt.Println("TEth：", generator.ToEth())
	fmt.Println("Neo：", generator.ToNeo())
	fmt.Println("Dash：", generator.ToDash())
	fmt.Println("Dogecoin：", generator.ToDogecoin())
	fmt.Println("LiteCoin：", generator.ToLiteCoin())
	fmt.Println("Atom：", generator.ToAtom())
	fmt.Println("Tron：", generator.ToTron())
}
