package bean

import (
    "encoding/hex"
    "BlockChainTest/mycrypto"
    "math/big"
    "encoding/json"
)

const (
    AddressLen = 20
)

func (tx *Transaction) TxId() (string, error) {
    b, err := json.Marshal(tx.Data)
    if err != nil {
        return "", err
    }
    id := hex.EncodeToString(mycrypto.Hash256(b))
    return id, nil
}

func (tx *Transaction) TxHash() ([]byte, error) {
    b, err := json.Marshal(tx)
    if err != nil {
        return nil, err
    }
    h := mycrypto.Hash256(b)
    return h, nil
}

func (tx *Transaction) TxSig(prvKey *mycrypto.PrivateKey) (*mycrypto.Signature, error) {
    b, err := json.Marshal(tx.Data)
    if err != nil {
        return nil, err
    }
    return mycrypto.Sign(prvKey, b)
}

func (tx *Transaction) Addr() Address {
    return Addr(tx.Data.PubKey)
}

func (tx *Transaction) GetGasQuantity() *big.Int {
    return new(big.Int).SetInt64(int64(100))
}

func (tx *Transaction) GetGasUsed() *big.Int {
    gasQuantity := tx.GetGasQuantity()
    gasPrice := new(big.Int).SetBytes(tx.Data.GasPrice)
    gasUsed := new(big.Int).Mul(gasQuantity, gasPrice)
    return gasUsed
}

func (block *Block) BlockID() (string, error) {
    b, err := json.Marshal(block.Header)
    if err != nil {
        return "", err
    }
    id := hex.EncodeToString(mycrypto.Hash256(b))
    return id, nil
}

func Height2Key(height int64) string {
    return hex.EncodeToString(new(big.Int).SetInt64(height).Bytes())
}

func (block *Block) DBKey() string {
    return Height2Key(block.Header.Height)
}

func (block *Block) DBMarshal() ([]byte, error) {
    _b, err := json.Marshal(block)
    if err != nil {
        return nil, err
    }
    b := make([]byte, len(_b) + 1)
    b[0] = byte(MsgTypeBlock)
    copy(b[1:], _b)
    return b, nil
}