
#  `JSON-RPC`接口说明文档

链接口
用于获取区块信息

## 1  . getblock
### 作用：用于获取区块信息
> 参数：
 1. height  usage: 当前区块高度

### 返回值：
区块明细信息

### 示例代码
#### 请求：

```shell
curl http://localhost:15645 -X POST --data '{"jsonrpc":"2.0","method":"chain_getBlock","params":[1], "id": 3}' -H "Content-Type:application/json"
```

#### 响应：

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
    "LeaderPubKey": "0x0374bf9c8ea268b5548686685dda4a74fc95903ca7c440e5b187a718b595c1f374",
    "MinorPubKeys": [
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


## 2  . getMaxHeight
### 作用：用于获取当前最高区块
> 参数：
 1. 无

### 返回值：
当前最高区块高度数值

### 示例代码
#### 请求：

```shell
curl http://localhost:15645 -X POST --data '{"jsonrpc":"2.0","method":"chain_getMaxHeight","params":[], "id": 3}' -H "Content-Type:application/json"
```

#### 响应：

```json
{"jsonrpc":"2.0","id":3,"result":193005}
````


## 3  . getBalance
### 作用：查询地址余额
> 参数：
 1. 待查询地址

### 返回值：
地址中的账号余额

### 示例代码
#### 请求：

```shell
curl http://localhost:15645 -X POST --data '{"jsonrpc":"2.0","method":"chain_getBalance","params":["0x8a8e541ddd1272d53729164c70197221a3c27486"], "id": 3}' -H "Content-Type:application/json"
```

#### 响应：

```json
{"jsonrpc":"2.0","id":3,"result":9987999999999984000000}
````


## 4  . getNonce
### 作用：查询地址在链上的nonce
> 参数：
 1. 待查询地址

### 返回值：
链上nonce

### 示例代码
#### 请求：

```shell
curl http://localhost:15645 -X POST --data '{"jsonrpc":"2.0","method":"chain_getNonce","params":["0x8a8e541ddd1272d53729164c70197221a3c27486"], "id": 3}' -H "Content-Type:application/json"
```

#### 响应：

```json
{"jsonrpc":"2.0","id":3,"result":0}
````


## 5  . getNonce
### 作用：查询地址的名誉值
> 参数：
 1. 待查询地址

### 返回值：
地址对应的名誉值

### 示例代码
#### 请求：

```shell
curl http://localhost:15645 -X POST --data '{"jsonrpc":"2.0","method":"chain_getReputation","params":["0x8a8e541ddd1272d53729164c70197221a3c27486"], "id": 3}' -H "Content-Type:application/json"
```

#### 响应：

```json
{"jsonrpc":"2.0","id":3,"result":1}
````


## 6  . getTransactionByBlockHeightAndIndex
### 作用：获取区块中特定序列的交易
> 参数：
 1. 区块高度
 2. 交易序列

### 返回值：
交易信息

### 示例代码
#### 请求：

```shell
curl http://localhost:15645 -X POST --data '{"jsonrpc":"2.0","method":"chain_getTransactionByBlockHeightAndIndex","params":[10000,1], "id": 3}' -H "Content-Type:application/json"
```

#### 响应：

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


## 7  . getAliasByAddress
### 作用：根据地址获取地址对应的别名
> 参数：
 1. 待查询地址

### 返回值：
地址别名

### 示例代码
#### 请求：

```shell
curl http://localhost:15645 -X POST --data '{"jsonrpc":"2.0","method":"chain_getAliasByAddress","params":["0x8a8e541ddd1272d53729164c70197221a3c27486"], "id": 3}' -H "Content-Type:application/json"
```

#### 响应：

```json
{"jsonrpc":"2.0","id":3,"result":"tom"}
````


## 8  . etAddressByAlias
### 作用：根据别名获取别名对应的地址
> 参数：
 1. 待查询地别名

### 返回值：
别名对应的地址

### 示例代码
#### 请求：

```shell
curl http://localhost:15645 -X POST --data '{"jsonrpc":"2.0","method":"chain_getAliasByAddress","params":["tom"], "id": 3}' -H "Content-Type:application/json"
```

#### 响应：

```json
{"jsonrpc":"2.0","id":3,"result":"0x8a8e541ddd1272d53729164c70197221a3c27486"}
````

