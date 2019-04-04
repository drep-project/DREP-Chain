package service

import (
	"errors"
	"fmt"
	"github.com/drep-project/dlog"
	"github.com/drep-project/drep-chain/app"
	chainTypes "github.com/drep-project/drep-chain/chain/types"
	"github.com/drep-project/drep-chain/common"
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
	errBalance = errors.New("no enough blance")
)

func (chainService *ChainService) ValidateBlock(b *chainTypes.Block) bool {
	var result error
	_ = chainService.DatabaseService.Transaction(func() error {
		_, result = chainService.executeBlock(b)
		return errors.New("just not commit")
	})
	if result != nil {
		return false
	}
	return true
}

func (chainService *ChainService) ValidateTransaction(tx *chainTypes.Transaction) (result *chainTypes.ExecuteReuslt) {
	_ = chainService.DatabaseService.Transaction(func() error {
		result = chainService.executeTransaction(tx)
		return errors.New("just not commit")
	})
	return result
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

	for _, tx := range b.Data.TxList {
		result := chainService.executeTransaction(tx)
		if result.Err != nil {
			return nil, result.Err
		}
		if result.GasFee != nil {
			total.Add(total, result.GasFee)
		}
	}

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
	cct := &chainTypes.CrossChainAction{
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

//TODO we can use tx handler but not switch if more tx types
func (chainService *ChainService) executeTransaction(tx *chainTypes.Transaction) *chainTypes.ExecuteReuslt {
	var result = &chainTypes.ExecuteReuslt{}
	nounce := tx.Nonce()
	fromAccount := tx.From()
	gasPrice := tx.GasPrice()
	gasLimit := tx.GasLimit()
	if !chainService.DatabaseService.ExistAccount(fromAccount, true) {
		result.Err = errors.New("the tx from account is not exist")
		return result
	}
	originBalance, err := chainService.DatabaseService.GetBalance(fromAccount, true)
	if err!=nil {
		result.Err = err
		return result
	}
	err = chainService.checkNonce(fromAccount, nounce)
	if err != nil {
		result.Err = err
		return result
	}

	switch tx.Type() {
	case chainTypes.RegisterAccountType:
		action := &chainTypes.RegisterAccountAction{}
		err = binary.Unmarshal(tx.GetData(), action)
		if err != nil{
			result.Err = err
			return result
		}
		//TODO check.name
		if chainService.DatabaseService.ExistAccount(action.Name, true) {
			result.Err = errors.New("account exists")
			return result
		}
		err = chainService.checkBalance(gasLimit, gasPrice, originBalance, chainTypes.GasTable[chainTypes.RegisterAccountType], nil)
		if err != nil {
			result.Err = err
			return result
		}

		result = chainService.executeRegisterAccountTransaction(action, fromAccount, originBalance, gasPrice, gasLimit, tx.ChainId())
	case chainTypes.RegisterMinerType:
		//TODO how to control add miner right
		panic(errors.New("unpupport add miners runtime"))
		action := &chainTypes.RegisterMinerAction{}
		err = binary.Unmarshal(tx.GetData(), action)
		if err != nil{
			result.Err = err
			return result
		}
		if !chainService.DatabaseService.ExistAccount(action.MinerAccount, true) {
			result.Err = errors.New("account not exists")
			return result
		}
		err = chainService.checkBalance(gasLimit, gasPrice, originBalance, chainTypes.GasTable[chainTypes.RegisterMinerType], nil)
		if err != nil {
			result.Err = err
			return result
		}

		result = chainService.executeRegisterMinerTransaction(action, fromAccount, originBalance, gasPrice, gasLimit, tx.ChainId())
	case chainTypes.TransferType:
		action := &chainTypes.TransferAction{}
		err = binary.Unmarshal(tx.GetData(), action)
		if err != nil{
			result.Err = err
			return result
		}
		if !chainService.DatabaseService.ExistAccount(action.To, true) {
			result.Err = errors.New("the accout that transfer isnot exit")
			return result
		}
		err = chainService.checkBalance(gasLimit, gasPrice, originBalance, chainTypes.GasTable[chainTypes.TransferType], nil)
		if err != nil {
			result.Err = err
			return result
		}
		result = chainService.executeTransferTransaction(action, fromAccount, originBalance, gasPrice, gasLimit, tx.ChainId())
	case chainTypes.CreateContractType:
		var msg *chainTypes.Message
		msg, err = chainTypes.TxToMessage(tx)
		if err != nil{
			result.Err = err
			return result
		}
		contractName := msg.Action.(*chainTypes.CreateContractAction).ContractName
		if chainService.DatabaseService.ExistAccount(contractName, true) {
			result.Err = errors.New("contract has exited")
			return result
		}
		err = chainService.checkBalance(gasLimit, gasPrice, originBalance,nil, chainTypes.GasTable[chainTypes.CreateContractType])
		if err != nil {
			result.Err = err
			return result
		}
		result = chainService.executeCreateContractTransaction(originBalance, gasPrice, msg)
	case chainTypes.CallContractType:
		var msg *chainTypes.Message
		msg, err = chainTypes.TxToMessage(tx)
		if err != nil{
			result.Err = err
			return result
		}
		if !chainService.DatabaseService.ExistAccount(msg.Action.(*chainTypes.CallContractAction).ContractName, true) {
			result.Err = errors.New("contract not exited")
			return result
		}
		err = chainService.checkBalance(gasLimit, gasPrice, originBalance,nil, chainTypes.GasTable[chainTypes.CallContractType])
		if err != nil {
			result.Err = err
			return result
		}
		result = chainService.executeCallContractTransaction(originBalance, gasPrice, msg)
	}
	if result.Err != nil {
		return result
	}
	//add nounce while success
	//TODO consider about
	//TODO validate:  success add nounce , fail not to do
	//TODO process tx in block: success add nounce , (fail not to do)?
	chainService.DatabaseService.PutNonce(fromAccount, nounce+1, true)
	return result
}

func (chainService *ChainService) executeRegisterAccountTransaction(action *chainTypes.RegisterAccountAction, fromAccount string, balance, gasPrice, gasLimit *big.Int, chainId app.ChainIdType) *chainTypes.ExecuteReuslt {
	result := &chainTypes.ExecuteReuslt{}
	gasUsed := new(big.Int).Set(chainTypes.GasTable[chainTypes.RegisterAccountType])
	gasFee := new(big.Int).Mul(gasUsed, gasPrice)
	result.GasFee = gasFee
	result.GasUsed = gasUsed
	leftBalance, gasFee := chainService.deduct(chainId, balance, gasFee)  //sub gas fee
	if leftBalance.Sign() >= 0 {
		storage := chainTypes.NewStorage(action.Name, chainId, action.ChainCode, action.Authority)
		chainService.DatabaseService.PutStorage(action.Name, storage, true)
		chainService.DatabaseService.PutBalance(fromAccount, leftBalance, true)
	}else {
		result.Err = errBalance
		return result
	}
	return result
}

func (chainService *ChainService) executeRegisterMinerTransaction(action *chainTypes.RegisterMinerAction, fromAccount string, balance, gasPrice, gasLimit *big.Int, chainId app.ChainIdType) *chainTypes.ExecuteReuslt {
	result := &chainTypes.ExecuteReuslt{}
	gasUsed := new(big.Int).Set(chainTypes.GasTable[chainTypes.RegisterMinerType])
	gasFee := new(big.Int).Mul(gasUsed, gasPrice)
	result.GasFee = gasFee
	result.GasUsed = gasUsed
	leftBalance, gasFee := chainService.deduct(chainId, balance, gasFee)  //sub gas fee
	if leftBalance.Sign() >= 0 {
		biosStorage, err := chainService.DatabaseService.GetStorage(chainService.Config.Bios, true)
		if err != nil {
			result.Err = errors.New("bios not exist")
			return result
		}
		biosStorage.Miner[action.MinerAccount] = &action.SignKey
		chainService.DatabaseService.PutStorage(chainService.Config.Bios, biosStorage, true)
		chainService.DatabaseService.PutBalance(fromAccount, leftBalance, true)
	}else {
		result.Err = errBalance
		return result
	}
	return result
}

func (chainService *ChainService) executeTransferTransaction(action *chainTypes.TransferAction, fromAccount string, balance, gasPrice, gasLimit *big.Int, chainId app.ChainIdType) *chainTypes.ExecuteReuslt {
	result := &chainTypes.ExecuteReuslt{}
	gasUsed := new(big.Int).Set(chainTypes.GasTable[chainTypes.TransferType])
	gasFee := new(big.Int).Mul(gasUsed, gasPrice)
	result.GasFee = gasFee
	result.GasUsed = gasUsed
	leftBalance, gasFee := chainService.deduct(chainId, balance, gasFee)  //sub gas fee
	if leftBalance.Cmp(&action.Amount) >= 0 {				//sub transfer amount
		leftBalance = new(big.Int).Sub(leftBalance, &action.Amount)
		balanceTo, err := chainService.DatabaseService.GetBalance(action.To, true)
		if err != nil {
			result.Err = err
			return result
		}
		balanceTo = new(big.Int).Add(balanceTo, &action.Amount)
		chainService.DatabaseService.PutBalance(fromAccount, leftBalance, true)
		chainService.DatabaseService.PutBalance(action.To, balanceTo, true)
	} else {
		result.Err = errBalance
		return result
	}
	return result
}

func (chainService *ChainService) executeCreateContractTransaction(balance, gasPrice *big.Int, message *chainTypes.Message) *chainTypes.ExecuteReuslt {
	result := &chainTypes.ExecuteReuslt{}
	executedValue, refundGas, err := chainService.VmService.ApplyMessage(message)
	if err!=nil {
		result.Err = err
		return result
	}
	balance,err = chainService.DatabaseService.GetBalance(message.From, true)
	if err!=nil {
		result.Err = err
		return result
	}
	gasUsed := new(big.Int).Sub(&message.Gas, new(big.Int).SetUint64(refundGas))
	gasFee  := new(big.Int).Mul(gasUsed, gasPrice)
	leftBalance, gasFee := chainService.deduct(message.ChainId, balance, gasFee)
	if leftBalance.Sign() >= 0{
		chainService.DatabaseService.PutBalance(message.From, leftBalance, true)
		result.Return = common.Bytes(executedValue)
		result.GasFee = gasFee
		result.GasUsed = gasUsed
		return result
	}else{
		result.Err = errBalance
		result.GasFee = gasFee
		result.GasUsed = gasUsed
		return result
	}
}

func (chainService *ChainService) executeCallContractTransaction(balance, gasPrice *big.Int, message *chainTypes.Message) *chainTypes.ExecuteReuslt {
	result := &chainTypes.ExecuteReuslt{}
	executedValue, returnGas, err := chainService.VmService.ApplyMessage(message)
	if err != nil {
		result.Err = err
		return result
	}
	balance, err = chainService.DatabaseService.GetBalance(message.From, true)
	if err != nil {
		result.Err = err
		return result
	}
	gasUsed := new(big.Int).Sub(&message.Gas, new(big.Int).SetUint64(returnGas))
	gasFee  := new(big.Int).Mul(gasUsed, gasPrice)
	leftBalance, gasFee := chainService.deduct(message.ChainId, balance, gasFee)
	if leftBalance.Sign() >= 0{
		chainService.DatabaseService.PutBalance(message.From, leftBalance, true)
		result.Return =  common.Bytes(executedValue)
		result.GasFee = gasFee
		result.GasUsed = gasUsed
		return result
	}else{
		result.Err = errBalance
		result.GasFee = gasFee
		result.GasUsed = gasUsed
		return result
	}
}

func  (chainService *ChainService) checkNonce(fromAccount string, nounce int64) error{
	nonce := chainService.DatabaseService.GetNonce(fromAccount, true)
	if nonce > nounce {
		return errors.New("error nouce")
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
//    err := binary.Unmarshal(t.Data.Data, cct)
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
//    err := binary.Unmarshal(t.Data.Data, &cct)
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
//    t.Data.Data, _ = binary.Marshal(cct)
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