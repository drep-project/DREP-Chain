package service

import (
	"errors"
	"fmt"
	"github.com/drep-project/dlog"
	"github.com/drep-project/drep-chain/app"
	chainTypes "github.com/drep-project/drep-chain/chain/types"
	"github.com/drep-project/drep-chain/crypto"
	"github.com/drep-project/drep-chain/crypto/secp256k1"
	"github.com/drep-project/drep-chain/crypto/secp256k1/schnorr"
	"github.com/drep-project/drep-chain/crypto/sha3"
	"github.com/drep-project/drep-chain/pkgs/evm/vm"
	"math/big"

	"bytes"
	"encoding/hex"
	"github.com/drep-project/binary"
	"net/http"
	"net/url"
	"strconv"
)

var (
	childTrans []*chainTypes.Transaction
	errBalance = errors.New("not enough balance")
)

func (chainService *ChainService) ValidateBlock(block *chainTypes.Block, skipCheckSig bool) bool {
	if !skipCheckSig {
		if !chainService.ValidateMultiSig(block) {
			dlog.Error("failed to validate block multiSig")
			return false
		}
	}
	var result error
	_ = chainService.DatabaseService.Transaction(func() error {
		_, result = chainService.executeBlock(block)
		return errors.New("just not commit")
	})
	if result != nil {
		return false
	}
	return true
}

func (chainService *ChainService) ValidateTransaction(tx *chainTypes.Transaction) error {
	var result error
	_ = chainService.DatabaseService.Transaction(func() error {
		_, _, result = chainService.executeTransaction(tx)
		return errors.New("just not commit")
	})
	return result
}


func (chainService *ChainService) ValidateTransactionsInBlock(blockdata *chainTypes.BlockData) error {
	var result error
	_ = chainService.DatabaseService.Transaction(func() error {
		_, result = chainService.executeTransactionInBlock(blockdata)
		return errors.New("just not commit")
	})
	return result
}

func (chainService *ChainService) ValidateMultiSig(b *chainTypes.Block) bool {
	participators := []*secp256k1.PublicKey{}
	for index, val := range b.MultiSig.Bitmap {
		if val == 1 {
			producer := chainService.Config.Producers[index]
			participators = append(participators, producer.Public)
		}
	}
	msg := b.ToMessage()
	sigmaPk := schnorr.CombinePubkeys(participators)
	return schnorr.Verify(sigmaPk, sha3.Hash256(msg), b.MultiSig.Sig.R, b.MultiSig.Sig.S)
}

func (chainService *ChainService) ExecuteBlock(b *chainTypes.Block) (gasUsed *big.Int, err error) {
	err = chainService.DatabaseService.Transaction(func() error {
		gasUsed, err = chainService.executeBlock(b)
		return err
	})
	return
}

func (chainService *ChainService) executeBlock(b *chainTypes.Block) (*big.Int, error) {
	if b == nil || b.Header == nil {
		return nil, errors.New("error block nil or header nil")
	}
	total := big.NewInt(0)
	if b.Data == nil {
		return total, nil
	}

	gasFee, err := chainService.executeTransactionInBlock(b.Data)
	if err != nil {
		return nil, err
	}
	total.Add(total, gasFee)

	stateRoot := chainService.DatabaseService.GetStateRoot()
	if bytes.Equal(b.Header.StateRoot, stateRoot) {
		dlog.Debug("matched ", "BlockStateRoot", hex.EncodeToString(b.Header.StateRoot), "CalcStateRoot", hex.EncodeToString(stateRoot))
		chainService.accumulateRewards(b,total)
		chainService.preSync(b)
		chainService.doSync(b.Header.Height)
		return total, nil
	} else {
		return nil, fmt.Errorf("%s not matched %s", hex.EncodeToString(b.Header.StateRoot), " vs ", hex.EncodeToString(stateRoot))
	}
}

func (chainService *ChainService) executeTransactionInBlock(data *chainTypes.BlockData) (*big.Int, error) {
	total := big.NewInt(0)
	for _, t := range data.TxList {
		_, gasFee, err := chainService.executeTransaction(t)
		if err != nil {
			return nil, err
		}
		if gasFee != nil {
			total.Add(total, gasFee)
		}
	}
	return total, nil
}

func (chainService *ChainService) preSync(block *chainTypes.Block) {
	if !chainService.isRelay && chainService.chainId != chainService.RootChain() {
		return
	}
	if childTrans == nil {
		childTrans = make([]*chainTypes.Transaction, 0)
	}
	childTrans = append(childTrans, block.Data.TxList...)
}

func (chainService *ChainService) doSync(height int64) {
	if !chainService.isRelay || chainService.chainId == chainService.RootChain() || height%2 != 0 || height == 0 {
		return
	}
	cct := &chainTypes.CrossChainTransaction{
		ChainId:   chainService.chainId,
		StateRoot: chainService.DatabaseService.GetStateRoot(),
		Trans:     childTrans,
	}
	data, err := binary.Marshal(cct)
	if err != nil {
		return
	}
	values := url.Values{}
	values.Add("data", string(data))
	body := values.Encode()
	urlStr := "http://localhost:" + strconv.Itoa(chainService.Config.RemotePort) + "/SyncChildChain?" + body
	http.Get(urlStr)
	childTrans = nil
}

func (chainService *ChainService) executeTransaction(tx *chainTypes.Transaction) (*big.Int, *big.Int, error) {
	to := tx.To()
	nounce := tx.Nonce()
	amount := tx.Amount()
	fromAccount := tx.From()
	gasPrice := tx.GasPrice()
	gasLimit := tx.GasLimit()

	if tx.Sig != nil {
		isValid, err := chainService.verify(tx)
		if err != nil {
			return nil, nil, err
		}
		if !isValid {
			return nil, nil, errors.New("signature not validate")
		}
	}else{
		return nil, nil, errors.New("must assign a signature for transaction")
	}

	originBalance := chainService.DatabaseService.GetBalance(fromAccount, true)
	err := chainService.checkNonce(fromAccount, nounce)
	if err != nil {
		return  nil, nil, err
	}

	var gasUsed *big.Int
	var gasFee *big.Int
	switch tx.Type() {
	case chainTypes.TransferType:
		err = chainService.checkBalance(gasLimit, gasPrice, originBalance, chainTypes.GasTable[chainTypes.TransferType], nil)
		if err != nil {
			return  nil, nil, err
		}
		gasUsed, gasFee, err = chainService.executeTransferTransaction(tx, fromAccount, to, originBalance, gasPrice, gasLimit, amount, tx.ChainId())
	case chainTypes.CreateContractType:
		err = chainService.checkBalance(gasLimit, gasPrice, originBalance,nil, chainTypes.GasTable[chainTypes.CreateContractType])
		if err != nil {
			return  nil, nil, err
		}
		gasUsed, gasFee, err = chainService.executeCreateContractTransaction(tx, fromAccount, to, originBalance, gasPrice, gasLimit, amount, tx.ChainId())
	case chainTypes.CallContractType:
		err = chainService.checkBalance(gasLimit, gasPrice, originBalance,nil, chainTypes.GasTable[chainTypes.CallContractType])
		if err != nil {
			return  nil, nil, err
		}
		gasUsed, gasFee, err = chainService.executeCallContractTransaction(tx, fromAccount, to, originBalance, gasPrice, gasLimit, amount, tx.ChainId())
	}
	if err != nil {
		dlog.Error("executeTransaction transaction error", "reason", err)
	}
	chainService.DatabaseService.PutNonce(fromAccount, nounce+1, true)
	return gasUsed, gasFee, nil
}

func (chainService *ChainService) verify(tx *chainTypes.Transaction) (bool, error){
	pk, _, err := secp256k1.RecoverCompact(tx.Sig, tx.TxHash().Bytes())
	if err != nil {
		return false, err
	}
	sig := secp256k1.RecoverSig(tx.Sig)
	isValid := sig.Verify(tx.TxHash().Bytes(), pk)
	return isValid, nil
}

func (chainService *ChainService) executeTransferTransaction(t *chainTypes.Transaction, fromAccount, to *crypto.CommonAddress, balance, gasPrice, gasLimit, amount *big.Int, chainId app.ChainIdType) (*big.Int, *big.Int, error) {
	gasUsed := new(big.Int).Set(chainTypes.GasTable[chainTypes.TransferType])
	gasFee := new(big.Int).Mul(gasUsed, gasPrice)
	leftBalance, gasFee := chainService.deduct(chainId, balance, gasFee)  //sub gas fee
	if leftBalance.Cmp(amount) >= 0 {				//sub transfer amount
		leftBalance = new(big.Int).Sub(leftBalance, amount)
		balanceTo := chainService.DatabaseService.GetBalance(to, true)
		balanceTo = new(big.Int).Add(balanceTo, amount)
		chainService.DatabaseService.PutBalance(fromAccount, leftBalance, true)
		chainService.DatabaseService.PutBalance(to, balanceTo, true)
	} else {
		return gasUsed, gasFee, errBalance
	}
	return gasUsed, gasFee, nil
}

func (chainService *ChainService) executeCreateContractTransaction(t *chainTypes.Transaction, fromAccount, to *crypto.CommonAddress, balance, gasPrice, gasLimit, amount *big.Int, chainId app.ChainIdType) (*big.Int, *big.Int, error) {
	evm := vm.NewEVM(chainService.DatabaseService)
	refundGas, err := chainService.VmService.ApplyTransaction(evm, t)
	balance = chainService.DatabaseService.GetBalance(fromAccount, true)
	gasUsed := new(big.Int).Sub(gasLimit, new(big.Int).SetUint64(refundGas))
	gasFee  := new(big.Int).Mul(gasUsed, gasPrice)
	leftBalance, gasFee := chainService.deduct(chainId, balance, gasFee)
	if leftBalance.Sign() >= 0{
		chainService.DatabaseService.PutBalance(fromAccount, leftBalance, true)
		return gasUsed, gasFee, err
	}else{
		return gasUsed, gasFee, errBalance
	}
}

func (chainService *ChainService) executeCallContractTransaction(t *chainTypes.Transaction, fromAccount, to *crypto.CommonAddress, balance, gasPrice, gasLimit, amount *big.Int, chainId app.ChainIdType) (*big.Int, *big.Int, error) {
	evm := vm.NewEVM(chainService.DatabaseService)
	returnGas, err := chainService.VmService.ApplyTransaction(evm, t)
	balance = chainService.DatabaseService.GetBalance(fromAccount, true)
	gasUsed := new(big.Int).Sub(gasLimit, new(big.Int).SetUint64(returnGas))
	gasFee  := new(big.Int).Mul(gasUsed, gasPrice)
	leftBalance, gasFee := chainService.deduct(t.ChainId(), balance, gasFee)
	if leftBalance.Sign() >= 0{
		chainService.DatabaseService.PutBalance(fromAccount, leftBalance, true)
		return gasUsed, gasFee, err
	}else{
		return gasUsed, gasFee, errBalance
	}
}

func  (chainService *ChainService) checkNonce(fromAccount *crypto.CommonAddress, nounce int64) error{
	nonce := chainService.DatabaseService.GetNonce(fromAccount, true)
	if nonce > nounce {
		return errors.New("error nounce")
	}
	return nil
}

func (chainService *ChainService) checkBalance(gaslimit, gasPrice, balance, gasFloor, gasCap *big.Int) error {
	if gasFloor != nil {
		amountFloor := new(big.Int).Mul(gasFloor, gasPrice)
		if gaslimit.Cmp(gasFloor) < 0 || amountFloor.Cmp(balance) > 0 {
			return errors.New("not enough gas")
		}
	}
	if gasCap != nil {
		amountCap := new(big.Int).Mul(gasCap, gasPrice)
		if amountCap.Cmp(balance) > 0 {
			return errors.New("too much gaslimit")
		}
	}
	return nil
}

func (chainService *ChainService) deduct(chainId app.ChainIdType, balance, gasFee *big.Int) (leftBalance, actualFee *big.Int) {
	leftBalance = new(big.Int).Sub(balance, gasFee)
	actualFee = new(big.Int).Set(gasFee)
	if leftBalance.Sign() < 0 {
		actualFee = new(big.Int).Set(balance)
		leftBalance = new(big.Int)
	}
	return leftBalance, actualFee
}














//func (chainService *ChainService) executeCrossChainTransaction(t *chainTypes.Transaction) (gasUsed *big.Int, gasFee *big.Int) {
//    var (
//        can bool
//        addr crypto.CommonAddress
//        balance, gasPrice *big.Int
//    )
//
//    gasUsed, gasFee = new(big.Int), new(big.Int)
//    can, addr,  _, _, gasPrice = chainService.canExecute(t, nil, CrossChainGas)
//    if !can {
//        return new(big.Int), new(big.Int)
//    }
//
//    cct := &chainTypes.CrossChainTransaction{}
//    err := json.Unmarshal(t.Data.Data, cct)
//    if err != nil {
//        fmt.Println("err: ", err)
//        return new(big.Int), new(big.Int)
//    }
//
//    gasSum := new(big.Int)
//    for _, tx := range cct.Trans {
//       if tx.Data.Type == CrossChainType {
//           continue
//       }
//       g, _ := chainService.executeTransaction(tx)
//       gasSum = new(big.Int).Add(gasSum, g)
//    }
//
//    if !bytes.Equal(chainService.databaseService.GetStateRoot(), cct.StateRoot) {
//       //subDt.Discard()
//    } else {
//        amountSum := new(big.Int).Mul(gasSum, gasPrice)
//        balance = chainService.databaseService.GetBalance(addr, t.Data.ChainId, true)
//        if balance.Cmp(amountSum) >= 0 {
//            gasUsed = new(big.Int).Set(gasSum)
//            gasFee = new(big.Int).Set(amountSum)
//            _, gasFee = chainService.deduct(addr, t.Data.ChainId, balance, gasFee)
//            //subDt.Commit()
//        } else {
//            //subDt.Discard()
//        }
//    }
//    return
//}

//func preExecuteCrossChainTransaction(dt database.Transactional, t *chainTypes.Transaction) (gasUsed, gasFee *big.Int) {
//    var (
//        can bool
//        addr crypto.CommonAddress
//        balance, gasPrice *big.Int
//    )
//
//    gasUsed, gasFee = new(big.Int), new(big.Int)
//    subDt := dt.BeginTransaction()
//    can, addr,  _, _, gasPrice = canExecute(subDt, t, nil, CrossChainGas)
//    if !can {
//        return new(big.Int), new(big.Int)
//    }
//
//    cct := &chainTypes.CrossChainTransaction{}
//    err := json.Unmarshal(t.Data.Data, &cct)
//    if err != nil {
//        return new(big.Int), new(big.Int)
//    }
//
//    gasSum := new(big.Int)
//    for _, tx := range cct.Trans {
//        if tx.Data.Type == CrossChainType {
//            continue
//        }
//        g, _ := executeTransaction(subDt, tx)
//        gasSum = new(big.Int).Add(gasSum, g)
//    }
//
//    cct.StateRoot = subDt.GetChainStateRoot(database.ChildCHAIN)
//    t.Data.Data, _ = json.Marshal(cct)
//
//    amountSum := new(big.Int).Mul(gasSum, gasPrice)
//    balance = database.GetBalance(addr, t.Data.ChainId)
//    if balance.Cmp(amountSum) >= 0 {
//        gasUsed = new(big.Int).Set(gasSum)
//        gasFee = new(big.Int).Set(amountSum)
//        _, gasFee = deduct(subDt, addr, t.Data.ChainId, balance, gasFee)
//        subDt.Commit()
//    } else {
//        subDt.Discard()
//    }
//
//    return
//}
