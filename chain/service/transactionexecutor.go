package service

import (
	"errors"
	"fmt"
	"github.com/drep-project/dlog"
	"github.com/drep-project/drep-chain/app"
	chainTypes "github.com/drep-project/drep-chain/chain/types"
	"github.com/drep-project/drep-chain/pkgs/evm/vm"
	"math/big"

	"bytes"
	"encoding/hex"
	"encoding/json"
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

func (chainService *ChainService) ValidateTransaction(tx *chainTypes.Transaction) error {
	var result error
	_ = chainService.DatabaseService.Transaction(func() error {
		_, _, result = chainService.executeTransaction(tx,)
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
		_, gasFee, err := chainService.executeTransaction(tx)
		if err != nil {
			return nil, err
		}
		if gasFee != nil {
			total.Add(total, gasFee)
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
	data, err := json.Marshal(cct)
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

//TODO we can use tx handlee but not switch if more tx types
func (chainService *ChainService) executeTransaction(tx *chainTypes.Transaction) (*big.Int, *big.Int, error) {
	nounce := tx.Nonce()
	amount := tx.Amount()
	fromAccount := tx.From()
	gasPrice := tx.GasPrice()
	gasLimit := tx.GasLimit()

	originBalance := chainService.DatabaseService.GetBalance(fromAccount, true)
	err := chainService.checkNonce(fromAccount, nounce)
	if err != nil {
		return  nil, nil, err
	}

	gasUsed := new (big.Int)
	gasFee := new (big.Int)
	switch tx.Type() {
	case chainTypes.RegisterAccountType:
		action := &chainTypes.RegisterAccountAction{}
		err = json.Unmarshal(tx.GetData(), action)
		if err != nil{
			return  nil, nil, err
		}
		err = chainService.checkBalance(gasLimit, gasPrice, originBalance, chainTypes.GasTable[chainTypes.RegisterAccountType], nil)
		if err != nil {
			return  nil, nil, err
		}

		gasUsed, gasFee, err = chainService.executeRegisterAccountTransaction(action, fromAccount, originBalance, gasPrice, gasLimit, amount, tx.ChainId())
		if err != nil {
			return  nil, nil, err
		}
	case chainTypes.RegisterMinerType:
		//TODO how to control add miner right
		panic(errors.New("unpupport add miners runtime"))
		action := &chainTypes.RegisterMinerAction{}
		err = json.Unmarshal(tx.GetData(), action)
		if err != nil{
			return  nil, nil, err
		}
		err = chainService.checkBalance(gasLimit, gasPrice, originBalance, chainTypes.GasTable[chainTypes.RegisterMinerType], nil)
		if err != nil {
			return  nil, nil, err
		}

		gasUsed, gasFee, err = chainService.executeRegisterMinerTransaction(action, fromAccount, originBalance, gasPrice, gasLimit, amount, tx.ChainId())
		if err != nil {
			return  nil, nil, err
		}
	case chainTypes.TransferType:
		action := &chainTypes.TransferAction{}
		err = json.Unmarshal(tx.GetData(), action)
		if err != nil{
			return  nil, nil, err
		}
		err = chainService.checkBalance(gasLimit, gasPrice, originBalance, chainTypes.GasTable[chainTypes.TransferType], nil)
		if err != nil {
			return  nil, nil, err
		}
		gasUsed, gasFee, err = chainService.executeTransferTransaction(action, fromAccount, originBalance, gasPrice, gasLimit, amount, tx.ChainId())
	case chainTypes.CreateContractType:
		var msg *chainTypes.Message
		msg, err = chainTypes.TxToMessage(tx)
		if err != nil{
			return  nil, nil, err
		}
		err = chainService.checkBalance(gasLimit, gasPrice, originBalance,nil, chainTypes.GasTable[chainTypes.CreateContractType])
		if err != nil {
			return  nil, nil, err
		}
		gasUsed, gasFee, err = chainService.executeCreateContractTransaction(originBalance, gasPrice, msg)
	case chainTypes.CallContractType:
		var msg *chainTypes.Message
		msg, err = chainTypes.TxToMessage(tx)
		if err != nil{
			return  nil, nil, err
		}
		err = chainService.checkBalance(gasLimit, gasPrice, originBalance,nil, chainTypes.GasTable[chainTypes.CallContractType])
		if err != nil {
			return  nil, nil, err
		}
		gasUsed, gasFee, err = chainService.executeCallContractTransaction(originBalance, gasPrice, msg)
	}
	if err != nil {
		dlog.Error("executeTransaction transaction error", "reason", err)
		return gasUsed, gasFee, err
	}
	chainService.DatabaseService.PutNonce(fromAccount, nounce+1, true)
	return gasUsed, gasFee, nil
}

func (chainService *ChainService) executeRegisterAccountTransaction(action *chainTypes.RegisterAccountAction, fromAccount string, balance, gasPrice, gasLimit, amount *big.Int, chainId app.ChainIdType) (*big.Int, *big.Int, error) {
	if fromAccount == ""||action.Name == "" {
		return  nil, nil, errors.New("invalidate account")
	}
	storage, _ := chainService.DatabaseService.GetStorage(action.Name, true)
	if storage != nil {
		return  nil, nil, errors.New("account exists")
	}
	table := chainTypes.GasTable
	gasUsed := new(big.Int).Set(table[chainTypes.RegisterAccountType])
	gasFee := new(big.Int).Mul(gasUsed, gasPrice)
	leftBalance, gasFee := chainService.deduct(chainId, balance, gasFee)  //sub gas fee
	if leftBalance.Sign() >= 0 {
		storage := chainTypes.NewStorage(action.Name, chainId, action.ChainCode, action.Authority)
		chainService.DatabaseService.PutStorage(action.Name,storage, true)
		chainService.DatabaseService.PutBalance(fromAccount, balance, true)
	}else {
		return gasUsed, gasFee, errBalance
	}
	return gasUsed, gasFee, nil
}

func (chainService *ChainService) executeRegisterMinerTransaction(action *chainTypes.RegisterMinerAction, fromAccount string, balance, gasPrice, gasLimit, amount *big.Int, chainId app.ChainIdType) (*big.Int, *big.Int, error) {
	if fromAccount == ""||action.MinerAccount == "" {
		return  nil, nil, errors.New("invalidate account")
	}
	gasUsed := new(big.Int).Set(chainTypes.GasTable[chainTypes.RegisterMinerType])
	gasFee := new(big.Int).Mul(gasUsed, gasPrice)
	leftBalance, gasFee := chainService.deduct(chainId, balance, gasFee)  //sub gas fee
	if leftBalance.Sign() >= 0 {
		biosStorage, err := chainService.DatabaseService.GetStorage(chainService.Config.Bios, true)
		if err != nil {
			return  nil, nil, errors.New("bios not exist")
		}
		biosStorage.Miner[action.MinerAccount] = &action.SignKey
		chainService.DatabaseService.PutStorage(chainService.Config.Bios, biosStorage, true)
		chainService.DatabaseService.PutBalance(fromAccount, balance, true)
	}else {
		return gasUsed, gasFee, errBalance
	}
	return gasUsed, gasFee, nil
}

func (chainService *ChainService) executeTransferTransaction(action *chainTypes.TransferAction, fromAccount string, balance, gasPrice, gasLimit, amount *big.Int, chainId app.ChainIdType) (*big.Int, *big.Int, error) {
	if fromAccount == ""||action.To == "" {
		return  nil, nil, errors.New("invalidate account")
	}
	gasUsed := new(big.Int).Set(chainTypes.GasTable[chainTypes.TransferType])
	gasFee := new(big.Int).Mul(gasUsed, gasPrice)
	leftBalance, gasFee := chainService.deduct(chainId, balance, gasFee)  //sub gas fee
	if leftBalance.Cmp(amount) >= 0 {				//sub transfer amount
		leftBalance = new(big.Int).Sub(leftBalance, amount)
		balanceTo := chainService.DatabaseService.GetBalance(action.To, true)
		balanceTo = new(big.Int).Add(balanceTo, amount)
		chainService.DatabaseService.PutBalance(fromAccount, balance, true)
		chainService.DatabaseService.PutBalance(action.To, balanceTo, true)
	} else {
		return gasUsed, gasFee, errBalance
	}
	return gasUsed, gasFee, nil
}

func (chainService *ChainService) executeCreateContractTransaction(balance, gasPrice *big.Int, message *chainTypes.Message) (*big.Int, *big.Int, error) {
	if message.From== "" {
		return  nil, nil, errors.New("invalidate account")
	}
	evm := vm.NewEVM(chainService.DatabaseService, message.ChainId)
	refundGas, err := chainService.VmService.ApplyMessage(evm, message)
	balance = chainService.DatabaseService.GetBalance(message.From, true)
	gasUsed := new(big.Int).Sub(message.Gas, new(big.Int).SetUint64(refundGas))
	gasFee  := new(big.Int).Mul(gasUsed, gasPrice)
	leftBalance, gasFee := chainService.deduct(message.ChainId, balance, gasFee)
	if leftBalance.Sign() >= 0{
		chainService.DatabaseService.PutBalance(message.From, leftBalance, true)
		return gasUsed, gasFee, err
	}else{
		return gasUsed, gasFee, errBalance
	}
}

func (chainService *ChainService) executeCallContractTransaction(balance, gasPrice *big.Int, message *chainTypes.Message) (*big.Int, *big.Int, error) {
	if message.From== "" || message.Action.(*chainTypes.CallContractAction).ContractName == "" {
		return  nil, nil, errors.New("invalidate account")
	}
	evm := vm.NewEVM(chainService.DatabaseService, message.ChainId)
	returnGas, err := chainService.VmService.ApplyMessage(evm, message)
	balance = chainService.DatabaseService.GetBalance(message.From, true)
	gasUsed := new(big.Int).Sub(message.Gas, new(big.Int).SetUint64(returnGas))
	gasFee  := new(big.Int).Mul(gasUsed, gasPrice)
	leftBalance, gasFee := chainService.deduct(message.ChainId, balance, gasFee)
	if leftBalance.Sign() >= 0{
		chainService.DatabaseService.PutBalance(message.From, leftBalance, true)
		return gasUsed, gasFee, err
	}else{
		return gasUsed, gasFee, errBalance
	}
}

func  (chainService *ChainService) checkNonce(fromAccount string, nounce int64) error{
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