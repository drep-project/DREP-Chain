
#  `JSON-RPC`RPC interface
## Block management
For processing block chain partial upper logic

### 1. blockMgr_sendRawTransaction
#### usage：Send signed transactions
> params：
 1. A signed transaction

#### return：transaction hash

#### example

```shell
curl http://localhost:10085 -X POST --data '{"jsonrpc":"2.0","method":"blockmgr_sendRawTransaction","params":["0x40a287b6d30b05313131317a4120dd8c23c40910d038fa43b2f8932d3681cbe5ee3079b6e9de0bea6e8e6b2a867a561aa26e1cd6b62aa0422a043186b593b784bf80845c3fd5a7fbfe62e61d8564"], "id": 3}' -H "Content-Type:application/json"
```

##### response：

```json
{"jsonrpc":"2.0","id":1,"result":"0xf30e858667fa63bc57ae395c3f57ede9bb3ad4969d12f4bce51d900fb5931538"}
````


### 2. blockMgr_gasPrice
#### usage：Get the recommended value of gasprice given by the system
> params：
 1. Query address

#### return：Price and error message

#### example

```shell
curl http://localhost:10085 -X POST --data '{"jsonrpc":"2.0","method":"blockmgr_gasPrice","params":[], "id": 3}' -H "Content-Type:application/json"
```

##### response：

```json

````


### 3. blockMgr_GetPoolTransactions
#### usage：Get trading information in the trading pool.
> params：
 1. Query address

#### return：All transactions in the pool

#### example

```shell
curl http://localhost:10085 -X POST --data '{"jsonrpc":"2.0","method":"blockmgr_getPoolTransactions","params":["0x8a8e541ddd1272d53729164c70197221a3c27486"], "id": 3}' -H "Content-Type:application/json"
```

##### response：

```json

````


### 4. blockMgr_GetTransactionCount
#### usage：Gets the total number of transactions issued by the address
> params：
 1. Query address

#### return：All transactions in the pool

#### example

```shell
curl http://localhost:10085 -X POST --data '{"jsonrpc":"2.0","method":"blockmgr_getTransactionCount","params":["0x8a8e541ddd1272d53729164c70197221a3c27486"], "id": 3}' -H "Content-Type:application/json"
```

##### response：

```json

````


### 5. blockMgr_GetPoolMiniPendingNonce
#### usage：Get the smallest Nonce in the pending queue
> params：
 1. Query address

#### return：The smallest nonce in the pending queue

#### example

```shell
curl http://localhost:10085 -X POST --data '{"jsonrpc":"2.0","method":"blockmgr_getPoolMiniPendingNonce","params":["0x8a8e541ddd1272d53729164c70197221a3c27486"], "id": 3}' -H "Content-Type:application/json"
```

##### response：

```json

````


### 6. blockMgr_GetTxInPool
#### usage：Checks whether the transaction is in the trading pool and, if so, returns the transaction
> params：
 1. The address at which the transfer was initiated

#### return：Complete transaction information

#### example

```shell
curl -H "Content-Type: application/json" -X post --data '{"jsonrpc":"2.0","method":"blockmgr_getTxInPool","params":["0x3ebcbe7cb440dd8c52940a2963472380afbb56c5"],"id":1}' http://127.0.0.1:10085
```

##### response：

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

Block chain API
Used to obtain block information

### 1. chain_getblock
#### usage：Used to obtain block information
> params：
 1. height  usage: Current block height

#### return：Block detail information

#### example

```shell
curl http://localhost:10085 -X POST --data '{"jsonrpc":"2.0","method":"chain_getBlock","params":[1], "id": 3}' -H "Content-Type:application/json"
```

##### response：

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
#### usage：To get the current highest block
> params：
 1. 无

#### return：Current maximum block height value

#### example

```shell
curl http://localhost:10085 -X POST --data '{"jsonrpc":"2.0","method":"chain_getMaxHeight","params":[], "id": 3}' -H "Content-Type:application/json"
```

##### response：

```json
{"jsonrpc":"2.0","id":3,"result":193005}
````


### 3. chain_getBlockGasInfo
#### usage：Obtain gas related information
> params：
 1. 无

#### return：Gas minimum value and maximum value required by the system; And the maximum gas value that the current block is set to

#### example

```shell
curl http://localhost:10085 -X POST --data '{"jsonrpc":"2.0","method":"chain_getBlockGasInfo","params":[], "id": 3}' -H "Content-Type:application/json"
```

##### response：

```json
{"jsonrpc":"2.0","id":3,"result":193005}
````


### 4. chain_getBalance
#### usage：Query address balance
> params：
 1. Query address

#### return：The account balance in the address

#### example

```shell
curl http://localhost:10085 -X POST --data '{"jsonrpc":"2.0","method":"chain_getBalance","params":["0x8a8e541ddd1272d53729164c70197221a3c27486"], "id": 3}' -H "Content-Type:application/json"
```

##### response：

```json
{"jsonrpc":"2.0","id":3,"result":9987999999999984000000}
````


### 5. chain_getNonce
#### usage：Query the nonce whose address is on the chain
> params：
 1. Query address

#### return：nonce

#### example

```shell
curl http://localhost:10085 -X POST --data '{"jsonrpc":"2.0","method":"chain_getNonce","params":["0x8a8e541ddd1272d53729164c70197221a3c27486"], "id": 3}' -H "Content-Type:application/json"
```

##### response：

```json
{"jsonrpc":"2.0","id":3,"result":0}
````


### 6. chain_GetReputation
#### usage：Query the reputation value of the address
> params：
 1. Query address

#### return：The reputation value corresponding to the address

#### example

```shell
curl http://localhost:10085 -X POST --data '{"jsonrpc":"2.0","method":"chain_getReputation","params":["0x8a8e541ddd1272d53729164c70197221a3c27486"], "id": 3}' -H "Content-Type:application/json"
```

##### response：

```json
{"jsonrpc":"2.0","id":3,"result":1}
````


### 7. chain_getTransactionByBlockHeightAndIndex
#### usage：Gets a particular sequence of transactions in a block
> params：
 1. block height
 2. Transaction sequence

#### return：transaction

#### example

```shell
curl http://localhost:10085 -X POST --data '{"jsonrpc":"2.0","method":"chain_getTransactionByBlockHeightAndIndex","params":[10000,1], "id": 3}' -H "Content-Type:application/json"
```

##### response：

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
#### usage：Gets the alias corresponding to the address according to the address
> params：
 1. address

#### return：Address the alias

#### example

```shell
curl http://localhost:10085 -X POST --data '{"jsonrpc":"2.0","method":"chain_getAliasByAddress","params":["0x8a8e541ddd1272d53729164c70197221a3c27486"], "id": 3}' -H "Content-Type:application/json"
```

##### response：

```json
{"jsonrpc":"2.0","id":3,"result":"tom"}
````


### 9. chain_getAddressByAlias
#### usage：Gets the address corresponding to the alias based on the alias
> params：
 1. Alias to be queried

#### return：The address corresponding to the alias

#### example

```shell
curl http://localhost:10085 -X POST --data '{"jsonrpc":"2.0","method":"chain_getAddressByAlias","params":["tom"], "id": 3}' -H "Content-Type:application/json"
```

##### response：

```json
{"jsonrpc":"2.0","id":3,"result":"0x8a8e541ddd1272d53729164c70197221a3c27486"}
````


### 10. chain_getReceipt
#### usage：Get the receipt information based on txhash
> params：
 1. txhash

#### return：receipt

#### example

```shell
curl http://localhost:10085 -X POST --data '{"jsonrpc":"2.0","method":"chain_getReceipt","params":["0x7d9dd32ca192e765ff2abd7c5f8931cc3f77f8f47d2d52170c7804c2ca2c5dd9"], "id": 3}' -H "Content-Type:application/json"
```

##### response：

```json
{"jsonrpc":"2.0","id":3,"result":""}
````


### 11. chain_getLogs
#### usage：Get the transaction log information based on txhash
> params：
 1. txhash

#### return：[]log

#### example

```shell
curl http://localhost:10085 -X POST --data '{"jsonrpc":"2.0","method":"chain_getLogs","params":["0x7d9dd32ca192e765ff2abd7c5f8931cc3f77f8f47d2d52170c7804c2ca2c5dd9"], "id": 3}' -H "Content-Type:application/json"
```

##### response：

```json
{"jsonrpc":"2.0","id":3,"result":""}
````


### 12. chain_getCancelCreditDetail
#### usage：Get the back pledge or back vote information according to txhash
> params：
 1. txhash

#### return：{}

#### example

```shell
curl http://localhost:10085 -X POST --data '{"jsonrpc":"2.0","method":"chain_getCancelCreditDetail","params":["0x7d9dd32ca192e765ff2abd7c5f8931cc3f77f8f47d2d52170c7804c2ca2c5dd9"], "id": 3}' -H "Content-Type:application/json"
```

##### response：

```json
{"jsonrpc":"2.0","id":3,"result":""}
````


### 13. chain_getByteCode
#### usage：Get bytecode by address
> params：
 1. address

#### return：bytecode

#### example

```shell
curl http://localhost:10085 -X POST --data '{"jsonrpc":"2.0","method":"chain_getByteCode","params":["0x8a8e541ddd1272d53729164c70197221a3c27486"], "id": 3}' -H "Content-Type:application/json"
```

##### response：

```json
{"jsonrpc":"2.0","id":3,"result":"0x00"}
````


### 14. chain_getVoteCreditDetails
#### usage：Get all the details of the stake according to the address
> params：
 1. address

#### return：bytecode

#### example

```shell
curl http://localhost:10085 -X POST --data '{"jsonrpc":"2.0","method":"chain_getCreditDetails","params":["0x8a8e541ddd1272d53729164c70197221a3c27486"], "id": 3}' -H "Content-Type:application/json"
```

##### response：

```json
{"jsonrpc":"2.0","id":3,"result":"[{\"Addr\":\"0xd05d5f324ada3c418e14cd6b497f2f36d60ba607\",\"HeghtValues\":[{\"CreditHeight\":1329,\"CreditValue\":\"0x11135\"}]}]"}
````


### 15. chain_GetCancelCreditDetails
#### usage：Get the details of all refund requests
> params：
 1. address

#### return：bytecode

#### example

```shell
curl http://localhost:10085 -X POST --data '{"jsonrpc":"2.0","method":"chain_getCancelCreditDetails","params":["0x8a8e541ddd1272d53729164c70197221a3c27486"], "id": 3}' -H "Content-Type:application/json"
```

##### response：

```json
{"jsonrpc":"2.0","id":3,"result":"{\"0x300fc5a14e578be28c64627c0e7e321771c58cd4\":\"0x3641100\"}"}
````


### 16. chain_GetCandidateAddrs
#### usage：Gets the addresses of all candidate nodes and the corresponding trust values
> params：
 1. address

#### return：[]

#### example

```shell
curl http://localhost:10085 -X POST --data '{"jsonrpc":"2.0","method":"chain_getCandidateAddrs","params":[""], "id": 3}' -H "Content-Type:application/json"
```

##### response：

```json
{"jsonrpc":"2.0","id":3,"result":"{\"0x300fc5a14e578be28c64627c0e7e321771c58cd4\":\"0x3641100\"}"}
````


### 17. chain_getChangeCycle
#### usage：Gets the transition period of the out - of - block node
> params：

#### return：Transition period

#### example

```shell
curl http://localhost:10085 -X POST --data '{"jsonrpc":"2.0","method":"chain_getChangeCycle","params":"", "id": 3}' -H "Content-Type:application/json"
```

##### response：

```json
{"jsonrpc":"2.0","id":3,"result":"{100}"}
````

p2p网络接口
设置查询网络状态

### 1. p2p_getPeers
#### usage：获取当前连接的节点
> params：

#### return：和本地建立连接的p2p对端信息

#### example

```shell
curl http://127.0.0.1:10085 -X POST --data '{"jsonrpc":"2.0","method":"p2p_getPeers","params":"", "id": 3}' -H "Content-Type:application/json"
```

##### response：

```json
{"jsonrpc":"2.0","id":3,"result":[{},{},{},{}]}
````


### 2. p2p_addPeer
#### usage：添加节点
> params：

#### return：nil

#### example

```shell
curl http://127.0.0.1:10085 -X POST --data '{"jsonrpc":"2.0","method":"p2p_addPeer","params":["enode://e1b2f83b7b0f5845cc74ca12bb40152e520842bbd0597b7770cb459bd40f109178811ebddd6d640100cdb9b661a3a43a9811d9fdc63770032a3f2524257fb62d@192.168.74.1:55555"], "id": 3}' -H "Content-Type:application/json"
```

##### response：

```json

````


### 3. p2p_removePeer
#### usage：移除节点
> params：

#### return：nil

#### example

```shell
curl http://127.0.0.1:10085 -X POST --data '{"jsonrpc":"2.0","method":"p2p_removePeer","params":["enode://e1b2f83b7b0f5845cc74ca12bb40152e520842bbd0597b7770cb459bd40f109178811ebddd6d640100cdb9b661a3a43a9811d9fdc63770032a3f2524257fb62d@192.168.74.1:55555"], "id": 3}' -H "Content-Type:application/json"
```

##### response：

```json

````

Logging RPC Api
Set the log level

### 1. log_setLevel
#### usage：Set the log level
> params：
 1. log level（&#34;debug&#34;,&#34;0&#34;）

#### return：无

#### example

```shell
curl http://localhost:10085 -X POST --data '{"jsonrpc":"2.0","method":"log_setLevel","params":["trace"], "id": 3}' -H "Content-Type:application/json"
```

##### response：

```json
{"jsonrpc":"2.0","id":3,"result":null}
````


### 2. log_setVmodule
#### usage：Set the level by module
> params：
 1. module name (txpool=5)

#### return：无

#### example

```shell
curl http://localhost:10085 -X POST --data '{"jsonrpc":"2.0","method":"log_setVmodule","params":["txpool=5"], "id": 3}' -H "Content-Type:application/json"
```

##### response：

```json
{"jsonrpc":"2.0","id":3,"result":null}
````

记录接口
查询交易地址等信息（需要开启记录模块）

### 1. trace_getRawTransaction
#### usage：根据交易hash查询交易字节
> params：
 1. 交易hash

#### return：交易字节信息

#### example

```shell
curl http://localhost10085 -X POST --data '{"jsonrpc":"2.0","method":"trace_getRawTransaction","params":["0x00001c9b8c8fdb1f53faf02321f76253704123e2b56cce065852bab93e526ae2"], "id": 3}' -H "Content-Type:application/json"
```

##### response：

```json
{
	  "jsonrpc": "2.0",
	  "id": 3,
	  "result": "0x02a7ae20007923a30bbfbcb998a6534d56b313e68c8e0c594a0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002011102011003030000bc9889d00b004120eba14c77eab7a154833ff14832d8769cfc0b30db288445d6a83ef2fe337aa09042f8174a593543c4acabe7fadf1ad5fceea9c835682cb9dbea3f1d8fec181fb9"
	}
````


### 2. trace_getTransaction
#### usage：根据交易hash查询交易详细信息
> params：
 1. 交易hash

#### return：交易详细信息

#### example

```shell
curl http://localhost10085 -X POST --data '{"jsonrpc":"2.0","method":"trace_getTransaction","params":["0x00001c9b8c8fdb1f53faf02321f76253704123e2b56cce065852bab93e526ae2"], "id": 3}' -H "Content-Type:application/json"
```

##### response：

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
#### usage：把交易字节信息反解析成交易详情
> params：
 1. 交易字节信息

#### return：交易详情

#### example

```shell
curl http://localhost10085 -X POST --data '{"jsonrpc":"2.0","method":"trace_decodeTrasnaction","params":["0x02a7ae20007923a30bbfbcb998a6534d56b313e68c8e0c594a0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002011102011003030000bc9889d00b004120eba14c77eab7a154833ff14832d8769cfc0b30db288445d6a83ef2fe337aa09042f8174a593543c4acabe7fadf1ad5fceea9c835682cb9dbea3f1d8fec181fb9"], "id": 3}' -H "Content-Type:application/json"
```

##### response：

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
#### usage：根据地址查询该地址发出的交易，支持分页
> params：
 1. 交易地址
 2. 分页号（从1开始）
 3. 页大小

#### return：交易列表

#### example

```shell
curl http://localhost10085 -X POST --data '{"jsonrpc":"2.0","method":"trace_getSendTransactionByAddr","params":["0x7923a30bbfbcb998a6534d56b313e68c8e0c594a",1,10], "id": 3}' -H "Content-Type:application/json"
```

##### response：

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
#### usage：根据地址查询该地址接受的交易，支持分页
> params：
 1. 交易地址
 2. 分页号（从1开始）
 3. 页大小

#### return：交易列表

#### example

```shell
curl http://localhost10085 -X POST --data '{"jsonrpc":"2.0","method":"trace_getReceiveTransactionByAddr","params":["0x3ebcbe7cb440dd8c52940a2963472380afbb56c5",1,10], "id": 3}' -H "Content-Type:application/json"
```

##### response：

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
#### usage：重建trace中的区块记录
> params：
 1. 起始块（包含）
 2. 终止块（不包含）

#### return：

#### example

```shell
curl http://localhost10085 -X POST --data '{"jsonrpc":"2.0","method":"trace_rebuild","params":[1,10], "id": 3}' -H "Content-Type:application/json"
```

##### response：

```json
{"jsonrpc":"2.0","id":3,"result":null}
````

Account RPC interface
Address management and initiate simple transactions

### 1. account_listAddress
#### usage：Lists all local addresses
> params：

#### return：Address of the array

#### example

```shell
curl http://localhost10085 -X POST --data '{"jsonrpc":"2.0","method":"account_listAddress","params":[], "id": 3}' -H "Content-Type:application/json"
```

##### response：

```json
{"jsonrpc":"2.0","id":3,"result":["0x3296d3336895b5baaa0eca3df911741bd0681c3f","0x3ebcbe7cb440dd8c52940a2963472380afbb56c5"]}
````


### 2. account_createAccount
#### usage：Create a local account
> params：
 1. password

#### return：New account address information

#### example

```shell
curl http://localhost10085 -X POST --data '{"jsonrpc":"2.0","method":"account_createAccount","params":["123456"], "id": 3}' -H "Content-Type:application/json"
```

##### response：

```json
{"jsonrpc":"2.0","id":3,"result":"0x2944c15c466fad03ec1282bab579dec5a0cf0fa3"}
````


### 3. account_createWallet
#### usage：Create a local wallet
> params：
 1. The wallet password

#### return：Failure returns the reason for the error, and success returns no information

#### example

```shell
curl http://localhost10085 -X POST --data '{"jsonrpc":"2.0","method":"account_createWallet","params":["123"], "id": 3}' -H "Content-Type:application/json"
```

##### response：

```json
{"jsonrpc":"2.0","id":3,"result":null}
````


### 4. account_lockAccount
#### usage：Lock the account
> params：

#### return：Failure returns the reason for the error, and success returns no information

#### example

```shell
curl http://localhost10085 -X POST --data '{"jsonrpc":"2.0","method":"account_lockAccount","params":["0x518b3fefa3fb9a72753c6ad10a2b68cc034ec391"], "id": 3}' -H "Content-Type:application/json"
```

##### response：

```json
{"jsonrpc":"2.0","id":3,"result":null}
````


### 5. account_account_unlockAccount
#### usage：Unlock the account
> params：
 1. The account address
 2. password

#### return：Failure returns the reason for the error, and success returns no information

#### example

```shell
curl http://localhost10085 -X POST --data '{"jsonrpc":"2.0","method":"account_unlockAccount","params":["0x518b3fefa3fb9a72753c6ad10a2b68cc034ec391", "123456"], "id": 3}' -H "Content-Type:application/json"
```

##### response：

```json
{"jsonrpc":"2.0","id":3,"result":null}
````


### 6. account_openWallet
#### usage：Open my wallet
> params：
 1. The wallet password

#### return：error or none

#### example

```shell
curl http://localhost10085 -X POST --data '{"jsonrpc":"2.0","method":"account_openWallet","params":["123"], "id": 3}' -H "Content-Type:application/json"
```

##### response：

```json
{"jsonrpc":"2.0","id":3,"result":null}
````


### 7. account_closeWallet
#### usage：close wallet
> params：

#### return：none

#### example

```shell
curl http://localhost10085 -X POST --data '{"jsonrpc":"2.0","method":"account_closeWallet","params":[], "id": 3}' -H "Content-Type:application/json"
```

##### response：

```json
{"jsonrpc":"2.0","id":3,"result":null}
````


### 8. account_transfer
#### usage：transfer
> params：
 1. The address at which the transfer was initiated
 2. Recipient&#39;s address
 3. Mount
 4. gas price
 5. gas limit
 6. commit

#### return：transaction hash

#### example

```shell
curl -H "Content-Type: application/json" -X post --data '{"jsonrpc":"2.0","method":"account_transfer","params":["0x3ebcbe7cb440dd8c52940a2963472380afbb56c5","0x3ebcbe7cb440dd8c52940a2963472380afbb56c5","0x111","0x110","0x30000",""],"id":1}' http://127.0.0.110085
```

##### response：

```json
{"jsonrpc":"2.0","id":1,"result":"0x3a3b59f90a21c2fd1b690aa3a2bc06dc2d40eb5bdc26fdd7ecb7e1105af2638e"}
````


### 9. account_transferWithNonce
#### usage：transfer with nonce
> params：
 1. The address at which the transfer was initiated
 2. Recipient&#39;s address
 3. Mount
 4. gas price
 5. gas limit
 6. commit
 7. nonce

#### return：transaction hash

#### example

```shell
curl -H "Content-Type: application/json" -X post --data '{"jsonrpc":"2.0","method":"account_transferWithNonce","params":["0x3ebcbe7cb440dd8c52940a2963472380afbb56c5","0x3ebcbe7cb440dd8c52940a2963472380afbb56c5","0x111","0x110","0x30000","",1],"id":1}' http://127.0.0.110085
```

##### response：

```json
{"jsonrpc":"2.0","id":1,"result":"0x3a3b59f90a21c2fd1b690aa3a2bc06dc2d40eb5bdc26fdd7ecb7e1105af2638e"}
````


### 10. account_setAlias
#### usage：Set an alias
> params：
 1. address
 2. alias
 3. gas price
 4. gas lowLimit

#### return：transaction hash

#### example

```shell
curl -H "Content-Type: application/json" -X post --data '{"jsonrpc":"2.0","method":"account_setAlias","params":["0x3ebcbe7cb440dd8c52940a2963472380afbb56c5","AAAAA","0x110","0x30000"],"id":1}' http://127.0.0.110085
```

##### response：

```json
{"jsonrpc":"2.0","id":1,"result":"0x5adb248f2943e12fb91c140bd3d0df6237712061e9abae97345b0869c3daa749"}
````


### 11. account_VoteCredit
#### usage：vote credit to candidate
> params：
 1. address of voter
 2. address of candidate
 3. amount
 4. gas price
 5. gas uplimit of transaction

#### return：transaction hash

#### example

```shell
curl -H "Content-Type: application/json" -X post --data '{"jsonrpc":"2.0","method":"account_voteCredit","params":["0x3ebcbe7cb440dd8c52940a2963472380afbb56c5","0x3ebcbe7cb440dd8c52940a2963472380afbb56c5","0x111","0x110","0x30000"],"id":1}' http://127.0.0.110085
```

##### response：

```json
{"jsonrpc":"2.0","id":1,"result":"0x3a3b59f90a21c2fd1b690aa3a2bc06dc2d40eb5bdc26fdd7ecb7e1105af2638e"}
````


### 12. account_CancelVoteCredit
#### usage：
> params：
 1. address of voter
 2. address of candidate
 3. amount
 4. gas price
 5. gas limit
 6. 备注

#### return：transaction hash

#### example

```shell
curl -H "Content-Type: application/json" -X post --data '{"jsonrpc":"2.0","method":"account_cancelVoteCredit","params":["0x3ebcbe7cb440dd8c52940a2963472380afbb56c5","0x3ebcbe7cb440dd8c52940a2963472380afbb56c5","0x111","0x110","0x30000"],"id":1}' http://127.0.0.110085
```

##### response：

```json
{"jsonrpc":"2.0","id":1,"result":"0x3a3b59f90a21c2fd1b690aa3a2bc06dc2d40eb5bdc26fdd7ecb7e1105af2638e"}
````


### 13. account_CandidateCredit
#### usage：Candidate node pledge
> params：
 1. The address of the pledger
 2. The pledge amount
 3. gas price
 4. gas limit
 5. The pubkey corresponding to the address of the pledger, and the P2p information of the pledger

#### return：transaction hash

#### example

```shell
curl -H "Content-Type: application/json" -X post --data '{"jsonrpc":"2.0","method":"account_candidateCredit","params":["0x3ebcbe7cb440dd8c52940a2963472380afbb56c5","0x111","0x110","0x30000","{\"Pubkey\":\"0x020e233ebaed5ade5e48d7ee7a999e173df054321f4ddaebecdb61756f8a43e91c\",\"Node\":\"enode://3f05da2475bf09ce20b790d76b42450996bc1d3c113a1848be1960171f9851c0@149.129.172.91:44444\"}"],"id":1}' http://127.0.0.110085
```

##### response：

```json
{"jsonrpc":"2.0","id":1,"result":"0x3a3b59f90a21c2fd1b690aa3a2bc06dc2d40eb5bdc26fdd7ecb7e1105af2638e"}
````


### 14. account_CancelCandidateCredit
#### usage：To cancel the candidate
> params：
 1. The address at which the transfer was cancel
 2. address of candidate
 3. amount
 4. gas price
 5. gas limit

#### return：transaction hash

#### example

```shell
curl -H "Content-Type: application/json" -X post --data '{"jsonrpc":"2.0","method":"account_cancelCandidateCredit","params":["0x3ebcbe7cb440dd8c52940a2963472380afbb56c5","0x111","0x110","0x30000",""],"id":1}' http://127.0.0.110085
```

##### response：

```json
{"jsonrpc":"2.0","id":1,"result":"0x3a3b59f90a21c2fd1b690aa3a2bc06dc2d40eb5bdc26fdd7ecb7e1105af2638e"}
````


### 15. account_readContract
#### usage：Read smart contract (no data modified)
> params：
 1. The account address of the transaction
 2. Contract address
 3. Contract api

#### return：The query results

#### example

```shell
curl -H "Content-Type: application/json" -X post --data '{"jsonrpc":"2.0","method":"account_readContract","params":["0xec61c03f719a5c214f60719c3f36bb362a202125","0xecfb51e10aa4c146bf6c12eee090339c99841efc","0x6d4ce63c"],"id":1}' http://127.0.0.110085
```

##### response：

```json
{"jsonrpc":"2.0","id":1,"result":""}
````


### 16. account_estimateGas
#### usage：Estimate how much gas is needed for the transaction
> params：
 1. The address at which the transfer was initiated
 2. amount
 3. commit
 4. Address of recipient

#### return：Evaluate the result, failure returns an error

#### example

```shell
curl -H "Content-Type: application/json" -X post --data '{"jsonrpc":"2.0","method":"account_estimateGas","params":["0xec61c03f719a5c214f60719c3f36bb362a202125","0xecfb51e10aa4c146bf6c12eee090339c99841efc","0x6d4ce63c","0x110","0x30000"],"id":1}' http://127.0.0.110085
```

##### response：

```json
{"jsonrpc":"2.0","id":1,"result":"0x5d74aba54ace5f01a5f0057f37bfddbbe646ea6de7265b368e2e7d17d9cdeb9c"}
````


### 17. account_executeContract
#### usage：Execute smart contract (cause data to be modified)
> params：
 1. The address of the caller
 2. Contract address
 3. Contract code
 4. gas price
 5. gas limit

#### return：transaction hash

#### example

```shell
curl -H "Content-Type: application/json" -X post --data '{"jsonrpc":"2.0","method":"account_executeContract","params":["0xec61c03f719a5c214f60719c3f36bb362a202125","0xecfb51e10aa4c146bf6c12eee090339c99841efc","0x6d4ce63c","0x110","0x30000"],"id":1}' http://127.0.0.110085
```

##### response：

```json
{"jsonrpc":"2.0","id":1,"result":"0x5d74aba54ace5f01a5f0057f37bfddbbe646ea6de7265b368e2e7d17d9cdeb9c"}
````


### 18. account_createCode
#### usage：Deployment of contract
> params：
 1. The account address of the deployment contract
 2. Content of the contract
 3. gas price
 4. gas limit

#### return：transaction hash

#### example

```shell
curl -H "Content-Type: application/json" -X post --data '{"jsonrpc":"2.0","method":"account_createCode","params":["0x3ebcbe7cb440dd8c52940a2963472380afbb56c5","0x608060405234801561001057600080fd5b5061018c806100206000396000f3fe608060405260043610610051576000357c0100000000000000000000000000000000000000000000000000000000900480634f2be91f146100565780636d4ce63c1461006d578063db7208e31461009e575b600080fd5b34801561006257600080fd5b5061006b6100dc565b005b34801561007957600080fd5b5061008261011c565b604051808260070b60070b815260200191505060405180910390f35b3480156100aa57600080fd5b506100da600480360360208110156100c157600080fd5b81019080803560070b9060200190929190505050610132565b005b60016000808282829054906101000a900460070b0192506101000a81548167ffffffffffffffff021916908360070b67ffffffffffffffff160217905550565b60008060009054906101000a900460070b905090565b806000806101000a81548167ffffffffffffffff021916908360070b67ffffffffffffffff1602179055505056fea165627a7a723058204b651e4313ab6bc4eda61084cac1f805699cefbb979ddfd3a2d7f970903307cd0029","0x111","0x110","0x30000"],"id":1}' http://127.0.0.110085
```

##### response：

```json
{"jsonrpc":"2.0","id":1,"result":"0x9a8d8d5d7d00bbe0eb1b9431a13a7219008e352241b751b177bfb29e4e75b0d1"}
````


### 19. account_dumpPrivkey
#### usage：The private key corresponding to the export address
> params：
 1. address

#### return：private key

#### example

```shell
curl http://localhost10085 -X POST --data '{"jsonrpc":"2.0","method":"account_dumpPrivkey","params":["0x3ebcbe7cb440dd8c52940a2963472380afbb56c5"], "id": 3}' -H "Content-Type:application/json"
```

##### response：

```json
{"jsonrpc":"2.0","id":3,"result":"0x270f4b122603999d1c07aec97e972a2ddf7bd8b5bfe3543c10814e6a19f13aaf"}
````


### 20. account_DumpPubkey
#### usage：Export the public key corresponding to the address
> params：
 1. address

#### return：public key

#### example

```shell
curl http://localhost10085 -X POST --data '{"jsonrpc":"2.0","method":"account_dumpPubkey","params":["0x3ebcbe7cb440dd8c52940a2963472380afbb56c5"], "id": 3}' -H "Content-Type:application/json"
```

##### response：

```json
{"jsonrpc":"2.0","id":3,"result":"0x270f4b122603999d1c07aec97e972a2ddf7bd8b5bfe3543c10814e6a19f13aaf"}
````


### 21. account_sign
#### usage：Signature transaction
> params：
 1. account of sig
 2. msg for sig

#### return：private key

#### example

```shell
curl http://localhost10085 -X POST --data '{"jsonrpc":"2.0","method":"account_sign","params":["0x3ebcbe7cb440dd8c52940a2963472380afbb56c5", "0x00001c9b8c8fdb1f53faf02321f76253704123e2b56cce065852bab93e526ae2"], "id": 3}' -H "Content-Type:application/json"
```

##### response：

```json
{"jsonrpc":"2.0","id":3,"result":"0x1f1d16412468dd9b67b568d31839ac608bdfddf2580666db4d364eefbe285fdaed569a3c8fa1decfebbfa0ed18b636059dbbf4c2106c45fc8846909833ef2cb1de"}
````


### 22. account_generateAddresses
#### usage：Generate the addresses of the other chains
> params：
 1. address of drep

#### return：{BTCaddress, ethAddress, neoAddress}

#### example

```shell
curl http://localhost10085 -X POST --data '{"jsonrpc":"2.0","method":"account_generateAddresses","params":["0x3ebcbe7cb440dd8c52940a2963472380afbb56c5"], "id": 3}' -H "Content-Type:application/json"
```

##### response：

```json
{"jsonrpc":"2.0","id":3,"result":""}
````


### 23. account_importKeyStore
#### usage：import keystore
> params：
 1. path
 2. password

#### return：address list

#### example

```shell
curl http://localhost10085 -X POST --data '{"jsonrpc":"2.0","method":"account_importKeyStore","params":["path","123"], "id": 3}' -H "Content-Type:application/json"
```

##### response：

```json
{"jsonrpc":"2.0","id":3,"result":["0x4082c96e38def8f3851831940485066234fe07b8"]}
````


### 24. account_importPrivkey
#### usage：import private key
> params：
 1. privkey(compress hex)
 2. password

#### return：address

#### example

```shell
curl http://localhost10085 -X POST --data '{"jsonrpc":"2.0","method":"account_importPrivkey","params":["0xe5510b32854ca52e7d7d41bb3196fd426d551951e2fd5f6b559a62889d87926c"], "id": 3}' -H "Content-Type:application/json"
```

##### response：

```json
{"jsonrpc":"2.0","id":3,"result":"0x748eb65493a964e568800c3c2885c63a0de9f9ae"}
````


### 25. account_getKeyStores
#### usage：get ketStores path
> params：

#### return：path of keystore

#### example

```shell
curl http://localhost10085 -X POST --data '{"jsonrpc":"2.0","method":"account_getKeyStores","params":[], "id": 3}' -H "Content-Type:application/json"
```

##### response：

```json
{"jsonrpc":"2.0","id":3,"result":"'path of keystores is: C:\\Users\\Kun\\AppData\\Local\\Drep\\keystore'"}
````

consensus api
Query the consensus node function

### 1. consensus_changeWaitTime
#### usage：Modify the waiting time of the leader (ms)
> params：
 1. wait time (ms)

#### return：

#### example

```shell
curl http://localhost10085 -X POST --data '{"jsonrpc":"2.0","method":"consensus_changeWaitTime","params":[100000], "id": 3}' -H "Content-Type:application/json"
```

##### response：

```json
{"jsonrpc":"2.0","id":3,"result":null}
````


### 2. consensus_getMiners()
#### usage：Gets the current mining node
> params：

#### return：mining nodes's pub key

#### example

```shell
curl http://localhost10085 -X POST --data '{"jsonrpc":"2.0","method":"consensus_getMiners","params":[""], "id": 3}' -H "Content-Type:application/json"
```

##### response：

```json
{"jsonrpc":"2.0","id":3,"result":['0x02c682c9f503465a27d1941d1a25547b5ea879a7145056283599a33869982513df', '0x036a09f9012cb3f73c11ceb2aae4242265c2aa35ebec20dbc28a78712802f457db']
}
````

