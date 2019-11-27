package solo

import (
	"github.com/drep-project/DREP-Chain/chain"
	"github.com/drep-project/DREP-Chain/crypto/secp256k1"
	"github.com/drep-project/DREP-Chain/crypto/sha3"
	"github.com/drep-project/DREP-Chain/types"
)

type SoloValidator struct {
	pubkey *secp256k1.PublicKey
}

func NewSoloValidator(pubkey *secp256k1.PublicKey) *SoloValidator {
	return &SoloValidator{
		pubkey: pubkey,
	}
}

func (soloValidator *SoloValidator) VerifyHeader(header, parent *types.BlockHeader) error {
	return nil
}

func (soloValidator *SoloValidator) VerifyBody(block *types.Block) error {
	hash := sha3.Keccak256(block.AsSignMessage())
	sig, err := secp256k1.ParseSignature(block.Proof.Evidence)
	if err != nil {
		return err
	}
	if sig.Verify(hash, soloValidator.pubkey) {
		return nil
	} else {
		return ErrCheckSigFail
	}
}

func (soloValidator *SoloValidator) ExecuteBlock(context *chain.BlockExecuteContext) error {
	return AccumulateRewards(soloValidator.pubkey, context.TrieStore, context.GasFee, context.Block.Header.Height)
}
