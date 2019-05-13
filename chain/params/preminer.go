package params

import (
	"github.com/drep-project/drep-chain/crypto"
	"math/big"
)

var Preminer = map[crypto.CommonAddress]*big.Int {
	crypto.String2Address("0xe91f67944ec2f7223bf6d0272557a5b13ecc1f28"):CoinFromNumer(10000000000),
}

func CoinFromNumer(number int64) *big.Int{
	return big.NewInt(0).Mul(big.NewInt(Coin), big.NewInt(number))
}