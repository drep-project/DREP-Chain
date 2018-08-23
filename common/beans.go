package common

const (
    WAITING = 0
    MSG_SETUP1 = 1 // LS ->CHA M received then commit
    MSG_BLOCK1_COMMIT = 3 // MS  MS->RESPONSE
    MSG_BLOCK1_CHALLENGE = 4 // LS ->SETUP2
    MSG_BLOCK1_RESPONSE = 5 // MS -> COMMIT

    MSG_SETUP2 = 6 // LS -> CHA
    MSG_BLOCK2_COMMIT = 7 // MS -> RESP
    MSG_BLOCK2_CHALLENGE = 8 // LS -> BLOCK
    MSG_BLOCK2_RESPONSE = 9 // MS -> BLOCK

    MSG_BLOCK = 9 // LS -> WAITING M -> WAITING
    MSG_TRANSACTION = 10
)
type Message struct {
    Type int
    Body interface{}
}

type Address string

//type BlockHeader struct {
//    Version      int32
//    PreviousHash []byte
//    GasLimit     big.Int
//    GasUsed      big.Int
//    Height       int32
//    CreatedTime  int32
//    MerkleRoot   []byte
//    TxHashes     [][]byte
//    LeaderPubKey []byte
//    MinorPubKeys [][]byte
//}
//
//type BlockData struct {
//    TxCount int32
//    TxList  []*Transaction
//}
//
//type Block struct {
//    Header   *BlockHeader
//    Data     *BlockData
//    MultiSig []byte
//}
//
//type Transaction struct {
//    Version      int32
//    Nonce        int64
//    ToAddress    []byte
//    Amount       big.Int
//    GasPrice     big.Int
//    GasLimit     big.Int
//    ProducedTime int32
//    PubKey       []byte
//    Sig          []byte
//}
//
//func (t *Transaction) GetId() string {
//    return ""
//}
//
//type BlockChain struct {
//    Blocks []Block
//}
//

