package solo

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"

	"github.com/drep-project/DREP-Chain/blockmgr"
	"github.com/drep-project/DREP-Chain/chain"
	"github.com/drep-project/DREP-Chain/crypto"
	"github.com/drep-project/DREP-Chain/crypto/secp256k1"
	"github.com/drep-project/DREP-Chain/crypto/sha3"
	"github.com/drep-project/DREP-Chain/database"
	"github.com/drep-project/DREP-Chain/network/p2p"
	"github.com/drep-project/DREP-Chain/params"
	consensusTypes "github.com/drep-project/DREP-Chain/pkgs/consensus/types"
	"github.com/drep-project/DREP-Chain/types"
)

type SoloConsensus struct {
	CoinBase       crypto.CommonAddress
	PrivKey        *secp256k1.PrivateKey
	Pubkey         *secp256k1.PublicKey
	blockGenerator blockmgr.IBlockBlockGenerator
	ChainService   chain.ChainServiceInterface
	DbService      *database.DatabaseService
}

func NewSoloConsensus(
	chainService chain.ChainServiceInterface,
	blockGenerator blockmgr.IBlockBlockGenerator,
	producer consensusTypes.Producer,
	dbService *database.DatabaseService) *SoloConsensus {
	return &SoloConsensus{
		blockGenerator: blockGenerator,
		ChainService:   chainService,
		DbService:      dbService,
		Pubkey:         producer.Pubkey,
	}
}

func (soloConsensus *SoloConsensus) Run(privKey *secp256k1.PrivateKey) (*types.Block, error) {
	soloConsensus.CoinBase = crypto.PubkeyToAddress(privKey.PubKey())
	soloConsensus.PrivKey = privKey
	//区块生成 共识 奖励 验证 完成
	log.Trace("node leader finishes process consensus")

	db := soloConsensus.DbService.BeginTransaction(false)
	block, gasFee, err := soloConsensus.blockGenerator.GenerateTemplate(db, soloConsensus.CoinBase)
	if err != nil {
		return nil, err
	}
	sig, err := soloConsensus.PrivKey.Sign(sha3.Keccak256(block.AsSignMessage()))
	if err != nil {
		log.Error("sign block error")
		return nil, errors.New("sign block error")
	}
	block.Proof = types.Proof{consensusTypes.Solo, sig.Serialize()}
	err = AccumulateRewards(soloConsensus.Pubkey, db, gasFee)
	if err != nil {
		return nil, err
	}
	db.Commit()
	block.Header.StateRoot = db.GetStateRoot()
	//verify
	if err := soloConsensus.verify(block); err != nil {
		return nil, err
	}
	return block, nil
}

func (soloConsensus *SoloConsensus) verify(block *types.Block) error {
	db := soloConsensus.DbService.BeginTransaction(false)
	gp := new(chain.GasPool).AddGas(block.Header.GasLimit.Uint64())
	//process transaction
	context := &chain.BlockExecuteContext{
		Db:      db,
		Block:   block,
		Gp:      gp,
		GasUsed: new(big.Int),
		GasFee:  new(big.Int),
	}
	for _, validator := range soloConsensus.ChainService.BlockValidator() {
		err := validator.ExecuteBlock(context)
		if err != nil {
			log.WithField("ExecuteBlock", err).Debug("multySigVerify")
			return err
		}
	}
	db.Commit()
	if block.Header.GasUsed.Cmp(context.GasUsed) == 0 {
		stateRoot := db.GetStateRoot()
		if !bytes.Equal(block.Header.StateRoot, stateRoot) {
			if !db.RecoverTrie(soloConsensus.ChainService.GetCurrentHeader().StateRoot) {
				log.Fatal("root not equal and recover trie err")
			}
			log.Error("rootcmd root !=")
			return fmt.Errorf("state root not equal")
		}
	} else {
		log.WithField("gasUsed", context.GasUsed).Debug("multySigVerify")
		return ErrGasUsed
	}
	return nil
}

func (soloConsensus *SoloConsensus) ReceiveMsg(peer *consensusTypes.PeerInfo, rw p2p.MsgReadWriter) error {
	return nil
}

// AccumulateRewards credits,The leader gets half of the reward and other ,Other participants get the average of the other half
func AccumulateRewards(pk *secp256k1.PublicKey, db *database.Database, totalGasBalance *big.Int) error {
	soloAddr := crypto.PubkeyToAddress(pk)
	db.AddBalance(&soloAddr, totalGasBalance)
	db.AddBalance(&soloAddr, params.CoinFromNumer(1000))
	return nil
}
