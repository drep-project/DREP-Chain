package params

import (
	"github.com/drep-project/DREP-Chain/crypto"
	"math/big"
)

var (
	HoleAddressStr = "0x0000000000000000000000000000000000000000"
)
var (
	HoleAddress = crypto.HexToAddress(HoleAddressStr)
	Preminer    = map[crypto.CommonAddress]*big.Int{
		crypto.HexToAddress("0xaD3dC2D8aedef155eabA42Ab72C1FE480699336c"): CoinFromNumer(10000000000),
	}
)

func CoinFromNumer(number int64) *big.Int {
	return big.NewInt(0).Mul(big.NewInt(Coin), big.NewInt(number))
}
