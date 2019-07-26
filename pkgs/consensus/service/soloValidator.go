package service

import (
	"github.com/drep-project/drep-chain/chain"
	"github.com/drep-project/drep-chain/crypto"
	"github.com/drep-project/drep-chain/crypto/secp256k1"
	"github.com/drep-project/drep-chain/crypto/sha3"
	types2 "github.com/drep-project/drep-chain/pkgs/consensus/types"
	"github.com/drep-project/drep-chain/types"
)

type SoloValidator struct {
	pubkey *secp256k1.PublicKey
	Producers []types2.Producer
}

func (soloValidator *SoloValidator) VerifyHeader(header, parent *types.BlockHeader) error {
	// check multisig
	// leader
	if !soloValidator.isInLocalBp(header.LeaderAddress) {
		return ErrBpNotInList
	}
	// minor
	for _, minor := range header.MinorAddresses {
		if !soloValidator.isInLocalBp(minor) {
			return ErrBpNotInList
		}
	}
	return nil
}

func (soloValidator *SoloValidator)VerifyBody(block *types.Block) error{
	hash := sha3.Keccak256(block.AsSignMessage())
	if block.MultiSig.Sig.Verify(hash,soloValidator.pubkey) {
		return nil
	} else {
		return  ErrMultiSig
	}

}

// isInLocalBp check the specific pubket  is a bp node
func (soloValidator *SoloValidator) isInLocalBp(key crypto.CommonAddress) bool {
	for _, bp := range soloValidator.Producers {
		if crypto.PubKey2Address(bp.Pubkey) == key {
			return true
		}
	}
	return false
}

func (soloValidator *SoloValidator)ExecuteBlock(context *chain.BlockExecuteContext) error{
	return nil
}
