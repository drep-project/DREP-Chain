package vm

import "BlockChainTest/accounts"

type Log struct {
    Address      accounts.CommonAddress
    ChainId      int64
    TxHash       []byte
    Topics       [][]byte
    Data         []byte
}