package service

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/drep-project/dlog"
	"github.com/drep-project/drep-chain/chain/params"
	chainTypes "github.com/drep-project/drep-chain/chain/types"
	"github.com/drep-project/drep-chain/crypto/secp256k1"
	"github.com/drep-project/drep-chain/crypto/secp256k1/schnorr"
	"github.com/drep-project/drep-chain/crypto/sha3"
	"math/big"
)

func (chainService *ChainService) VerifyHeader(header, parent *chainTypes.BlockHeader) error {
	if header.Timestamp <= parent.Timestamp{
		return errors.New("timestamp equals parent's")
	}
	// Verify that the gas limit is <= 2^63-1
	cap := uint64(0x7fffffffffffffff)
	if header.GasLimit.Uint64() > cap {
		return fmt.Errorf("invalid gasLimit: have %v, max %v", header.GasLimit, cap)
	}
	// Verify that the gasUsed is <= gasLimit
	if header.GasUsed.Uint64() > header.GasLimit.Uint64() {
		return fmt.Errorf("invalid gasUsed: have %d, gasLimit %d", header.GasUsed, header.GasLimit)
	}

	//TODO Verify that the gas limit remains within allowed bounds
	nextGasLimit := chainService.CalcGasLimit(parent, params.MinGasLimit, params.MaxGasLimit)
	if nextGasLimit.Cmp(&header.GasLimit) != 0 {
		return fmt.Errorf("invalid gas limit: have %d, want %d += %d", header.GasLimit, parent.GasLimit, nextGasLimit)
	}
	// Verify that the block number is parent's +1
	if  header.Height - parent.Height != 1 {
		return errors.New("invalid block number")
	}

	return nil
}

func (chainService *ChainService) ValidateBody(block *chainTypes.Block) error {
	// Header validity is known at this point, check the uncles and transactions
	header := block.Header
	if hash := chainService.deriveMerkleRoot(block.Data.TxList); !bytes.Equal(hash , header.TxRoot) {
		return fmt.Errorf("transaction root hash mismatch: have %x, want %x", hash, header.TxRoot)
	}
	return nil
}

func (chainService *ChainService) ValidateState(block *chainTypes.Block) error {
	stateRoot := chainService.DatabaseService.GetStateRoot()
	if bytes.Equal(block.Header.StateRoot, stateRoot) {
		dlog.Debug("matched ", "BlockStateRoot", hex.EncodeToString(block.Header.StateRoot), "CalcStateRoot", hex.EncodeToString(stateRoot))
		return nil
	} else {
		return fmt.Errorf("%s not matched %s", hex.EncodeToString(block.Header.StateRoot), hex.EncodeToString(stateRoot))
	}
}

func (chainService *ChainService) ValidateMultiSig(b *chainTypes.Block, skipCheckSig bool) bool {
	if skipCheckSig {  //just for solo
		return true
	}
	participators := []*secp256k1.PublicKey{}
	for index, val := range b.MultiSig.Bitmap {
		if val == 1 {
			producer := chainService.Config.Producers[index]
			participators = append(participators, producer.Pubkey)
		}
	}
	msg := b.ToMessage()
	sigmaPk := schnorr.CombinePubkeys(participators)
	return schnorr.Verify(sigmaPk, sha3.Hash256(msg), b.MultiSig.Sig.R, b.MultiSig.Sig.S)
}

func (chainService *ChainService) executeTransactionInBlock(block *chainTypes.Block, gp *GasPool) (*big.Int, error) {
	total := big.NewInt(0)
	if len(block.Data.TxList)  <0 {
		return new (big.Int), nil
	}
	for _, t := range block.Data.TxList {
		_, gasFee, err := chainService.executeTransaction(t, gp, block.Header)
		if err != nil {
			return nil, err
			//dlog.Debug("execute transaction fail", "txhash", t.Data, "reason", err.Error())
		}
		if gasFee != nil {
			total.Add(total, gasFee)
		}
	}
	return total, nil
}


