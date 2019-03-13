package types

import "github.com/drep-project/drep-chain/crypto/secp256k1"

type Key struct {
	Pubkey *secp256k1.PublicKey
	PrivKey *secp256k1.PrivateKey
}