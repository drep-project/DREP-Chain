package params

import (
	"github.com/drep-project/drep-chain/crypto"
	"math/big"
)

var (
	HoleAddressStr = "0x0000000000000000000000000000000000000000"
)
var (
	HoleAddress = crypto.String2Address(HoleAddressStr)
	Preminer    = map[crypto.CommonAddress]*big.Int{
		crypto.String2Address("0xec61c03f719a5c214f60719c3f36bb362a202125"): CoinFromNumer(10000000000),
	}
)

func CoinFromNumer(number int64) *big.Int {
	return big.NewInt(0).Mul(big.NewInt(Coin), big.NewInt(number))
}
