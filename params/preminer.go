package params

import (
	"github.com/drep-project/drep-chain/crypto"
	"math/big"
)

var (
	HoleAddressStr = "0x0000000000000000000000000000000000000000"
)
var (
	HoleAddress = crypto.HexToAddress(HoleAddressStr)
)

func CoinFromNumer(number int64) *big.Int {
	return big.NewInt(0).Mul(big.NewInt(Coin), big.NewInt(number))
}
