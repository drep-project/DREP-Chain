package trace

import (
	"github.com/drep-project/binary"
	"github.com/drep-project/drep-chain/common"
	"github.com/drep-project/drep-chain/crypto"
	"github.com/drep-project/drep-chain/types"
)

/*
name: 记录接口
usage: 查询交易地址等信息（需要开启记录模块）
prefix:trace
*/
type TraceApi struct {
	blockAnalysis *BlockAnalysis
	traceService  *TraceService
}

/*
 name: getRawTransaction
 usage: 根据交易hash查询交易字节
 params:
	1. 交易hash
 return: 交易字节信息
 example:  curl http://localhost:15645 -X POST --data '{"jsonrpc":"2.0","method":"trace_getRawTransaction","params":["0x00001c9b8c8fdb1f53faf02321f76253704123e2b56cce065852bab93e526ae2"], "id": 3}' -H "Content-Type:application/json"
 response:
   {
	  "jsonrpc": "2.0",
	  "id": 3,
	  "result": "0x02a7ae20007923a30bbfbcb998a6534d56b313e68c8e0c594a0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002011102011003030000bc9889d00b004120eba14c77eab7a154833ff14832d8769cfc0b30db288445d6a83ef2fe337aa09042f8174a593543c4acabe7fadf1ad5fceea9c835682cb9dbea3f1d8fec181fb9"
	}
*/
func (traceApi *TraceApi) GetRawTransaction(txHash *crypto.Hash) (string, error) {
	rawData, err := traceApi.blockAnalysis.store.GetRawTransaction(txHash)
	if err != nil {
		return "", err
	}
	return common.Encode(rawData), nil
}

/*
 name: getTransaction
 usage: 根据交易hash查询交易详细信息
 params:
	1. 交易hash
 return: 交易详细信息
 example: curl http://localhost:15645 -X POST --data '{"jsonrpc":"2.0","method":"trace_getTransaction","params":["0x00001c9b8c8fdb1f53faf02321f76253704123e2b56cce065852bab93e526ae2"], "id": 3}' -H "Content-Type:application/json"
 response:
   {
	  "jsonrpc": "2.0",
	  "id": 3,
	  "result": {
		"Hash": "0x00001c9b8c8fdb1f53faf02321f76253704123e2b56cce065852bab93e526ae2",
		"From": "0x7923a30bbfbcb998a6534d56b313e68c8e0c594a",
		"Version": 1,
		"Nonce": 530215,
		"Type": 0,
		"To": "0x7923a30bbfbcb998a6534d56b313e68c8e0c594a",
		"ChainId": "00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
		"Amount": "0x111",
		"GasPrice": "0x110",
		"GasLimit": "0x30000",
		"Timestamp": 1560356382,
		"Data": null,
		"Sig": "0x20eba14c77eab7a154833ff14832d8769cfc0b30db288445d6a83ef2fe337aa09042f8174a593543c4acabe7fadf1ad5fceea9c835682cb9dbea3f1d8fec181fb9"
	  }
	}
*/
func (traceApi *TraceApi) GetTransaction(txHash *crypto.Hash) (*RpcTransaction, error) {
	rpcTx, err := traceApi.blockAnalysis.store.GetTransaction(txHash)
	if err != nil {
		return nil, err
	}
	return rpcTx, nil
}

/*
 name: decodeTrasnaction
 usage: 把交易字节信息反解析成交易详情
 params:
	1. 交易字节信息
 return: 交易详情
 example: curl http://localhost:15645 -X POST --data '{"jsonrpc":"2.0","method":"trace_decodeTrasnaction","params":["0x02a7ae20007923a30bbfbcb998a6534d56b313e68c8e0c594a0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002011102011003030000bc9889d00b004120eba14c77eab7a154833ff14832d8769cfc0b30db288445d6a83ef2fe337aa09042f8174a593543c4acabe7fadf1ad5fceea9c835682cb9dbea3f1d8fec181fb9"], "id": 3}' -H "Content-Type:application/json"
 response:
   {
	  "jsonrpc": "2.0",
	  "id": 3,
	  "result": {
		"Hash": "0x00001c9b8c8fdb1f53faf02321f76253704123e2b56cce065852bab93e526ae2",
		"From": "0x7923a30bbfbcb998a6534d56b313e68c8e0c594a",
		"Version": 1,
		"Nonce": 530215,
		"Type": 0,
		"To": "0x7923a30bbfbcb998a6534d56b313e68c8e0c594a",
		"ChainId": "00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
		"Amount": "0x111",
		"GasPrice": "0x110",
		"GasLimit": "0x30000",
		"Timestamp": 1560356382,
		"Data": null,
		"Sig": "0x20eba14c77eab7a154833ff14832d8769cfc0b30db288445d6a83ef2fe337aa09042f8174a593543c4acabe7fadf1ad5fceea9c835682cb9dbea3f1d8fec181fb9"
	  }
	}
*/
func (traceApi *TraceApi) DecodeTrasnaction(bytes common.Bytes) (*RpcTransaction, error) {
	tx := &types.Transaction{}
	err := binary.Unmarshal(bytes[:], tx)
	if err != nil {
		return nil, err
	}
	rpcTx := &RpcTransaction{}
	rpcTx.FromTx(tx)
	return rpcTx, nil
}

/*
 name: getSendTransactionByAddr
 usage: 根据地址查询该交易发出的交易，支持分页
 params:
	1. 交易地址
	2. 分页号（从1开始）
    3. 页大小
 return: 交易列表
 example: curl http://localhost:15645 -X POST --data '{"jsonrpc":"2.0","method":"trace_getSendTransactionByAddr","params":["0x7923a30bbfbcb998a6534d56b313e68c8e0c594a",1,10], "id": 3}' -H "Content-Type:application/json"
 response:
   {
	  "jsonrpc": "2.0",
	  "id": 3,
	  "result": [
		{
		  "Hash": "0x00001c9b8c8fdb1f53faf02321f76253704123e2b56cce065852bab93e526ae2",
		  "From": "0x7923a30bbfbcb998a6534d56b313e68c8e0c594a",
		  "Version": 1,
		  "Nonce": 530215,
		  "Type": 0,
		  "To": "0x7923a30bbfbcb998a6534d56b313e68c8e0c594a",
		  "ChainId": "00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
		  "Amount": "0x111",
		  "GasPrice": "0x110",
		  "GasLimit": "0x30000",
		  "Timestamp": 1560356382,
		  "Data": null,
		  "Sig": "0x20eba14c77eab7a154833ff14832d8769cfc0b30db288445d6a83ef2fe337aa09042f8174a593543c4acabe7fadf1ad5fceea9c835682cb9dbea3f1d8fec181fb9"
		}
	  ]
	}
*/
func (traceApi *TraceApi) GetSendTransactionByAddr(addr *crypto.CommonAddress, pageIndex, pageSize int) []*RpcTransaction {
	return traceApi.blockAnalysis.store.GetSendTransactionsByAddr(addr, pageIndex, pageSize)
}

/*
 name: getReceiveTransactionByAd
 usage: 根据地址查询该交易接受的交易，支持分页
 params:
	1. 交易地址
	2. 分页号（从1开始）
    3. 页大小
 return: 交易列表
 example: curl http://localhost:15645 -X POST --data '{"jsonrpc":"2.0","method":"trace_getReceiveTransactionByAddr","params":["0x3ebcbe7cb440dd8c52940a2963472380afbb56c5",1,10], "id": 3}' -H "Content-Type:application/json"
 response:
   {
	  "jsonrpc": "2.0",
	  "id": 3,
	  "result": [
		{
		  "Hash": "0x3d3e7da272a5128bec6fd7ad10d8557b08e0fb9de4af6753641e29740eb7054e",
		  "From": "0x7923a30bbfbcb998a6534d56b313e68c8e0c594a",
		  "Version": 1,
		  "Nonce": 553770,
		  "Type": 0,
		  "To": "0x3ebcbe7cb440dd8c52940a2963472380afbb56c5",
		  "ChainId": "00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
		  "Amount": "0xde0b6b3a7640000",
		  "GasPrice": "0x12c",
		  "GasLimit": "0x7530",
		  "Timestamp": 1560403673,
		  "Data": null,
		  "Sig": "0x1f073cd3f2621abe15ef949b27c7d0a16d69a64aaa9e95973b9c94de2d7b8f4b103928988478d2f248ae7a9dc6a156d12d300adc5e9059decc037a67e94fe0c3a2"
		}
	  ]
	}
*/
func (traceApi *TraceApi) GetReceiveTransactionByAddr(addr *crypto.CommonAddress, pageIndex, pageSize int) []*RpcTransaction {
	return traceApi.blockAnalysis.store.GetReceiveTransactionsByAddr(addr, pageIndex, pageSize)
}

/*
 name: rebuild
 usage: 重建trace中的区块记录
 params:
	1. 起始块（包含）
	2. 终止块（不包含）
 return:
 example: curl http://localhost:15645 -X POST --data '{"jsonrpc":"2.0","method":"trace_rebuild","params":[1,10], "id": 3}' -H "Content-Type:application/json"
 response:
  	{"jsonrpc":"2.0","id":3,"result":null}
*/
func (traceApi *TraceApi) Rebuild(from, end int) error {
	if from < 0 {
		from = 0
	}
	if end < 0 {
		end = int(traceApi.traceService.ChainService.BestChain().Height())
	}
	if from > end {
		return nil
	}
	return traceApi.blockAnalysis.Rebuild(from, end)
}
