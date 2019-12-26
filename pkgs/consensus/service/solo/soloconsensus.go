package solo

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/drep-project/DREP-Chain/blockmgr"
	"github.com/drep-project/DREP-Chain/chain"
	"github.com/drep-project/DREP-Chain/chain/store"
	"github.com/drep-project/DREP-Chain/crypto"
	"github.com/drep-project/DREP-Chain/crypto/secp256k1"
	"github.com/drep-project/DREP-Chain/crypto/sha3"
	"github.com/drep-project/DREP-Chain/database"
	"github.com/drep-project/DREP-Chain/params"
	consensusTypes "github.com/drep-project/DREP-Chain/pkgs/consensus/types"
	"github.com/drep-project/DREP-Chain/types"
	"math/big"
	"time"
)

type SoloConsensus struct {
	CoinBase       crypto.CommonAddress
	PrivKey        *secp256k1.PrivateKey
	Pubkey         *secp256k1.PublicKey
	blockGenerator blockmgr.IBlockBlockGenerator
	ChainService   chain.ChainServiceInterface
	DbService      *database.DatabaseService
	config         *SoloConfig
}

func NewSoloConsensus(
	chainService chain.ChainServiceInterface,
	blockGenerator blockmgr.IBlockBlockGenerator,
	myPk *secp256k1.PublicKey,
	dbService *database.DatabaseService,
	config *SoloConfig) *SoloConsensus {
	return &SoloConsensus{
		blockGenerator: blockGenerator,
		ChainService:   chainService,
		DbService:      dbService,
		Pubkey:         myPk,
		config:         config,
	}
}

func (soloConsensus *SoloConsensus) Run(privKey *secp256k1.PrivateKey) (*types.Block, error) {
	soloConsensus.CoinBase = crypto.PubkeyToAddress(privKey.PubKey())
	soloConsensus.PrivKey = privKey
	//区块生成 共识 奖励 验证 完成
	log.Trace("node leader finishes process consensus")

	trieStore, err := store.TrieStoreFromStore(soloConsensus.DbService.LevelDb(), soloConsensus.ChainService.BestChain().Tip().StateRoot)
	if err != nil {
		return nil, err
	}

	fmt.Println("run 0", time.Now())
	block, gasFee, err := soloConsensus.blockGenerator.GenerateTemplate(trieStore, soloConsensus.CoinBase, soloConsensus.config.BlockInterval)
	if err != nil {
		return nil, err
	}

	fmt.Println("run 1", time.Now())
	sig, err := soloConsensus.PrivKey.Sign(sha3.Keccak256(block.AsSignMessage()))
	if err != nil {
		log.Error("sign block error")
		return nil, errors.New("sign block error")
	}

	fmt.Println("run 2", time.Now())
	block.Proof = types.Proof{consensusTypes.Solo, sig.Serialize()}
	err = AccumulateRewards(soloConsensus.Pubkey, trieStore, gasFee, block.Header.Height)
	if err != nil {
		return nil, err
	}

	fmt.Println("run 3", time.Now())
	block.Header.StateRoot = trieStore.GetStateRoot()
	//verify
	if err := soloConsensus.verify(block); err != nil {
		return nil, err
	}

	fmt.Println("run 4", time.Now())
	return block, nil
}

func (soloConsensus *SoloConsensus) verify(block *types.Block) error {
	parent, err := soloConsensus.ChainService.GetBlockHeaderByHeight(block.Header.Height - 1)
	if err != nil {
		return err
	}

	dbstore := &chain.ChainStore{soloConsensus.DbService.LevelDb()}
	trieStore, err := store.TrieStoreFromStore(soloConsensus.DbService.LevelDb(), parent.StateRoot)
	if err != nil {
		return err
	}
	gp := new(chain.GasPool).AddGas(block.Header.GasLimit.Uint64())
	//process transaction

	context := chain.NewBlockExecuteContext(trieStore, gp, dbstore, block)
	validators := soloConsensus.ChainService.BlockValidator()
	for _, validator := range validators {
		err = validator.ExecuteBlock(context)
		if err != nil {
			log.WithField("ExecuteBlock", err).Debug("multySigVerify")
			return err
		}
	}
	stateRoot := trieStore.GetStateRoot()

	if block.Header.GasUsed.Cmp(context.GasUsed) == 0 {
		if !bytes.Equal(block.Header.StateRoot, stateRoot) {
			if !trieStore.RecoverTrie(soloConsensus.ChainService.GetCurrentHeader().StateRoot) {
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

func (soloConsensus *SoloConsensus) ReceiveMsg(peer *consensusTypes.PeerInfo, t uint64, buf []byte) {
}

// AccumulateRewards credits,The leader gets half of the reward and other ,Other participants get the average of the other half
func AccumulateRewards(pubkey *secp256k1.PublicKey, trieStore store.StoreInterface, totalGasBalance *big.Int, height uint64) error {
	soloAddr := crypto.PubkeyToAddress(pubkey)
	err := trieStore.AddBalance(&soloAddr, height, totalGasBalance)
	if err != nil {
		return err
	}
	err = trieStore.AddBalance(&soloAddr, height, params.CoinFromNumer(1000))
	if err != nil {
		return err
	}
	return nil
}
