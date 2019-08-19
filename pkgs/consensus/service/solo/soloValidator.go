package solo

import (
	"github.com/drep-project/drep-chain/chain"
	"github.com/drep-project/drep-chain/crypto/secp256k1"
	"github.com/drep-project/drep-chain/crypto/sha3"
	"github.com/drep-project/drep-chain/types"
)

type SoloValidator struct {
	consensus *SoloConsensus
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
	sig, err := secp256k1.ParseSignature(block.Proof)
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
	return soloValidator.consensus.AccumulateRewards(context.Db, context.GasFee)
}
