
#  `JSON-RPC`接口说明文档
## 区块
用于处理区块链偏上层逻辑

### 1. blockMgr_sendRawTransaction
#### 作用：发送已签名的交易.
> 参数：
 1. 已签名的交易

#### 返回值：交易hash

#### 示例代码
##### 请求：

```shell
curl http://localhost:15645 -X POST --data '{"jsonrpc":"2.0","method":"blockmgr_sendRawTransaction","params":["0x40a287b6d30b05313131317a4120dd8c23c40910d038fa43b2f8932d3681cbe5ee3079b6e9de0bea6e8e6b2a867a561aa26e1cd6b62aa0422a043186b593b784bf80845c3fd5a7fbfe62e61d8564"], "id": 3}' -H "Content-Type:application/json"
```

##### 响应：

```json
{"jsonrpc":"2.0","id":1,"result":"0xf30e858667fa63bc57ae395c3f57ede9bb3ad4969d12f4bce51d900fb5931538"}
````


### 2. blockMgr_gasPrice
#### 作用：获取系统的给出的gasprice建议值
> 参数：
 1. 待查询地址

#### 返回值：价格和是否错误信息

#### 示例代码
##### 请求：

```shell
curl http://localhost:15645 -X POST --data '{"jsonrpc":"2.0","method":"blockmgr_gasPrice","params":[], "id": 3}' -H "Content-Type:application/json"
```

##### 响应：

```json

````


### 3. blockMgr_GetPoolTransactions
#### 作用：获取交易池中的交易信息.
> 参数：
 1. 待查询地址

#### 返回值：交易池中所有交易

#### 示例代码
##### 请求：

```shell
curl http://localhost:15645 -X POST --data '{"jsonrpc":"2.0","method":"blockmgr_getPoolTransactions","params":["0x8a8e541ddd1272d53729164c70197221a3c27486"], "id": 3}' -H "Content-Type:application/json"
```

##### 响应：

```json

````


### 4. blockMgr_GetTransactionCount
#### 作用：获取地址发出的交易总数
> 参数：
 1. 待查询地址

#### 返回值：交易池中所有交易

#### 示例代码
##### 请求：

```shell
curl http://localhost:15645 -X POST --data '{"jsonrpc":"2.0","method":"blockmgr_getTransactionCount","params":["0x8a8e541ddd1272d53729164c70197221a3c27486"], "id": 3}' -H "Content-Type:application/json"
```

##### 响应：

```json

````


### 5. blockMgr_GetTxInPool
#### 作用：查询交易是否在交易池，如果在，返回交易
> 参数：
 1. 发起转账的地址

#### 返回值：交易完整信息

#### 示例代码
##### 请求：

```shell
curl -H "Content-Type: application/json" -X post --data '{"jsonrpc":"2.0","method":"blockmgr_getTxInPool","params":["0x3ebcbe7cb440dd8c52940a2963472380afbb56c5"],"id":1}' http://127.0.0.1:15645
```

##### 响应：

```json
{
  "jsonrpc": "2.0",
  "id": 3,
  "result": {
    "Hash": "0xfa5c34114ff459b4c97e7cd268c507c0ccfcfc89d3ccdcf71e96402f9899d040",
    "From": "0x7923a30bbfbcb998a6534d56b313e68c8e0c594a",
    "Version": 1,
    "Nonce": 15632,
    "Type": 0,
    "To": "0x7923a30bbfbcb998a6534d56b313e68c8e0c594a",
    "ChainId": "00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
    "Amount": "0x111",
    "GasPrice": "0x110",
    "GasLimit": "0x30000",
    "Timestamp": 1559322808,
    "Data": null,
    "Sig": "0x20f25b86c4bf73aa4fa0bcb01e2f5731de3a3917c8861d1ce0574a8d8331aedcf001e678000f6afc95d35a53ef623a2055fce687f85c2fd752dc455ab6db802b1f"
  }
}
````

链接口
用于获取区块信息

### 1. chain_getblock
#### 作用：用于获取区块信息
> 参数：
 1. height  usage: 当前区块高度

#### 返回值：区块明细信息

#### 示例代码
##### 请求：

```shell
curl http://localhost:15645 -X POST --data '{"jsonrpc":"2.0","method":"chain_getBlock","params":[1], "id": 3}' -H "Content-Type:application/json"
```

##### 响应：

```json
{
  "jsonrpc": "2.0",
  "id": 3,
  "result": {
    "Hash": "0xcfa283a5b591da5a15971bf62fffae87e649bcf749776f4c83ffe50e65920f8e",
    "ChainId": "00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
    "Version": 1,
    "PreviousHash": "0x1717b4b9f740cebeb2659886122a29c0876ed906dd05370319fee4ecf219b1e9",
    "GasLimit": 180000000,
    "GasUsed": 0,
    "Height": 1,
    "Timestamp": 1559272779,
    "StateRoot": "0xd7bd5b3af4f2f1fb3d484743052c2e911f9fb7b04131660912244347508f16a9",
    "TxRoot": "0x",
    "LeaderAddress": "0x0374bf9c8ea268b5548686685dda4a74fc95903ca7c440e5b187a718b595c1f374",
    "MinorAddresses": [
      "0x0374bf9c8ea268b5548686685dda4a74fc95903ca7c440e5b187a718b595c1f374",
      "0x02f11cfd138eaaaba5f8c0a7f1f2791bdabd0b0c404734dceac820aa9b683bfb1a",
      "0x03949aad279a32536ce20f0957c9c6ba592532ea70e5f174332bed4c94382354e3",
      "0x0263bc5628fa7033727d14b5d6714ac7d6a5d34bc5db994a896f54499f12db9b0b"
    ],
    "Txs": [

    ]
  }
}
````


### 2. chain_getMaxHeight
#### 作用：用于获取当前最高区块
> 参数：
 1. 无

#### 返回值：当前最高区块高度数值

#### 示例代码
##### 请求：

```shell
curl http://localhost:15645 -X POST --data '{"jsonrpc":"2.0","method":"chain_getMaxHeight","params":[], "id": 3}' -H "Content-Type:application/json"
```

##### 响应：

```json
{"jsonrpc":"2.0","id":3,"result":193005}
````


### 3. chain_getBlockGasInfo
#### 作用：获取gas相关信息
> 参数：
 1. 无

#### 返回值：系统需要的gas最小值、最大值；和当前块被设置的最大gas值

#### 示例代码
##### 请求：

```shell
curl http://localhost:15645 -X POST --data '{"jsonrpc":"2.0","method":"chain_getBlockGasInfo","params":[], "id": 3}' -H "Content-Type:application/json"
```

##### 响应：

```json
{"jsonrpc":"2.0","id":3,"result":193005}
````


### 4. chain_getBalance
#### 作用：查询地址余额
> 参数：
 1. 待查询地址

#### 返回值：地址中的账号余额

#### 示例代码
##### 请求：

```shell
curl http://localhost:15645 -X POST --data '{"jsonrpc":"2.0","method":"chain_getBalance","params":["0x8a8e541ddd1272d53729164c70197221a3c27486"], "id": 3}' -H "Content-Type:application/json"
```

##### 响应：

```json
{"jsonrpc":"2.0","id":3,"result":9987999999999984000000}
````


### 5. chain_getNonce
#### 作用：查询地址在链上的nonce
> 参数：
 1. 待查询地址

#### 返回值：链上nonce

#### 示例代码
##### 请求：

```shell
curl http://localhost:15645 -X POST --data '{"jsonrpc":"2.0","method":"chain_getNonce","params":["0x8a8e541ddd1272d53729164c70197221a3c27486"], "id": 3}' -H "Content-Type:application/json"
```

##### 响应：

```json
{"jsonrpc":"2.0","id":3,"result":0}
````


### 6. chain_getNonce
#### 作用：查询地址的名誉值
> 参数：
 1. 待查询地址

#### 返回值：地址对应的名誉值

#### 示例代码
##### 请求：

```shell
curl http://localhost:15645 -X POST --data '{"jsonrpc":"2.0","method":"chain_getReputation","params":["0x8a8e541ddd1272d53729164c70197221a3c27486"], "id": 3}' -H "Content-Type:application/json"
```

##### 响应：

```json
{"jsonrpc":"2.0","id":3,"result":1}
````


### 7. chain_getTransactionByBlockHeightAndIndex
#### 作用：获取区块中特定序列的交易
> 参数：
 1. 区块高度
 2. 交易序列

#### 返回值：交易信息

#### 示例代码
##### 请求：

```shell
curl http://localhost:15645 -X POST --data '{"jsonrpc":"2.0","method":"chain_getTransactionByBlockHeightAndIndex","params":[10000,1], "id": 3}' -H "Content-Type:application/json"
```

##### 响应：

```json
{
  "jsonrpc": "2.0",
  "id": 3,
  "result": {
    "Hash": "0xfa5c34114ff459b4c97e7cd268c507c0ccfcfc89d3ccdcf71e96402f9899d040",
    "From": "0x7923a30bbfbcb998a6534d56b313e68c8e0c594a",
    "Version": 1,
    "Nonce": 15632,
    "Type": 0,
    "To": "0x7923a30bbfbcb998a6534d56b313e68c8e0c594a",
    "ChainId": "00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
    "Amount": "0x111",
    "GasPrice": "0x110",
    "GasLimit": "0x30000",
    "Timestamp": 1559322808,
    "Data": null,
    "Sig": "0x20f25b86c4bf73aa4fa0bcb01e2f5731de3a3917c8861d1ce0574a8d8331aedcf001e678000f6afc95d35a53ef623a2055fce687f85c2fd752dc455ab6db802b1f"
  }
}
````


### 8. chain_getAliasByAddress
#### 作用：根据地址获取地址对应的别名
> 参数：
 1. 待查询地址

#### 返回值：地址别名

#### 示例代码
##### 请求：

```shell
curl http://localhost:15645 -X POST --data '{"jsonrpc":"2.0","method":"chain_getAliasByAddress","params":["0x8a8e541ddd1272d53729164c70197221a3c27486"], "id": 3}' -H "Content-Type:application/json"
```

##### 响应：

```json
{"jsonrpc":"2.0","id":3,"result":"tom"}
````


### 9. chain_getAddressByAlias
#### 作用：根据别名获取别名对应的地址
> 参数：
 1. 待查询地别名

#### 返回值：别名对应的地址

#### 示例代码
##### 请求：

```shell
curl http://localhost:15645 -X POST --data '{"jsonrpc":"2.0","method":"chain_getAddressByAlias","params":["tom"], "id": 3}' -H "Content-Type:application/json"
```

##### 响应：

```json
{"jsonrpc":"2.0","id":3,"result":"0x8a8e541ddd1272d53729164c70197221a3c27486"}
````


### 10. chain_getReceipt
#### 作用：根据txhash获取receipt信息
> 参数：
 1. txhash

#### 返回值：receipt

#### 示例代码
##### 请求：

```shell
curl http://localhost:15645 -X POST --data '{"jsonrpc":"2.0","method":"chain_getReceipt","params":["0x7d9dd32ca192e765ff2abd7c5f8931cc3f77f8f47d2d52170c7804c2ca2c5dd9"], "id": 3}' -H "Content-Type:application/json"
```

##### 响应：

```json
{"jsonrpc":"2.0","id":3,"result":""}
````


### 11. chain_getLogs
#### 作用：根据txhash获取交易log信息
> 参数：
 1. txhash

#### 返回值：[]log

#### 示例代码
##### 请求：

```shell
curl http://localhost:15645 -X POST --data '{"jsonrpc":"2.0","method":"chain_getLogs","params":["0x7d9dd32ca192e765ff2abd7c5f8931cc3f77f8f47d2d52170c7804c2ca2c5dd9"], "id": 3}' -H "Content-Type:application/json"
```

##### 响应：

```json
{"jsonrpc":"2.0","id":3,"result":""}
````


### 12. chain_getInterset
#### 作用：根据txhash获取退质押或者投票利息信息
> 参数：
 1. txhash

#### 返回值：{}

#### 示例代码
##### 请求：

```shell
curl http://localhost:15645 -X POST --data '{"jsonrpc":"2.0","method":"chain_getInterset","params":["0x7d9dd32ca192e765ff2abd7c5f8931cc3f77f8f47d2d52170c7804c2ca2c5dd9"], "id": 3}' -H "Content-Type:application/json"
```

##### 响应：

```json
{"jsonrpc":"2.0","id":3,"result":""}
````


### 13. chain_getByteCode
#### 作用：根据地址获取bytecode
> 参数：
 1. 地址

#### 返回值：bytecode

#### 示例代码
##### 请求：

```shell
curl http://localhost:15645 -X POST --data '{"jsonrpc":"2.0","method":"chain_getByteCode","params":["0x8a8e541ddd1272d53729164c70197221a3c27486"], "id": 3}' -H "Content-Type:application/json"
```

##### 响应：

```json
{"jsonrpc":"2.0","id":3,"result":"0x00"}
````


### 14. chain_getVoteCreditDetails
#### 作用：根据地址获取stake 所有细节信息
> 参数：
 1. 地址

#### 返回值：bytecode

#### 示例代码
##### 请求：

```shell
curl http://localhost:15645 -X POST --data '{"jsonrpc":"2.0","method":"chain_getCreditDetails","params":["0x8a8e541ddd1272d53729164c70197221a3c27486"], "id": 3}' -H "Content-Type:application/json"
```

##### 响应：

```json
{"jsonrpc":"2.0","id":3,"result":"[{\"Addr\":\"0xd05d5f324ada3c418e14cd6b497f2f36d60ba607\",\"HeghtValues\":[{\"CreditHeight\":1329,\"CreditValue\":\"0x11135\"}]}]"}
````


### 15. chain_GetCancelCreditDetails
#### 作用：获取所有退票请求的细节
> 参数：
 1. 地址

#### 返回值：bytecode

#### 示例代码
##### 请求：

```shell
curl http://localhost:15645 -X POST --data '{"jsonrpc":"2.0","method":"chain_getCancelCreditDetails","params":["0x8a8e541ddd1272d53729164c70197221a3c27486"], "id": 3}' -H "Content-Type:application/json"
```

##### 响应：

```json
{"jsonrpc":"2.0","id":3,"result":"{\"0x300fc5a14e578be28c64627c0e7e321771c58cd4\":\"0x3641100\"}"}
````


### 16. chain_GetCandidateAddrs
#### 作用：获取所有候选节点地址和对应的信任值
> 参数：
 1. 地址

#### 返回值：[]

#### 示例代码
##### 请求：

```shell
curl http://localhost:15645 -X POST --data '{"jsonrpc":"2.0","method":"chain_getCandidateAddrs","params":[""], "id": 3}' -H "Content-Type:application/json"
```

##### 响应：

```json
{"jsonrpc":"2.0","id":3,"result":"{\"0x300fc5a14e578be28c64627c0e7e321771c58cd4\":\"0x3641100\"}"}
````


### 17. chain_getInterestRate
#### 作用：获取3个月内、3-6个月、6-12个月、12个月以上的利率
> 参数：

#### 返回值：年华后三个月利息, 年华后六个月利息, 一年期利息, 一年以上期利息

#### 示例代码
##### 请求：

```shell
curl http://localhost:15645 -X POST --data '{"jsonrpc":"2.0","method":"chain_getInterestRate","params":"", "id": 3}' -H "Content-Type:application/json"
```

##### 响应：

```json
{"jsonrpc":"2.0","id":3,"result":"{\"ThreeMonthRate\":4,\"SixMonthRate\":12,\"OneYearRate\":25,\"MoreOneYearRate\":51}"}
````


### 18. chain_getChangeCycle
#### 作用：获取出块节点换届周期
> 参数：

#### 返回值：换届周期

#### 示例代码
##### 请求：

```shell
curl http://localhost:15645 -X POST --data '{"jsonrpc":"2.0","method":"chain_getChangeCycle","params":"", "id": 3}' -H "Content-Type:application/json"
```

##### 响应：

```json
{"jsonrpc":"2.0","id":3,"result":"{100}"}
````

p2p网络接口
设置查询网络状态

### 1. p2p_getPeers
#### 作用：获取当前连接的节点
> 参数：

#### 返回值：和本地建立连接的p2p对端信息

#### 示例代码
##### 请求：

```shell
curl http://127.0.0.1:15645 -X POST --data '{"jsonrpc":"2.0","method":"p2p_getPeers","params":"", "id": 3}' -H "Content-Type:application/json"
```

##### 响应：

```json
{"jsonrpc":"2.0","id":3,"result":[{},{},{},{}]}
````


### 2. p2p_addPeer
#### 作用：添加节点
> 参数：

#### 返回值：nil

#### 示例代码
##### 请求：

```shell
curl http://127.0.0.1:15645 -X POST --data '{"jsonrpc":"2.0","method":"p2p_addPeer","params":["enode://e1b2f83b7b0f5845cc74ca12bb40152e520842bbd0597b7770cb459bd40f109178811ebddd6d640100cdb9b661a3a43a9811d9fdc63770032a3f2524257fb62d@192.168.74.1:55555"], "id": 3}' -H "Content-Type:application/json"
```

##### 响应：

```json

````


### 3. p2p_removePeer
#### 作用：移除节点
> 参数：

#### 返回值：nil

#### 示例代码
##### 请求：

```shell
curl http://127.0.0.1:15645 -X POST --data '{"jsonrpc":"2.0","method":"p2p_removePeer","params":["enode://e1b2f83b7b0f5845cc74ca12bb40152e520842bbd0597b7770cb459bd40f109178811ebddd6d640100cdb9b661a3a43a9811d9fdc63770032a3f2524257fb62d@192.168.74.1:55555"], "id": 3}' -H "Content-Type:application/json"
```

##### 响应：

```json

````

日志rpc接口
设置日志级别

### 1. log_setLevel
#### 作用：设置日志级别
> 参数：
 1. 日志级别（&#34;debug&#34;,&#34;0&#34;）

#### 返回值：无

#### 示例代码
##### 请求：

```shell
curl http://localhost:15645 -X POST --data '{"jsonrpc":"2.0","method":"log_setLevel","params":["trace"], "id": 3}' -H "Content-Type:application/json"
```

##### 响应：

```json
{"jsonrpc":"2.0","id":3,"result":null}
````


### 2. log_setVmodule
#### 作用：分模块设置级别
> 参数：
 1. 模块日志级别(txpool=5)

#### 返回值：无

#### 示例代码
##### 请求：

```shell
curl http://localhost:15645 -X POST --data '{"jsonrpc":"2.0","method":"log_setVmodule","params":["txpool=5"], "id": 3}' -H "Content-Type:application/json"
```

##### 响应：

```json
{"jsonrpc":"2.0","id":3,"result":null}
````

记录接口
查询交易地址等信息（需要开启记录模块）

### 1. trace_getRawTransaction
#### 作用：根据交易hash查询交易字节
> 参数：
 1. 交易hash

#### 返回值：交易字节信息

#### 示例代码
##### 请求：

```shell
curl http://localhost:15645 -X POST --data '{"jsonrpc":"2.0","method":"trace_getRawTransaction","params":["0x00001c9b8c8fdb1f53faf02321f76253704123e2b56cce065852bab93e526ae2"], "id": 3}' -H "Content-Type:application/json"
```

##### 响应：

```json
{
	  "jsonrpc": "2.0",
	  "id": 3,
	  "result": "0x02a7ae20007923a30bbfbcb998a6534d56b313e68c8e0c594a0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002011102011003030000bc9889d00b004120eba14c77eab7a154833ff14832d8769cfc0b30db288445d6a83ef2fe337aa09042f8174a593543c4acabe7fadf1ad5fceea9c835682cb9dbea3f1d8fec181fb9"
	}
````


### 2. trace_getTransaction
#### 作用：根据交易hash查询交易详细信息
> 参数：
 1. 交易hash

#### 返回值：交易详细信息

#### 示例代码
##### 请求：

```shell
curl http://localhost:15645 -X POST --data '{"jsonrpc":"2.0","method":"trace_getTransaction","params":["0x00001c9b8c8fdb1f53faf02321f76253704123e2b56cce065852bab93e526ae2"], "id": 3}' -H "Content-Type:application/json"
```

##### 响应：

```json
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
````


### 3. trace_decodeTrasnaction
#### 作用：把交易字节信息反解析成交易详情
> 参数：
 1. 交易字节信息

#### 返回值：交易详情

#### 示例代码
##### 请求：

```shell
curl http://localhost:15645 -X POST --data '{"jsonrpc":"2.0","method":"trace_decodeTrasnaction","params":["0x02a7ae20007923a30bbfbcb998a6534d56b313e68c8e0c594a0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002011102011003030000bc9889d00b004120eba14c77eab7a154833ff14832d8769cfc0b30db288445d6a83ef2fe337aa09042f8174a593543c4acabe7fadf1ad5fceea9c835682cb9dbea3f1d8fec181fb9"], "id": 3}' -H "Content-Type:application/json"
```

##### 响应：

```json
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
````


### 4. trace_getSendTransactionByAddr
#### 作用：根据地址查询该地址发出的交易，支持分页
> 参数：
 1. 交易地址
 2. 分页号（从1开始）
 3. 页大小

#### 返回值：交易列表

#### 示例代码
##### 请求：

```shell
curl http://localhost:15645 -X POST --data '{"jsonrpc":"2.0","method":"trace_getSendTransactionByAddr","params":["0x7923a30bbfbcb998a6534d56b313e68c8e0c594a",1,10], "id": 3}' -H "Content-Type:application/json"
```

##### 响应：

```json
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
````


### 5. trace_getReceiveTransactionByAd
#### 作用：根据地址查询该地址接受的交易，支持分页
> 参数：
 1. 交易地址
 2. 分页号（从1开始）
 3. 页大小

#### 返回值：交易列表

#### 示例代码
##### 请求：

```shell
curl http://localhost:15645 -X POST --data '{"jsonrpc":"2.0","method":"trace_getReceiveTransactionByAddr","params":["0x3ebcbe7cb440dd8c52940a2963472380afbb56c5",1,10], "id": 3}' -H "Content-Type:application/json"
```

##### 响应：

```json
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
````


### 6. trace_rebuild
#### 作用：重建trace中的区块记录
> 参数：
 1. 起始块（包含）
 2. 终止块（不包含）

#### 返回值：

#### 示例代码
##### 请求：

```shell
curl http://localhost:15645 -X POST --data '{"jsonrpc":"2.0","method":"trace_rebuild","params":[1,10], "id": 3}' -H "Content-Type:application/json"
```

##### 响应：

```json
{"jsonrpc":"2.0","id":3,"result":null}
````

账号rpc接口
地址管理及发起简单交易

### 1. account_listAddress
#### 作用：列出所有本地地址
> 参数：

#### 返回值：地址数组

#### 示例代码
##### 请求：

```shell
curl http://localhost:15645 -X POST --data '{"jsonrpc":"2.0","method":"account_listAddress","params":[], "id": 3}' -H "Content-Type:application/json"
```

##### 响应：

```json
{"jsonrpc":"2.0","id":3,"result":["0x3296d3336895b5baaa0eca3df911741bd0681c3f","0x3ebcbe7cb440dd8c52940a2963472380afbb56c5"]}
````


### 2. account_createAccount
#### 作用：创建本地账号
> 参数：

#### 返回值：新账号地址信息

#### 示例代码
##### 请求：

```shell
curl http://localhost:15645 -X POST --data '{"jsonrpc":"2.0","method":"account_createAccount","params":[], "id": 3}' -H "Content-Type:application/json"
```

##### 响应：

```json
{"jsonrpc":"2.0","id":3,"result":"0x2944c15c466fad03ec1282bab579dec5a0cf0fa3"}
````


### 3. account_createWallet
#### 作用：创建本地钱包
> 参数：
 1. 钱包密码

#### 返回值：失败返回错误原因，成功不返回任何信息

#### 示例代码
##### 请求：

```shell
curl http://localhost:15645 -X POST --data '{"jsonrpc":"2.0","method":"account_createWallet","params":["123"], "id": 3}' -H "Content-Type:application/json"
```

##### 响应：

```json
{"jsonrpc":"2.0","id":3,"result":null}
````


### 4. account_lockAccount
#### 作用：锁定账号
> 参数：

#### 返回值：失败返回错误原因，成功不返回任何信息

#### 示例代码
##### 请求：

```shell
curl http://localhost:15645 -X POST --data '{"jsonrpc":"2.0","method":"account_lockAccount","params":["0x518b3fefa3fb9a72753c6ad10a2b68cc034ec391"], "id": 3}' -H "Content-Type:application/json"
```

##### 响应：

```json
{"jsonrpc":"2.0","id":3,"result":null}
````


### 5. account_account_unlockAccount
#### 作用：解锁账号
> 参数：
 1. 账号地址

#### 返回值：失败返回错误原因，成功不返回任何信息

#### 示例代码
##### 请求：

```shell
curl http://localhost:15645 -X POST --data '{"jsonrpc":"2.0","method":"account_unlockAccount","params":["0x518b3fefa3fb9a72753c6ad10a2b68cc034ec391"], "id": 3}' -H "Content-Type:application/json"
```

##### 响应：

```json
{"jsonrpc":"2.0","id":3,"result":null}
````


### 6. account_openWallet
#### 作用：打开钱包
> 参数：
 1. 钱包密码

#### 返回值：无

#### 示例代码
##### 请求：

```shell
curl http://localhost:15645 -X POST --data '{"jsonrpc":"2.0","method":"account_openWallet","params":["123"], "id": 3}' -H "Content-Type:application/json"
```

##### 响应：

```json
{"jsonrpc":"2.0","id":3,"result":null}
````


### 7. account_closeWallet
#### 作用：关闭钱包
> 参数：

#### 返回值：无

#### 示例代码
##### 请求：

```shell
curl http://localhost:15645 -X POST --data '{"jsonrpc":"2.0","method":"account_closeWallet","params":[], "id": 3}' -H "Content-Type:application/json"
```

##### 响应：

```json
{"jsonrpc":"2.0","id":3,"result":null}
````


### 8. account_transfer
#### 作用：转账
> 参数：
 1. 发起转账的地址
 2. 接受者的地址
 3. 金额
 4. gas价格
 5. gas上限
 6. 备注/data

#### 返回值：交易地址

#### 示例代码
##### 请求：

```shell
curl -H "Content-Type: application/json" -X post --data '{"jsonrpc":"2.0","method":"account_transfer","params":["0x3ebcbe7cb440dd8c52940a2963472380afbb56c5","0x3ebcbe7cb440dd8c52940a2963472380afbb56c5","0x111","0x110","0x30000",""],"id":1}' http://127.0.0.1:15645
```

##### 响应：

```json
{"jsonrpc":"2.0","id":1,"result":"0x3a3b59f90a21c2fd1b690aa3a2bc06dc2d40eb5bdc26fdd7ecb7e1105af2638e"}
````


### 9. account_transferWithNonce
#### 作用：转账
> 参数：
 1. 发起转账的地址
 2. 接受者的地址
 3. 金额
 4. gas价格
 5. gas上限
 6. 备注
 7. nonce

#### 返回值：交易地址

#### 示例代码
##### 请求：

```shell
curl -H "Content-Type: application/json" -X post --data '{"jsonrpc":"2.0","method":"account_transferWithNonce","params":["0x3ebcbe7cb440dd8c52940a2963472380afbb56c5","0x3ebcbe7cb440dd8c52940a2963472380afbb56c5","0x111","0x110","0x30000","",1],"id":1}' http://127.0.0.1:15645
```

##### 响应：

```json
{"jsonrpc":"2.0","id":1,"result":"0x3a3b59f90a21c2fd1b690aa3a2bc06dc2d40eb5bdc26fdd7ecb7e1105af2638e"}
````


### 10. account_setAlias
#### 作用：设置别名
> 参数：
 1. 带设置别名的地址
 2. 别名
 3. gas价格
 4. gas上限

#### 返回值：交易地址

#### 示例代码
##### 请求：

```shell
curl -H "Content-Type: application/json" -X post --data '{"jsonrpc":"2.0","method":"account_setAlias","params":["0x3ebcbe7cb440dd8c52940a2963472380afbb56c5","AAAAA","0x110","0x30000"],"id":1}' http://127.0.0.1:15645
```

##### 响应：

```json
{"jsonrpc":"2.0","id":1,"result":"0x5adb248f2943e12fb91c140bd3d0df6237712061e9abae97345b0869c3daa749"}
````


### 11. account_VoteCredit
#### 作用：投票
> 参数：
 1. 发起转账的地址
 2. 接受者的地址
 3. 金额
 4. gas价格
 5. gas上线
 6. 备注

#### 返回值：交易地址

#### 示例代码
##### 请求：

```shell
curl -H "Content-Type: application/json" -X post --data '{"jsonrpc":"2.0","method":"account_voteCredit","params":["0x3ebcbe7cb440dd8c52940a2963472380afbb56c5","0x3ebcbe7cb440dd8c52940a2963472380afbb56c5","0x111","0x110","0x30000"],"id":1}' http://127.0.0.1:15645
```

##### 响应：

```json
{"jsonrpc":"2.0","id":1,"result":"0x3a3b59f90a21c2fd1b690aa3a2bc06dc2d40eb5bdc26fdd7ecb7e1105af2638e"}
````


### 12. account_CancelVoteCredit
#### 作用：
> 参数：
 1. 发起转账的地址
 2. 接受者的地址
 3. 金额
 4. gas价格
 5. gas上线
 6. 备注

#### 返回值：交易地址

#### 示例代码
##### 请求：

```shell
curl -H "Content-Type: application/json" -X post --data '{"jsonrpc":"2.0","method":"account_cancelVoteCredit","params":["0x3ebcbe7cb440dd8c52940a2963472380afbb56c5","0x3ebcbe7cb440dd8c52940a2963472380afbb56c5","0x111","0x110","0x30000"],"id":1}' http://127.0.0.1:15645
```

##### 响应：

```json
{"jsonrpc":"2.0","id":1,"result":"0x3a3b59f90a21c2fd1b690aa3a2bc06dc2d40eb5bdc26fdd7ecb7e1105af2638e"}
````


### 13. account_CandidateCredit
#### 作用：候选节点质押
> 参数：
 1. 质押者的地址
 2. 质押金额
 3. gas价格
 4. gas上线
 5. 质押者地址对应的pubkey，质押者的P2p信息

#### 返回值：交易地址

#### 示例代码
##### 请求：

```shell
curl -H "Content-Type: application/json" -X post --data '{"jsonrpc":"2.0","method":"account_candidateCredit","params":["0x3ebcbe7cb440dd8c52940a2963472380afbb56c5","0x111","0x110","0x30000","{\"Pubkey\":\"0x020e233ebaed5ade5e48d7ee7a999e173df054321f4ddaebecdb61756f8a43e91c\",\"Node\":\"enode://3f05da2475bf09ce20b790d76b42450996bc1d3c113a1848be1960171f9851c0@149.129.172.91:44444\"}"],"id":1}' http://127.0.0.1:15645
```

##### 响应：

```json
{"jsonrpc":"2.0","id":1,"result":"0x3a3b59f90a21c2fd1b690aa3a2bc06dc2d40eb5bdc26fdd7ecb7e1105af2638e"}
````


### 14. account_CancelCandidateCredit
#### 作用：取消候选
> 参数：
 1. 发起转账的地址
 2. 接受者的地址
 3. 金额
 4. gas价格
 5. gas上线
 6. 备注

#### 返回值：交易地址

#### 示例代码
##### 请求：

```shell
curl -H "Content-Type: application/json" -X post --data '{"jsonrpc":"2.0","method":"account_cancelCandidateCredit","params":["0x3ebcbe7cb440dd8c52940a2963472380afbb56c5","0x111","0x110","0x30000",""],"id":1}' http://127.0.0.1:15645
```

##### 响应：

```json
{"jsonrpc":"2.0","id":1,"result":"0x3a3b59f90a21c2fd1b690aa3a2bc06dc2d40eb5bdc26fdd7ecb7e1105af2638e"}
````


### 15. account_readContract
#### 作用：读取智能合约（无数据被修改）
> 参数：
 1. 发交易的账户地址
 2. 合约地址
 3. 合约接口

#### 返回值：查询结果

#### 示例代码
##### 请求：

```shell
curl -H "Content-Type: application/json" -X post --data '{"jsonrpc":"2.0","method":"account_readContract","params":["0xec61c03f719a5c214f60719c3f36bb362a202125","0xecfb51e10aa4c146bf6c12eee090339c99841efc","0x6d4ce63c"],"id":1}' http://127.0.0.1:15645
```

##### 响应：

```json
{"jsonrpc":"2.0","id":1,"result":""}
````


### 16. account_estimateGas
#### 作用：估算交易需要多少gas
> 参数：
 1. 发起转账的地址
 2. 金额
 3. 备注/data
 4. 接受者的地址

#### 返回值：评估结果，失败返回错误

#### 示例代码
##### 请求：

```shell
curl -H "Content-Type: application/json" -X post --data '{"jsonrpc":"2.0","method":"account_estimateGas","params":["0xec61c03f719a5c214f60719c3f36bb362a202125","0xecfb51e10aa4c146bf6c12eee090339c99841efc","0x6d4ce63c","0x110","0x30000"],"id":1}' http://127.0.0.1:15645
```

##### 响应：

```json
{"jsonrpc":"2.0","id":1,"result":"0x5d74aba54ace5f01a5f0057f37bfddbbe646ea6de7265b368e2e7d17d9cdeb9c"}
````


### 17. account_executeContract
#### 作用：执行智能合约（导致数据被修改）
> 参数：
 1. 调用者的地址
 2. 合约地址
 3. 代码
 4. gas价格
 5. gas上限

#### 返回值：交易hash

#### 示例代码
##### 请求：

```shell
curl -H "Content-Type: application/json" -X post --data '{"jsonrpc":"2.0","method":"account_executeContract","params":["0xec61c03f719a5c214f60719c3f36bb362a202125","0xecfb51e10aa4c146bf6c12eee090339c99841efc","0x6d4ce63c","0x110","0x30000"],"id":1}' http://127.0.0.1:15645
```

##### 响应：

```json
{"jsonrpc":"2.0","id":1,"result":"0x5d74aba54ace5f01a5f0057f37bfddbbe646ea6de7265b368e2e7d17d9cdeb9c"}
````


### 18. account_createCode
#### 作用：部署合约
> 参数：
 1. 部署合约的地址
 2. 合约内容
 3. 金额
 4. gas价格
 5. gas上限

#### 返回值：合约地址

#### 示例代码
##### 请求：

```shell
curl -H "Content-Type: application/json" -X post --data '{"jsonrpc":"2.0","method":"account_createCode","params":["0x3ebcbe7cb440dd8c52940a2963472380afbb56c5","0x608060405234801561001057600080fd5b5061018c806100206000396000f3fe608060405260043610610051576000357c0100000000000000000000000000000000000000000000000000000000900480634f2be91f146100565780636d4ce63c1461006d578063db7208e31461009e575b600080fd5b34801561006257600080fd5b5061006b6100dc565b005b34801561007957600080fd5b5061008261011c565b604051808260070b60070b815260200191505060405180910390f35b3480156100aa57600080fd5b506100da600480360360208110156100c157600080fd5b81019080803560070b9060200190929190505050610132565b005b60016000808282829054906101000a900460070b0192506101000a81548167ffffffffffffffff021916908360070b67ffffffffffffffff160217905550565b60008060009054906101000a900460070b905090565b806000806101000a81548167ffffffffffffffff021916908360070b67ffffffffffffffff1602179055505056fea165627a7a723058204b651e4313ab6bc4eda61084cac1f805699cefbb979ddfd3a2d7f970903307cd0029","0x111","0x110","0x30000"],"id":1}' http://127.0.0.1:15645
```

##### 响应：

```json
{"jsonrpc":"2.0","id":1,"result":"0x9a8d8d5d7d00bbe0eb1b9431a13a7219008e352241b751b177bfb29e4e75b0d1"}
````


### 19. account_dumpPrivkey
#### 作用：导出地址对应的私钥
> 参数：
 1. 地址

#### 返回值：私钥

#### 示例代码
##### 请求：

```shell
curl http://localhost:15645 -X POST --data '{"jsonrpc":"2.0","method":"account_dumpPrivkey","params":["0x3ebcbe7cb440dd8c52940a2963472380afbb56c5"], "id": 3}' -H "Content-Type:application/json"
```

##### 响应：

```json
{"jsonrpc":"2.0","id":3,"result":"0x270f4b122603999d1c07aec97e972a2ddf7bd8b5bfe3543c10814e6a19f13aaf"}
````


### 20. account_DumpPubkey
#### 作用：导出地址对应的公钥
> 参数：
 1. 地址

#### 返回值：公钥

#### 示例代码
##### 请求：

```shell
curl http://localhost:15645 -X POST --data '{"jsonrpc":"2.0","method":"account_dumpPubkey","params":["0x3ebcbe7cb440dd8c52940a2963472380afbb56c5"], "id": 3}' -H "Content-Type:application/json"
```

##### 响应：

```json
{"jsonrpc":"2.0","id":3,"result":"0x270f4b122603999d1c07aec97e972a2ddf7bd8b5bfe3543c10814e6a19f13aaf"}
````


### 21. account_sign
#### 作用：关闭钱包
> 参数：
 1. 地址
 2. 消息hash

#### 返回值：私钥

#### 示例代码
##### 请求：

```shell
curl http://localhost:15645 -X POST --data '{"jsonrpc":"2.0","method":"account_sign","params":["0x3ebcbe7cb440dd8c52940a2963472380afbb56c5", "0x00001c9b8c8fdb1f53faf02321f76253704123e2b56cce065852bab93e526ae2"], "id": 3}' -H "Content-Type:application/json"
```

##### 响应：

```json
{"jsonrpc":"2.0","id":3,"result":"0x1f1d16412468dd9b67b568d31839ac608bdfddf2580666db4d364eefbe285fdaed569a3c8fa1decfebbfa0ed18b636059dbbf4c2106c45fc8846909833ef2cb1de"}
````


### 22. account_generateAddresses
#### 作用：生成其他链的地址
> 参数：
 1. drep地址

#### 返回值：{BTCaddress, ethAddress, neoAddress}

#### 示例代码
##### 请求：

```shell
curl http://localhost:15645 -X POST --data '{"jsonrpc":"2.0","method":"account_generateAddresses","params":["0x3ebcbe7cb440dd8c52940a2963472380afbb56c5"], "id": 3}' -H "Content-Type:application/json"
```

##### 响应：

```json
{"jsonrpc":"2.0","id":3,"result":""}
````


### 23. account_importKeyStore
#### 作用：导入keystore
> 参数：
 1. path
 2. password

#### 返回值：address list

#### 示例代码
##### 请求：

```shell
curl http://localhost:15645 -X POST --data '{"jsonrpc":"2.0","method":"account_importKeyStore","params":["path","123"], "id": 3}' -H "Content-Type:application/json"
```

##### 响应：

```json
{"jsonrpc":"2.0","id":3,"result":["0x4082c96e38def8f3851831940485066234fe07b8"]}
````


### 24. account_importPrivkey
#### 作用：导入私钥
> 参数：
 1. privkey(compress hex)

#### 返回值：address

#### 示例代码
##### 请求：

```shell
curl http://localhost:15645 -X POST --data '{"jsonrpc":"2.0","method":"account_importPrivkey","params":["0xe5510b32854ca52e7d7d41bb3196fd426d551951e2fd5f6b559a62889d87926c"], "id": 3}' -H "Content-Type:application/json"
```

##### 响应：

```json
{"jsonrpc":"2.0","id":3,"result":"0x748eb65493a964e568800c3c2885c63a0de9f9ae"}
````

共识rpc接口
查询共识节点功能

### 1. consensus_changeWaitTime
#### 作用：修改leader等待时间 (ms)
> 参数：
 1. 等待时间(ms)

#### 返回值：私钥

#### 示例代码
##### 请求：

```shell
curl http://localhost:15645 -X POST --data '{"jsonrpc":"2.0","method":"consensus_changeWaitTime","params":[100000], "id": 3}' -H "Content-Type:application/json"
```

##### 响应：

```json
{"jsonrpc":"2.0","id":3,"result":null}
````


### 2. consensus_getMiners()
#### 作用：获取当前出块节点
> 参数：

#### 返回值：出块节点信息

#### 示例代码
##### 请求：

```shell
curl http://localhost:15645 -X POST --data '{"jsonrpc":"2.0","method":"consensus_getMiners","params":[""], "id": 3}' -H "Content-Type:application/json"
```

##### 响应：

```json
{"jsonrpc":"2.0","id":3,"result":['0x02c682c9f503465a27d1941d1a25547b5ea879a7145056283599a33869982513df', '0x036a09f9012cb3f73c11ceb2aae4242265c2aa35ebec20dbc28a78712802f457db']
}
````

