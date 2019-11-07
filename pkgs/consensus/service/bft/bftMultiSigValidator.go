package bft

import (
	"github.com/drep-project/binary"
	"github.com/drep-project/drep-chain/chain"
	"github.com/drep-project/drep-chain/crypto"
	"github.com/drep-project/drep-chain/crypto/secp256k1"
	"github.com/drep-project/drep-chain/crypto/secp256k1/schnorr"
	"github.com/drep-project/drep-chain/crypto/sha3"
	"github.com/drep-project/drep-chain/types"
)

type GetProducers func(uint64,int) ([]*Producer, error)
type GetBlock func(hash *crypto.Hash) (*types.Block, error)

type BlockMultiSigValidator struct {
	getProducers GetProducers
	getBlock     GetBlock
	producerNum int
}

func NewBlockMultiSigValidator(getProducers GetProducers, getBlock GetBlock,producerNum int) *BlockMultiSigValidator {
	return &BlockMultiSigValidator{getProducers, getBlock,producerNum}
}

func (blockMultiSigValidator *BlockMultiSigValidator) VerifyHeader(header, parent *types.BlockHeader) error {
	// check multisig
	// leader
	return nil
}

func (blockMultiSigValidator *BlockMultiSigValidator) VerifyBody(block *types.Block) error {
	participators := []*secp256k1.PublicKey{}
	multiSig := &MultiSignature{}
	err := binary.Unmarshal(block.Proof.Evidence, multiSig)
	if err != nil {
		return err
	}
	parentBlock, err := blockMultiSigValidator.getBlock(&block.Header.PreviousHash)
	if err != nil {
		return err
	}
	producers, err := blockMultiSigValidator.getProducers(parentBlock.Header.Height,blockMultiSigValidator.producerNum)
	if err != nil {
		return err
	}
	for index, val := range multiSig.Bitmap {
		if val == 1 {
			producer := producers[index]
			participators = append(participators, producer.Pubkey)
		}
	}
	msg := block.AsSignMessage()
	sigmaPk := schnorr.CombinePubkeys(participators)

	if !schnorr.Verify(sigmaPk, sha3.Keccak256(msg), multiSig.Sig.R, multiSig.Sig.S) {
		return ErrMultiSig
	}
	return nil
}

func (blockMultiSigValidator *BlockMultiSigValidator) ExecuteBlock(context *chain.BlockExecuteContext) error {
	multiSig := &MultiSignature{}
	parentBlock, err := blockMultiSigValidator.getBlock(&context.Block.Header.PreviousHash)
	if err != nil {
		return err
	}
	producers, err := blockMultiSigValidator.getProducers(parentBlock.Header.Height, blockMultiSigValidator.producerNum)
	if err != nil {
		return err
	}
	err = binary.Unmarshal(context.Block.Proof.Evidence, multiSig)
	if err != nil {
		return nil
	}
	calculator := NewRewardCalculator(context.TrieStore, multiSig, producers, context.GasFee, context.Block.Header.Height)
	return calculator.AccumulateRewards()
}
