package solo

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/drep-project/drep-chain/blockmgr"
	"github.com/drep-project/drep-chain/chain"
	"github.com/drep-project/drep-chain/crypto"
	"github.com/drep-project/drep-chain/crypto/secp256k1"
	"github.com/drep-project/drep-chain/crypto/sha3"
	"github.com/drep-project/drep-chain/database"
	"github.com/drep-project/drep-chain/network/p2p"
	consensusTypes "github.com/drep-project/drep-chain/pkgs/consensus/types"
	"github.com/drep-project/drep-chain/types"
	"math/big"
)

type SoloConsensus struct {
	CoinBase     crypto.CommonAddress
	PrivKey      *secp256k1.PrivateKey
	BlockMgr     *blockmgr.BlockMgr
	ChainService chain.ChainServiceInterface
	DbService    *database.DatabaseService
}

func NewSoloConsensus(chainService chain.ChainServiceInterface, blockMgr *blockmgr.BlockMgr, dbService *database.DatabaseService, privKey *secp256k1.PrivateKey) *SoloConsensus {
	return &SoloConsensus{
		CoinBase:     crypto.PubKey2Address(privKey.PubKey()),
		PrivKey:      privKey,
		BlockMgr:     blockMgr,
		ChainService: chainService,
		DbService:    dbService,
	}
}

func (soloConsensus *SoloConsensus) Run() (*types.Block, error) {
	//区块生成 共识 奖励 验证 完成
	log.Trace("node leader finishes process consensus")

	db := soloConsensus.DbService.BeginTransaction(false)
	block, gasFee, err := soloConsensus.BlockMgr.GenerateBlock(db, soloConsensus.CoinBase)
	if err != nil {
		return nil, err
	}
	sig, err := soloConsensus.PrivKey.Sign(sha3.Keccak256(block.AsSignMessage()))
	if err != nil {
		log.Error("sign block error")
		return nil, errors.New("sign block error")
	}
	block.Proof = sig.Serialize()
	err = soloConsensus.ChainService.AccumulateRewards(db, block, gasFee)
	if err != nil {
		return nil, err
	}

	db.Commit()
	block.Header.StateRoot = db.GetStateRoot()

	//verify
	db = soloConsensus.DbService.BeginTransaction(false)
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
		_, _, _, err := validator.ExecuteBlock(context)
		if err != nil {
			log.WithField("ExecuteBlock", err).Debug("multySigVerify")
			return nil, err
		}
	}
	err = soloConsensus.ChainService.AccumulateRewards(db, block, gasFee)
	if err != nil {
		log.WithField("AccumulateRewards", err).Debug("multySigVerify")
		return nil, err
	}

	db.Commit()
	if block.Header.GasUsed.Cmp(context.GasUsed) == 0 {
		stateRoot := db.GetStateRoot()
		if !bytes.Equal(block.Header.StateRoot, stateRoot) {
			if !db.RecoverTrie(soloConsensus.ChainService.GetCurrentHeader().StateRoot){
				log.Fatal("root not equal and recover trie err")
			}

			log.Error("rootcmd root !=")
			return nil, fmt.Errorf("state root not equal")
		}
	} else {
		log.WithField("gasUsed", context.GasUsed).Debug("multySigVerify")
		return nil, err
	}
	return block, nil
}

func (soloConsensus *SoloConsensus) ReceiveMsg(peer *consensusTypes.PeerInfo, rw p2p.MsgReadWriter) error {
	return nil
}
