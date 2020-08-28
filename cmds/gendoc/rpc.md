
#  RPC interface
## Block management
For processing block chain partial upper logic

### 1. blockMgr_sendRawTransaction
- **Usage：**  

&emsp;&emsp;Send signed transactions

- **Params：**  

&emsp;&emsp; 1. A signed transaction

- **Return：transaction hash**

- **Example:**  

**shell:**
```shell
curl http://localhost:10085 -X POST --data '{"jsonrpc":"2.0","method":"blockmgr_sendRawTransaction","params":["0x40a287b6d30b05313131317a4120dd8c23c40910d038fa43b2f8932d3681cbe5ee3079b6e9de0bea6e8e6b2a867a561aa26e1cd6b62aa0422a043186b593b784bf80845c3fd5a7fbfe62e61d8564"], "id": 3}' -H "Content-Type:application/json"
```
**cli:**
```cli
drepClient 127.0.0.1:10085 blockMgr_sendRawTransaction 3
```

- **Response：**

```json
{"jsonrpc":"2.0","id":1,"result":"0xf30e858667fa63bc57ae395c3f57ede9bb3ad4969d12f4bce51d900fb5931538"}
```

---


### 2. blockMgr_gasPrice
- **Usage：**  

&emsp;&emsp;Get the recommended value of gasprice given by the system

- **Params：**  

&emsp;&emsp; 1. Query address

- **Return：Price and error message**

- **Example:**  

**shell:**
```shell
curl http://localhost:10085 -X POST --data '{"jsonrpc":"2.0","method":"blockmgr_gasPrice","params":[], "id": 3}' -H "Content-Type:application/json"
```
**cli:**
```cli
drepClient 127.0.0.1:10085 blockMgr_gasPrice 3
```

- **Response：**

```json

```

---


### 3. blockMgr_GetPoolTransactions
- **Usage：**  

&emsp;&emsp;Get trading information in the trading pool.

- **Params：**  

&emsp;&emsp; 1. Query address

- **Return：All transactions in the pool**

- **Example:**  

**shell:**
```shell
curl http://localhost:10085 -X POST --data '{"jsonrpc":"2.0","method":"blockmgr_getPoolTransactions","params":["0x8a8e541ddd1272d53729164c70197221a3c27486"], "id": 3}' -H "Content-Type:application/json"
```
**cli:**
```cli
drepClient 127.0.0.1:10085 blockMgr_GetPoolTransactions 3
```

- **Response：**

```json

```

---


### 4. blockMgr_GetTransactionCount
- **Usage：**  

&emsp;&emsp;Gets the total number of transactions issued by the address

- **Params：**  

&emsp;&emsp; 1. Query address

- **Return：All transactions in the pool**

- **Example:**  

**shell:**
```shell
curl http://localhost:10085 -X POST --data '{"jsonrpc":"2.0","method":"blockmgr_getTransactionCount","params":["0x8a8e541ddd1272d53729164c70197221a3c27486"], "id": 3}' -H "Content-Type:application/json"
```
**cli:**
```cli
drepClient 127.0.0.1:10085 blockMgr_GetTransactionCount 3
```

- **Response：**

```json

```

---


### 5. blockMgr_GetPoolMiniPendingNonce
- **Usage：**  

&emsp;&emsp;Get the smallest Nonce in the pending queue

- **Params：**  

&emsp;&emsp; 1. Query address

- **Return：The smallest nonce in the pending queue**

- **Example:**  

**shell:**
```shell
curl http://localhost:10085 -X POST --data '{"jsonrpc":"2.0","method":"blockmgr_getPoolMiniPendingNonce","params":["0x8a8e541ddd1272d53729164c70197221a3c27486"], "id": 3}' -H "Content-Type:application/json"
```
**cli:**
```cli
drepClient 127.0.0.1:10085 blockMgr_GetPoolMiniPendingNonce 3
```

- **Response：**

```json

```

---


### 6. blockMgr_GetTxInPool
- **Usage：**  

&emsp;&emsp;Checks whether the transaction is in the trading pool and, if so, returns the transaction

- **Params：**  

&emsp;&emsp; 1. The address at which the transfer was initiated

- **Return：Complete transaction information**

- **Example:**  

**shell:**
```shell
curl -H "Content-Type: application/json" -X post --data '{"jsonrpc":"2.0","method":"blockmgr_getTxInPool","params":["0x3ebcbe7cb440dd8c52940a2963472380afbb56c5"],"id":1}' http://127.0.0.1:10085
```
**cli:**
```cli
drepClient 127.0.0.1:10085 blockMgr_GetTxInPool 3
```

- **Response：**

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
```

---

## Block chain API
Used to obtain block information

### 1. chain_getblock
- **Usage：**  

&emsp;&emsp;Used to obtain block information

- **Params：**  

&emsp;&emsp; 1. height  usage: Current block height

- **Return：Block detail information**

- **Example:**  

**shell:**
```shell
curl http://localhost:10085 -X POST --data '{"jsonrpc":"2.0","method":"chain_getBlock","params":[1], "id": 3}' -H "Content-Type:application/json"
```
**cli:**
```cli
drepClient 127.0.0.1:10085 chain_getblock 3
```

- **Response：**

```json
{
    "jsonrpc":"2.0",
    "id":3,
    "result":{
        "Header":{
            "ChainId":0,
            "Version":1,
            "PreviousHash":"0x1fbae528a8eed0f09201bfd2c7e52fef66f5f35619e9868cd6d02dabac60e4e6",
            "GasLimit":18000000,
            "GasUsed":0,
            "Height":1,
            "Timestamp":1592365562,
            "StateRoot":"UpMnHA5WmmTxU4T4jFQvpt6bFwigN+fg1Jx0fSD91MA=",
            "TxRoot":null,
            "ReceiptRoot":"0x0000000000000000000000000000000000000000000000000000000000000000",
            "Bloom":"0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"
        },
        "Data":{
            "TxCount":0,
            "TxList":null
        },
        "Proof":{
            "Type":0,
            "Evidence":"MEUCIQDIZnsow/WbAmQ7jJ21EcVxzQkKA33LJfw8anhzkNjBzAIgMexycsYJlYEv0rbPvleoAx1iahzUx6FrMNZhh8uq6Lg="
        }
    }
}
```

---


### 2. chain_getMaxHeight
- **Usage：**  

&emsp;&emsp;To get the current highest block

- **Params：**  

&emsp;&emsp; 1. 无

- **Return：Current maximum block height value**

- **Example:**  

**shell:**
```shell
curl http://localhost:10085 -X POST --data '{"jsonrpc":"2.0","method":"chain_getMaxHeight","params":[], "id": 3}' -H "Content-Type:application/json"
```
**cli:**
```cli
drepClient 127.0.0.1:10085 chain_getMaxHeight 3
```

- **Response：**

```json
{"jsonrpc":"2.0","id":3,"result":193005}
```

---


### 3. chain_getBlockGasInfo
- **Usage：**  

&emsp;&emsp;Obtain gas related information

- **Params：**  

&emsp;&emsp; 1. 无

- **Return：Gas minimum value and maximum value required by the system; And the maximum gas value that the current block is set to**

- **Example:**  

**shell:**
```shell
curl http://localhost:10085 -X POST --data '{"jsonrpc":"2.0","method":"chain_getBlockGasInfo","params":[], "id": 3}' -H "Content-Type:application/json"
```
**cli:**
```cli
drepClient 127.0.0.1:10085 chain_getBlockGasInfo 3
```

- **Response：**

```json
{"jsonrpc":"2.0","id":3,"result":193005}
```

---


### 4. chain_getBalance
- **Usage：**  

&emsp;&emsp;Query address balance

- **Params：**  

&emsp;&emsp; 1. Query address

- **Return：The account balance in the address**

- **Example:**  

**shell:**
```shell
curl http://localhost:10085 -X POST --data '{"jsonrpc":"2.0","method":"chain_getBalance","params":["0x8a8e541ddd1272d53729164c70197221a3c27486"], "id": 3}' -H "Content-Type:application/json"
```
**cli:**
```cli
drepClient 127.0.0.1:10085 chain_getBalance 3
```

- **Response：**

```json
{"jsonrpc":"2.0","id":3,"result":9987999999999984000000}
```

---


### 5. chain_getNonce
- **Usage：**  

&emsp;&emsp;Query the nonce whose address is on the chain

- **Params：**  

&emsp;&emsp; 1. Query address

- **Return：nonce**

- **Example:**  

**shell:**
```shell
curl http://localhost:10085 -X POST --data '{"jsonrpc":"2.0","method":"chain_getNonce","params":["0x8a8e541ddd1272d53729164c70197221a3c27486"], "id": 3}' -H "Content-Type:application/json"
```
**cli:**
```cli
drepClient 127.0.0.1:10085 chain_getNonce 3
```

- **Response：**

```json
{"jsonrpc":"2.0","id":3,"result":0}
```

---


### 6. chain_GetReputation
- **Usage：**  

&emsp;&emsp;Query the reputation value of the address

- **Params：**  

&emsp;&emsp; 1. Query address

- **Return：The reputation value corresponding to the address**

- **Example:**  

**shell:**
```shell
curl http://localhost:10085 -X POST --data '{"jsonrpc":"2.0","method":"chain_getReputation","params":["0x8a8e541ddd1272d53729164c70197221a3c27486"], "id": 3}' -H "Content-Type:application/json"
```
**cli:**
```cli
drepClient 127.0.0.1:10085 chain_GetReputation 3
```

- **Response：**

```json
{"jsonrpc":"2.0","id":3,"result":1}
```

---


### 7. chain_getTransactionByBlockHeightAndIndex
- **Usage：**  

&emsp;&emsp;Gets a particular sequence of transactions in a block

- **Params：**  

&emsp;&emsp; 1. block height
 2. Transaction sequence

- **Return：transaction**

- **Example:**  

**shell:**
```shell
curl http://localhost:10085 -X POST --data '{"jsonrpc":"2.0","method":"chain_getTransactionByBlockHeightAndIndex","params":[10000,1], "id": 3}' -H "Content-Type:application/json"
```
**cli:**
```cli
drepClient 127.0.0.1:10085 chain_getTransactionByBlockHeightAndIndex 3
```

- **Response：**

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
```

---


### 8. chain_getAliasByAddress
- **Usage：**  

&emsp;&emsp;Gets the alias corresponding to the address according to the address

- **Params：**  

&emsp;&emsp; 1. address

- **Return：Address the alias**

- **Example:**  

**shell:**
```shell
curl http://localhost:10085 -X POST --data '{"jsonrpc":"2.0","method":"chain_getAliasByAddress","params":["0x8a8e541ddd1272d53729164c70197221a3c27486"], "id": 3}' -H "Content-Type:application/json"
```
**cli:**
```cli
drepClient 127.0.0.1:10085 chain_getAliasByAddress 3
```

- **Response：**

```json
{"jsonrpc":"2.0","id":3,"result":"tom"}
```

---


### 9. chain_getAddressByAlias
- **Usage：**  

&emsp;&emsp;Gets the address corresponding to the alias based on the alias

- **Params：**  

&emsp;&emsp; 1. Alias to be queried

- **Return：The address corresponding to the alias**

- **Example:**  

**shell:**
```shell
curl http://localhost:10085 -X POST --data '{"jsonrpc":"2.0","method":"chain_getAddressByAlias","params":["tom"], "id": 3}' -H "Content-Type:application/json"
```
**cli:**
```cli
drepClient 127.0.0.1:10085 chain_getAddressByAlias 3
```

- **Response：**

```json
{"jsonrpc":"2.0","id":3,"result":"0x8a8e541ddd1272d53729164c70197221a3c27486"}
```

---


### 10. chain_getReceipt
- **Usage：**  

&emsp;&emsp;Get the receipt information based on txhash

- **Params：**  

&emsp;&emsp; 1. txhash

- **Return：receipt**

- **Example:**  

**shell:**
```shell
curl http://localhost:10085 -X POST --data '{"jsonrpc":"2.0","method":"chain_getReceipt","params":["0x7d9dd32ca192e765ff2abd7c5f8931cc3f77f8f47d2d52170c7804c2ca2c5dd9"], "id": 3}' -H "Content-Type:application/json"
```
**cli:**
```cli
drepClient 127.0.0.1:10085 chain_getReceipt 3
```

- **Response：**

```json
{"jsonrpc":"2.0","id":3,"result":""}
```

---


### 11. chain_getLogs
- **Usage：**  

&emsp;&emsp;Get the transaction log information based on txhash

- **Params：**  

&emsp;&emsp; 1. txhash

- **Return：[]log**

- **Example:**  

**shell:**
```shell
curl http://localhost:10085 -X POST --data '{"jsonrpc":"2.0","method":"chain_getLogs","params":["0x7d9dd32ca192e765ff2abd7c5f8931cc3f77f8f47d2d52170c7804c2ca2c5dd9"], "id": 3}' -H "Content-Type:application/json"
```
**cli:**
```cli
drepClient 127.0.0.1:10085 chain_getLogs 3
```

- **Response：**

```json
{"jsonrpc":"2.0","id":3,"result":""}
```

---


### 12. chain_getCancelCreditDetailByTXHash
- **Usage：**  

&emsp;&emsp;Get the back pledge or back vote information according to txhash

- **Params：**  

&emsp;&emsp; 1. txhash

- **Return：{}**

- **Example:**  

**shell:**
```shell
curl http://localhost:10085 -X POST --data '{"jsonrpc":"2.0","method":"chain_getCancelCreditDetailByTXHash","params":["0x7d9dd32ca192e765ff2abd7c5f8931cc3f77f8f47d2d52170c7804c2ca2c5dd9"], "id": 3}' -H "Content-Type:application/json"
```
**cli:**
```cli
drepClient 127.0.0.1:10085 chain_getCancelCreditDetailByTXHash 3
```

- **Response：**

```json
{"jsonrpc":"2.0","id":3,"result":[]}
```

---


### 13. chain_getByteCode
- **Usage：**  

&emsp;&emsp;Get bytecode by address

- **Params：**  

&emsp;&emsp; 1. address

- **Return：bytecode**

- **Example:**  

**shell:**
```shell
curl http://localhost:10085 -X POST --data '{"jsonrpc":"2.0","method":"chain_getByteCode","params":["0x8a8e541ddd1272d53729164c70197221a3c27486"], "id": 3}' -H "Content-Type:application/json"
```
**cli:**
```cli
drepClient 127.0.0.1:10085 chain_getByteCode 3
```

- **Response：**

```json
{"jsonrpc":"2.0","id":3,"result":"0x00"}
```

---


### 14. chain_getCreditDetails
- **Usage：**  

&emsp;&emsp;Get all the details of the stake according to the address

- **Params：**  

&emsp;&emsp; 1. address

- **Return：bytecode**

- **Example:**  

**shell:**
```shell
curl http://localhost:10085 -X POST --data '{"jsonrpc":"2.0","method":"chain_getCreditDetails","params":["0x8a8e541ddd1272d53729164c70197221a3c27486"], "id": 3}' -H "Content-Type:application/json"
```
**cli:**
```cli
drepClient 127.0.0.1:10085 chain_getCreditDetails 3
```

- **Response：**

```json
{"jsonrpc":"2.0","id":3,"result":"[{\"Addr\":\"DREPd05d5f324ada3c418e14cd6b497f2f36d60ba607\",\"HeightValues\":[{\"CreditHeight\":1329,\"CreditValue\":\"0x11135\"}]}]"}
```

---


### 15. chain_GetCancelCreditDetails
- **Usage：**  

&emsp;&emsp;Get the details of all refund requests

- **Params：**  

&emsp;&emsp; 1. address

- **Return：bytecode**

- **Example:**  

**shell:**
```shell
curl http://localhost:10085 -X POST --data '{"jsonrpc":"2.0","method":"chain_getCancelCreditDetails","params":["0x8a8e541ddd1272d53729164c70197221a3c27486"], "id": 3}' -H "Content-Type:application/json"
```
**cli:**
```cli
drepClient 127.0.0.1:10085 chain_GetCancelCreditDetails 3
```

- **Response：**

```json
{"jsonrpc":"2.0","id":3,"result":"{\"DREP300fc5a14e578be28c64627c0e7e321771c58cd4\":\"0x3641100\"}"}
```

---


### 16. chain_GetCandidateAddrs
- **Usage：**  

&emsp;&emsp;Gets the addresses of all candidate nodes and the corresponding trust values

- **Params：**  

&emsp;&emsp; 1. address

- **Return：[]**

- **Example:**  

**shell:**
```shell
curl http://localhost:10085 -X POST --data '{"jsonrpc":"2.0","method":"chain_getCandidateAddrs","params":[""], "id": 3}' -H "Content-Type:application/json"
```
**cli:**
```cli
drepClient 127.0.0.1:10085 chain_GetCandidateAddrs 3
```

- **Response：**

```json
{"jsonrpc":"2.0","id":3,"result":"{\"DREP300fc5a14e578be28c64627c0e7e321771c58cd4\":\"0x3641100\"}"}
```

---


### 17. chain_getChangeCycle
- **Usage：**  

&emsp;&emsp;Gets the transition period of the out - of - block node

- **Params：**  

&emsp;&emsp;
- **Return：Transition period**

- **Example:**  

**shell:**
```shell
curl http://localhost:10085 -X POST --data '{"jsonrpc":"2.0","method":"chain_getChangeCycle","params":"", "id": 3}' -H "Content-Type:application/json"
```
**cli:**
```cli
drepClient 127.0.0.1:10085 chain_getChangeCycle 3
```

- **Response：**

```json
{"jsonrpc":"2.0","id":3,"result":"{100}"}
```

---

## p2p network interface
Set or query network status

### 1. p2p_getPeers
- **Usage：**  

&emsp;&emsp;Get currently connected nodes

- **Params：**  

&emsp;&emsp;
- **Return：P2P peer-to-peer information connecting with local**

- **Example:**  

**shell:**
```shell
curl http://127.0.0.1:10085 -X POST --data '{"jsonrpc":"2.0","method":"p2p_getPeers","params":"", "id": 3}' -H "Content-Type:application/json"
```
**cli:**
```cli
drepClient 127.0.0.1:10085 p2p_getPeers 3
```

- **Response：**

```json
{"jsonrpc":"2.0","id":3,"result":[{},{},{},{}]}
```

---


### 2. p2p_addPeer
- **Usage：**  

&emsp;&emsp;Add peer node

- **Params：**  

&emsp;&emsp;
- **Return：nil**

- **Example:**  

**shell:**
```shell
curl http://127.0.0.1:10085 -X POST --data '{"jsonrpc":"2.0","method":"p2p_addPeer","params":["enode://e1b2f83b7b0f5845cc74ca12bb40152e520842bbd0597b7770cb459bd40f109178811ebddd6d640100cdb9b661a3a43a9811d9fdc63770032a3f2524257fb62d@192.168.74.1:55555"], "id": 3}' -H "Content-Type:application/json"
```
**cli:**
```cli
drepClient 127.0.0.1:10085 p2p_addPeer 3
```

- **Response：**

```json

```

---


### 3. p2p_removePeer
- **Usage：**  

&emsp;&emsp;remove peer node

- **Params：**  

&emsp;&emsp;
- **Return：nil**

- **Example:**  

**shell:**
```shell
curl http://127.0.0.1:10085 -X POST --data '{"jsonrpc":"2.0","method":"p2p_removePeer","params":["enode://e1b2f83b7b0f5845cc74ca12bb40152e520842bbd0597b7770cb459bd40f109178811ebddd6d640100cdb9b661a3a43a9811d9fdc63770032a3f2524257fb62d@192.168.74.1:55555"], "id": 3}' -H "Content-Type:application/json"
```
**cli:**
```cli
drepClient 127.0.0.1:10085 p2p_removePeer 3
```

- **Response：**

```json

```

---


### 4. p2p_localNode
- **Usage：**  

&emsp;&emsp;Need to get local eNode for P2P link

- **Params：**  

&emsp;&emsp;
- **Return：local enode**

- **Example:**  

**shell:**
```shell
curl http://127.0.0.1:10085 -X POST --data '{"jsonrpc":"2.0","method":"p2p_localNode","params":[""], "id": 3}' -H "Content-Type:application/json"
```
**cli:**
```cli
drepClient 127.0.0.1:10085 p2p_localNode 3
```

- **Response：**

```json
{"enode://9064107749f41ffffd9177f27af7bb854d702d930462c4be2d91d1772b3f03f3@192.168.31.63:10086}
```

---

## Logging RPC Api
Set the log level

### 1. log_setLevel
- **Usage：**  

&emsp;&emsp;Set the log level

- **Params：**  

&emsp;&emsp; 1. log level（&#34;debug&#34;,&#34;0&#34;）

- **Return：无**

- **Example:**  

**shell:**
```shell
curl http://localhost:10085 -X POST --data '{"jsonrpc":"2.0","method":"log_setLevel","params":["trace"], "id": 3}' -H "Content-Type:application/json"
```
**cli:**
```cli
drepClient 127.0.0.1:10085 log_setLevel 3
```

- **Response：**

```json
{"jsonrpc":"2.0","id":3,"result":null}
```

---


### 2. log_setVmodule
- **Usage：**  

&emsp;&emsp;Set the level by module

- **Params：**  

&emsp;&emsp; 1. module name (txpool=5)

- **Return：无**

- **Example:**  

**shell:**
```shell
curl http://localhost:10085 -X POST --data '{"jsonrpc":"2.0","method":"log_setVmodule","params":["txpool=5"], "id": 3}' -H "Content-Type:application/json"
```
**cli:**
```cli
drepClient 127.0.0.1:10085 log_setVmodule 3
```

- **Response：**

```json
{"jsonrpc":"2.0","id":3,"result":null}
```

---

## history record interface
Query transaction address and other information (need to open the record module)

### 1. trace_getRawTransaction
- **Usage：**  

&emsp;&emsp;Query transaction bytes according to transaction hash

- **Params：**  

&emsp;&emsp; 1. transaction hash

- **Return：transaction byte code**

- **Example:**  

**shell:**
```shell
curl http://localhost:10085 -X POST --data '{"jsonrpc":"2.0","method":"trace_getRawTransaction","params":["0x00001c9b8c8fdb1f53faf02321f76253704123e2b56cce065852bab93e526ae2"], "id": 3}' -H "Content-Type:application/json"
```
**cli:**
```cli
drepClient 127.0.0.1:10085 trace_getRawTransaction 3
```

- **Response：**

```json
{
	  "jsonrpc": "2.0",
	  "id": 3,
	  "result": "0x02a7ae20007923a30bbfbcb998a6534d56b313e68c8e0c594a0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002011102011003030000bc9889d00b004120eba14c77eab7a154833ff14832d8769cfc0b30db288445d6a83ef2fe337aa09042f8174a593543c4acabe7fadf1ad5fceea9c835682cb9dbea3f1d8fec181fb9"
	}
```

---


### 2. trace_getTransaction
- **Usage：**  

&emsp;&emsp;Query transaction details according to transaction hash

- **Params：**  

&emsp;&emsp; 1. transaction hash

- **Return：Transaction details**

- **Example:**  

**shell:**
```shell
curl http://localhost:10085 -X POST --data '{"jsonrpc":"2.0","method":"trace_getTransaction","params":["0x00001c9b8c8fdb1f53faf02321f76253704123e2b56cce065852bab93e526ae2"], "id": 3}' -H "Content-Type:application/json"
```
**cli:**
```cli
drepClient 127.0.0.1:10085 trace_getTransaction 3
```

- **Response：**

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
```

---


### 3. trace_decodeTrasnaction
- **Usage：**  

&emsp;&emsp;De parsing transaction byte information into transaction details

- **Params：**  

&emsp;&emsp; 1. Transaction byte information

- **Return：transaction details**

- **Example:**  

**shell:**
```shell
curl http://localhost:10085 -X POST --data '{"jsonrpc":"2.0","method":"trace_decodeTrasnaction","params":["0x02a7ae20007923a30bbfbcb998a6534d56b313e68c8e0c594a0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002011102011003030000bc9889d00b004120eba14c77eab7a154833ff14832d8769cfc0b30db288445d6a83ef2fe337aa09042f8174a593543c4acabe7fadf1ad5fceea9c835682cb9dbea3f1d8fec181fb9"], "id": 3}' -H "Content-Type:application/json"
```
**cli:**
```cli
drepClient 127.0.0.1:10085 trace_decodeTrasnaction 3
```

- **Response：**

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
```

---


### 4. trace_getSendTransactionByAddr
- **Usage：**  

&emsp;&emsp;Query the transaction sent from the address according to the address, and pagination is supported

- **Params：**  

&emsp;&emsp; 1. address
 2. Page number (from 1)
 3. Page size

- **Return：Transaction list**

- **Example:**  

**shell:**
```shell
curl http://localhost:10085 -X POST --data '{"jsonrpc":"2.0","method":"trace_getSendTransactionByAddr","params":["DREP7923a30bbfbcb998a6534d56b313e68c8e0c594a",1,10], "id": 3}' -H "Content-Type:application/json"
```
**cli:**
```cli
drepClient 127.0.0.1:10085 trace_getSendTransactionByAddr 3
```

- **Response：**

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
```

---


### 5. trace_getReceiveTransactionByAd
- **Usage：**  

&emsp;&emsp;Query the transaction accepted by the address and support paging

- **Params：**  

&emsp;&emsp; 1. addr
 2. Page number (from 1)
 3. page size

- **Return：transaction list**

- **Example:**  

**shell:**
```shell
curl http://localhost:10085 -X POST --data '{"jsonrpc":"2.0","method":"trace_getReceiveTransactionByAddr","params":["DREP3ebcbe7cb440dd8c52940a2963472380afbb56c5",1,10], "id": 3}' -H "Content-Type:application/json"
```
**cli:**
```cli
drepClient 127.0.0.1:10085 trace_getReceiveTransactionByAd 3
```

- **Response：**

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
```

---


### 6. trace_rebuild
- **Usage：**  

&emsp;&emsp;Reconstructing block records in trace

- **Params：**  

&emsp;&emsp; 1. Start block (included)
 2. Termination block (not included)

- **Return：**

- **Example:**  

**shell:**
```shell
curl http://localhost:10085 -X POST --data '{"jsonrpc":"2.0","method":"trace_rebuild","params":[1,10], "id": 3}' -H "Content-Type:application/json"
```
**cli:**
```cli
drepClient 127.0.0.1:10085 trace_rebuild 3
```

- **Response：**

```json
{"jsonrpc":"2.0","id":3,"result":null}
```

---

## Account RPC interface
Address management and initiate simple transactions

### 1. account_listAddress
- **Usage：**  

&emsp;&emsp;Lists all local addresses

- **Params：**  

&emsp;&emsp;
- **Return：Address of the array**

- **Example:**  

**shell:**
```shell
curl http://localhost:10085 -X POST --data '{"jsonrpc":"2.0","method":"account_listAddress","params":[], "id": 3}' -H "Content-Type:application/json"
```
**cli:**
```cli
drepClient 127.0.0.1:10085 account_listAddress 3
```

- **Response：**

```json
{"jsonrpc":"2.0","id":3,"result":["0x3296d3336895b5baaa0eca3df911741bd0681c3f","0x3ebcbe7cb440dd8c52940a2963472380afbb56c5"]}
```

---


### 2. account_createAccount
- **Usage：**  

&emsp;&emsp;Create a local account

- **Params：**  

&emsp;&emsp; 1. password

- **Return：New account address information**

- **Example:**  

**shell:**
```shell
curl http://localhost:10085 -X POST --data '{"jsonrpc":"2.0","method":"account_createAccount","params":["123456"], "id": 3}' -H "Content-Type:application/json"
```
**cli:**
```cli
drepClient 127.0.0.1:10085 account_createAccount 3
```

- **Response：**

```json
{"jsonrpc":"2.0","id":3,"result":"0x2944c15c466fad03ec1282bab579dec5a0cf0fa3"}
```

---


### 3. account_createWallet
- **Usage：**  

&emsp;&emsp;Create a local wallet

- **Params：**  

&emsp;&emsp; 1. The wallet password

- **Return：Failure returns the reason for the error, and success returns no information**

- **Example:**  

**shell:**
```shell
curl http://localhost:10085 -X POST --data '{"jsonrpc":"2.0","method":"account_createWallet","params":["123"], "id": 3}' -H "Content-Type:application/json"
```
**cli:**
```cli
drepClient 127.0.0.1:10085 account_createWallet 3
```

- **Response：**

```json
{"jsonrpc":"2.0","id":3,"result":null}
```

---


### 4. account_lockAccount
- **Usage：**  

&emsp;&emsp;Lock the account

- **Params：**  

&emsp;&emsp;
- **Return：Failure returns the reason for the error, and success returns no information**

- **Example:**  

**shell:**
```shell
curl http://localhost:10085 -X POST --data '{"jsonrpc":"2.0","method":"account_lockAccount","params":["0x518b3fefa3fb9a72753c6ad10a2b68cc034ec391"], "id": 3}' -H "Content-Type:application/json"
```
**cli:**
```cli
drepClient 127.0.0.1:10085 account_lockAccount 3
```

- **Response：**

```json
{"jsonrpc":"2.0","id":3,"result":null}
```

---


### 5. account_account_unlockAccount
- **Usage：**  

&emsp;&emsp;Unlock the account

- **Params：**  

&emsp;&emsp; 1. The account address
 2. password

- **Return：Failure returns the reason for the error, and success returns no information**

- **Example:**  

**shell:**
```shell
curl http://localhost:10085 -X POST --data '{"jsonrpc":"2.0","method":"account_unlockAccount","params":["0x518b3fefa3fb9a72753c6ad10a2b68cc034ec391", "123456"], "id": 3}' -H "Content-Type:application/json"
```
**cli:**
```cli
drepClient 127.0.0.1:10085 account_account_unlockAccount 3
```

- **Response：**

```json
{"jsonrpc":"2.0","id":3,"result":null}
```

---


### 6. account_openWallet
- **Usage：**  

&emsp;&emsp;Open my wallet

- **Params：**  

&emsp;&emsp; 1. The wallet password

- **Return：error or none**

- **Example:**  

**shell:**
```shell
curl http://localhost:10085 -X POST --data '{"jsonrpc":"2.0","method":"account_openWallet","params":["123"], "id": 3}' -H "Content-Type:application/json"
```
**cli:**
```cli
drepClient 127.0.0.1:10085 account_openWallet 3
```

- **Response：**

```json
{"jsonrpc":"2.0","id":3,"result":null}
```

---


### 7. account_closeWallet
- **Usage：**  

&emsp;&emsp;close wallet

- **Params：**  

&emsp;&emsp;
- **Return：none**

- **Example:**  

**shell:**
```shell
curl http://localhost:10085 -X POST --data '{"jsonrpc":"2.0","method":"account_closeWallet","params":[], "id": 3}' -H "Content-Type:application/json"
```
**cli:**
```cli
drepClient 127.0.0.1:10085 account_closeWallet 3
```

- **Response：**

```json
{"jsonrpc":"2.0","id":3,"result":null}
```

---


### 8. account_transfer
- **Usage：**  

&emsp;&emsp;transfer

- **Params：**  

&emsp;&emsp; 1. The address at which the transfer was initiated
 2. Recipient&#39;s address
 3. Mount
 4. gas price
 5. gas limit
 6. commit

- **Return：transaction hash**

- **Example:**  

**shell:**
```shell
curl -H "Content-Type: application/json" -X post --data '{"jsonrpc":"2.0","method":"account_transfer","params":["0x3ebcbe7cb440dd8c52940a2963472380afbb56c5","0x3ebcbe7cb440dd8c52940a2963472380afbb56c5","0x111","0x110","0x30000",""],"id":1}' http://127.0.0.1:10085
```
**cli:**
```cli
drepClient 127.0.0.1:10085 account_transfer 3
```

- **Response：**

```json
{"jsonrpc":"2.0","id":1,"result":"0x3a3b59f90a21c2fd1b690aa3a2bc06dc2d40eb5bdc26fdd7ecb7e1105af2638e"}
```

---


### 9. account_transferWithNonce
- **Usage：**  

&emsp;&emsp;transfer with nonce

- **Params：**  

&emsp;&emsp; 1. The address at which the transfer was initiated
 2. Recipient&#39;s address
 3. Mount
 4. gas price
 5. gas limit
 6. commit
 7. nonce

- **Return：transaction hash**

- **Example:**  

**shell:**
```shell
curl -H "Content-Type: application/json" -X post --data '{"jsonrpc":"2.0","method":"account_transferWithNonce","params":["0x3ebcbe7cb440dd8c52940a2963472380afbb56c5","0x3ebcbe7cb440dd8c52940a2963472380afbb56c5","0x111","0x110","0x30000","",1],"id":1}' http://127.0.0.1:10085
```
**cli:**
```cli
drepClient 127.0.0.1:10085 account_transferWithNonce 3
```

- **Response：**

```json
{"jsonrpc":"2.0","id":1,"result":"0x3a3b59f90a21c2fd1b690aa3a2bc06dc2d40eb5bdc26fdd7ecb7e1105af2638e"}
```

---


### 10. account_setAlias
- **Usage：**  

&emsp;&emsp;Set an alias

- **Params：**  

&emsp;&emsp; 1. address
 2. alias
 3. gas price
 4. gas lowLimit

- **Return：transaction hash**

- **Example:**  

**shell:**
```shell
curl -H "Content-Type: application/json" -X post --data '{"jsonrpc":"2.0","method":"account_setAlias","params":["0x3ebcbe7cb440dd8c52940a2963472380afbb56c5","AAAAA","0x110","0x30000"],"id":1}' http://127.0.0.1:10085
```
**cli:**
```cli
drepClient 127.0.0.1:10085 account_setAlias 3
```

- **Response：**

```json
{"jsonrpc":"2.0","id":1,"result":"0x5adb248f2943e12fb91c140bd3d0df6237712061e9abae97345b0869c3daa749"}
```

---


### 11. account_VoteCredit
- **Usage：**  

&emsp;&emsp;vote credit to candidate

- **Params：**  

&emsp;&emsp; 1. address of voter
 2. address of candidate
 3. amount
 4. gas price
 5. gas uplimit of transaction

- **Return：transaction hash**

- **Example:**  

**shell:**
```shell
curl -H "Content-Type: application/json" -X post --data '{"jsonrpc":"2.0","method":"account_voteCredit","params":["0x3ebcbe7cb440dd8c52940a2963472380afbb56c5","0x3ebcbe7cb440dd8c52940a2963472380afbb56c5","0x111","0x110","0x30000"],"id":1}' http://127.0.0.1:10085
```
**cli:**
```cli
drepClient 127.0.0.1:10085 account_VoteCredit 3
```

- **Response：**

```json
{"jsonrpc":"2.0","id":1,"result":"0x3a3b59f90a21c2fd1b690aa3a2bc06dc2d40eb5bdc26fdd7ecb7e1105af2638e"}
```

---


### 12. account_CancelVoteCredit
- **Usage：**  

&emsp;&emsp;

- **Params：**  

&emsp;&emsp; 1. address of voter
 2. address of candidate
 3. amount
 4. gas price
 5. gas limit
 6. 备注

- **Return：transaction hash**

- **Example:**  

**shell:**
```shell
curl -H "Content-Type: application/json" -X post --data '{"jsonrpc":"2.0","method":"account_cancelVoteCredit","params":["0x3ebcbe7cb440dd8c52940a2963472380afbb56c5","0x3ebcbe7cb440dd8c52940a2963472380afbb56c5","0x111","0x110","0x30000"],"id":1}' http://127.0.0.1:10085
```
**cli:**
```cli
drepClient 127.0.0.1:10085 account_CancelVoteCredit 3
```

- **Response：**

```json
{"jsonrpc":"2.0","id":1,"result":"0x3a3b59f90a21c2fd1b690aa3a2bc06dc2d40eb5bdc26fdd7ecb7e1105af2638e"}
```

---


### 13. account_CandidateCredit
- **Usage：**  

&emsp;&emsp;Candidate node pledge

- **Params：**  

&emsp;&emsp; 1. The address of the pledger
 2. The pledge amount
 3. gas price
 4. gas limit
 5. The pubkey corresponding to the address of the pledger, and the P2p information of the pledger

- **Return：transaction hash**

- **Example:**  

**shell:**
```shell
curl -H "Content-Type: application/json" -X post --data '{"jsonrpc":"2.0","method":"account_candidateCredit","params":["0x3ebcbe7cb440dd8c52940a2963472380afbb56c5","0x111","0x110","0x30000","{\"Pubkey\":\"0x020e233ebaed5ade5e48d7ee7a999e173df054321f4ddaebecdb61756f8a43e91c\",\"Node\":\"enode://3f05da2475bf09ce20b790d76b42450996bc1d3c113a1848be1960171f9851c0@149.129.172.91:44444\"}"],"id":1}' http://127.0.0.1:10085
```
**cli:**
```cli
drepClient 127.0.0.1:10085 account_CandidateCredit 3
```

- **Response：**

```json
{"jsonrpc":"2.0","id":1,"result":"0x3a3b59f90a21c2fd1b690aa3a2bc06dc2d40eb5bdc26fdd7ecb7e1105af2638e"}
```

---


### 14. account_CancelCandidateCredit
- **Usage：**  

&emsp;&emsp;To cancel the candidate

- **Params：**  

&emsp;&emsp; 1. The address at which the transfer was cancel
 2. address of candidate
 3. amount
 4. gas price
 5. gas limit

- **Return：transaction hash**

- **Example:**  

**shell:**
```shell
curl -H "Content-Type: application/json" -X post --data '{"jsonrpc":"2.0","method":"account_cancelCandidateCredit","params":["0x3ebcbe7cb440dd8c52940a2963472380afbb56c5","0x111","0x110","0x30000",""],"id":1}' http://127.0.0.1:10085
```
**cli:**
```cli
drepClient 127.0.0.1:10085 account_CancelCandidateCredit 3
```

- **Response：**

```json
{"jsonrpc":"2.0","id":1,"result":"0x3a3b59f90a21c2fd1b690aa3a2bc06dc2d40eb5bdc26fdd7ecb7e1105af2638e"}
```

---


### 15. account_readContract
- **Usage：**  

&emsp;&emsp;Read smart contract (no data modified)

- **Params：**  

&emsp;&emsp; 1. The account address of the transaction
 2. Contract address
 3. Contract api

- **Return：The query results**

- **Example:**  

**shell:**
```shell
curl -H "Content-Type: application/json" -X post --data '{"jsonrpc":"2.0","method":"account_readContract","params":["0xec61c03f719a5c214f60719c3f36bb362a202125","0xecfb51e10aa4c146bf6c12eee090339c99841efc","0x6d4ce63c"],"id":1}' http://127.0.0.1:10085
```
**cli:**
```cli
drepClient 127.0.0.1:10085 account_readContract 3
```

- **Response：**

```json
{"jsonrpc":"2.0","id":1,"result":""}
```

---


### 16. account_estimateGas
- **Usage：**  

&emsp;&emsp;Estimate how much gas is needed for the transaction

- **Params：**  

&emsp;&emsp; 1. The address at which the transfer was initiated
 2. amount
 3. commit
 4. Address of recipient

- **Return：Evaluate the result, failure returns an error**

- **Example:**  

**shell:**
```shell
curl -H "Content-Type: application/json" -X post --data '{"jsonrpc":"2.0","method":"account_estimateGas","params":["0xec61c03f719a5c214f60719c3f36bb362a202125","0xecfb51e10aa4c146bf6c12eee090339c99841efc","0x6d4ce63c","0x110","0x30000"],"id":1}' http://127.0.0.1:10085
```
**cli:**
```cli
drepClient 127.0.0.1:10085 account_estimateGas 3
```

- **Response：**

```json
{"jsonrpc":"2.0","id":1,"result":"0x5d74aba54ace5f01a5f0057f37bfddbbe646ea6de7265b368e2e7d17d9cdeb9c"}
```

---


### 17. account_executeContract
- **Usage：**  

&emsp;&emsp;Execute smart contract (cause data to be modified)

- **Params：**  

&emsp;&emsp; 1. The address of the caller
 2. Contract address
 3. Contract code
 4. gas price
 5. gas limit

- **Return：transaction hash**

- **Example:**  

**shell:**
```shell
curl -H "Content-Type: application/json" -X post --data '{"jsonrpc":"2.0","method":"account_executeContract","params":["0xec61c03f719a5c214f60719c3f36bb362a202125","0xecfb51e10aa4c146bf6c12eee090339c99841efc","0x6d4ce63c","0x110","0x30000"],"id":1}' http://127.0.0.1:10085
```
**cli:**
```cli
drepClient 127.0.0.1:10085 account_executeContract 3
```

- **Response：**

```json
{"jsonrpc":"2.0","id":1,"result":"0x5d74aba54ace5f01a5f0057f37bfddbbe646ea6de7265b368e2e7d17d9cdeb9c"}
```

---


### 18. account_createCode
- **Usage：**  

&emsp;&emsp;Deployment of contract

- **Params：**  

&emsp;&emsp; 1. The account address of the deployment contract
 2. Content of the contract
 3. gas price
 4. gas limit

- **Return：transaction hash**

- **Example:**  

**shell:**
```shell
curl -H "Content-Type: application/json" -X post --data '{"jsonrpc":"2.0","method":"account_createCode","params":["0x3ebcbe7cb440dd8c52940a2963472380afbb56c5","0x608060405234801561001057600080fd5b5061018c806100206000396000f3fe608060405260043610610051576000357c0100000000000000000000000000000000000000000000000000000000900480634f2be91f146100565780636d4ce63c1461006d578063db7208e31461009e575b600080fd5b34801561006257600080fd5b5061006b6100dc565b005b34801561007957600080fd5b5061008261011c565b604051808260070b60070b815260200191505060405180910390f35b3480156100aa57600080fd5b506100da600480360360208110156100c157600080fd5b81019080803560070b9060200190929190505050610132565b005b60016000808282829054906101000a900460070b0192506101000a81548167ffffffffffffffff021916908360070b67ffffffffffffffff160217905550565b60008060009054906101000a900460070b905090565b806000806101000a81548167ffffffffffffffff021916908360070b67ffffffffffffffff1602179055505056fea165627a7a723058204b651e4313ab6bc4eda61084cac1f805699cefbb979ddfd3a2d7f970903307cd0029","0x111","0x110","0x30000"],"id":1}' http://127.0.0.1:10085
```
**cli:**
```cli
drepClient 127.0.0.1:10085 account_createCode 3
```

- **Response：**

```json
{"jsonrpc":"2.0","id":1,"result":"0x9a8d8d5d7d00bbe0eb1b9431a13a7219008e352241b751b177bfb29e4e75b0d1"}
```

---


### 19. account_dumpPrivkey
- **Usage：**  

&emsp;&emsp;The private key corresponding to the export address

- **Params：**  

&emsp;&emsp; 1. address

- **Return：private key**

- **Example:**  

**shell:**
```shell
curl http://localhost:10085 -X POST --data '{"jsonrpc":"2.0","method":"account_dumpPrivkey","params":["0x3ebcbe7cb440dd8c52940a2963472380afbb56c5"], "id": 3}' -H "Content-Type:application/json"
```
**cli:**
```cli
drepClient 127.0.0.1:10085 account_dumpPrivkey 3
```

- **Response：**

```json
{"jsonrpc":"2.0","id":3,"result":"0x270f4b122603999d1c07aec97e972a2ddf7bd8b5bfe3543c10814e6a19f13aaf"}
```

---


### 20. account_DumpPubkey
- **Usage：**  

&emsp;&emsp;Export the public key corresponding to the address

- **Params：**  

&emsp;&emsp; 1. address

- **Return：public key**

- **Example:**  

**shell:**
```shell
curl http://localhost:10085 -X POST --data '{"jsonrpc":"2.0","method":"account_dumpPubkey","params":["0x3ebcbe7cb440dd8c52940a2963472380afbb56c5"], "id": 3}' -H "Content-Type:application/json"
```
**cli:**
```cli
drepClient 127.0.0.1:10085 account_DumpPubkey 3
```

- **Response：**

```json
{"jsonrpc":"2.0","id":3,"result":"0x270f4b122603999d1c07aec97e972a2ddf7bd8b5bfe3543c10814e6a19f13aaf"}
```

---


### 21. account_sign
- **Usage：**  

&emsp;&emsp;Signature transaction

- **Params：**  

&emsp;&emsp; 1. account of sig
 2. msg for sig

- **Return：private key**

- **Example:**  

**shell:**
```shell
curl http://localhost:10085 -X POST --data '{"jsonrpc":"2.0","method":"account_sign","params":["0x3ebcbe7cb440dd8c52940a2963472380afbb56c5", "0x00001c9b8c8fdb1f53faf02321f76253704123e2b56cce065852bab93e526ae2"], "id": 3}' -H "Content-Type:application/json"
```
**cli:**
```cli
drepClient 127.0.0.1:10085 account_sign 3
```

- **Response：**

```json
{"jsonrpc":"2.0","id":3,"result":"0x1f1d16412468dd9b67b568d31839ac608bdfddf2580666db4d364eefbe285fdaed569a3c8fa1decfebbfa0ed18b636059dbbf4c2106c45fc8846909833ef2cb1de"}
```

---


### 22. account_generateAddresses
- **Usage：**  

&emsp;&emsp;Generate the addresses of the other chains

- **Params：**  

&emsp;&emsp; 1. address of drep

- **Return：{BTCaddress, ethAddress, neoAddress}**

- **Example:**  

**shell:**
```shell
curl http://localhost:10085 -X POST --data '{"jsonrpc":"2.0","method":"account_generateAddresses","params":["0x3ebcbe7cb440dd8c52940a2963472380afbb56c5"], "id": 3}' -H "Content-Type:application/json"
```
**cli:**
```cli
drepClient 127.0.0.1:10085 account_generateAddresses 3
```

- **Response：**

```json
{"jsonrpc":"2.0","id":3,"result":""}
```

---


### 23. account_importKeyStore
- **Usage：**  

&emsp;&emsp;import keystore

- **Params：**  

&emsp;&emsp; 1. path
 2. password

- **Return：address list**

- **Example:**  

**shell:**
```shell
curl http://localhost:10085 -X POST --data '{"jsonrpc":"2.0","method":"account_importKeyStore","params":["path","123"], "id": 3}' -H "Content-Type:application/json"
```
**cli:**
```cli
drepClient 127.0.0.1:10085 account_importKeyStore 3
```

- **Response：**

```json
{"jsonrpc":"2.0","id":3,"result":["0x4082c96e38def8f3851831940485066234fe07b8"]}
```

---


### 24. account_importPrivkey
- **Usage：**  

&emsp;&emsp;import private key

- **Params：**  

&emsp;&emsp; 1. privkey(compress hex)
 2. password

- **Return：address**

- **Example:**  

**shell:**
```shell
curl http://localhost:10085 -X POST --data '{"jsonrpc":"2.0","method":"account_importPrivkey","params":["0xe5510b32854ca52e7d7d41bb3196fd426d551951e2fd5f6b559a62889d87926c"], "id": 3}' -H "Content-Type:application/json"
```
**cli:**
```cli
drepClient 127.0.0.1:10085 account_importPrivkey 3
```

- **Response：**

```json
{"jsonrpc":"2.0","id":3,"result":"0x748eb65493a964e568800c3c2885c63a0de9f9ae"}
```

---


### 25. account_getKeyStores
- **Usage：**  

&emsp;&emsp;get ketStores path

- **Params：**  

&emsp;&emsp;
- **Return：path of keystore**

- **Example:**  

**shell:**
```shell
curl http://localhost:10085 -X POST --data '{"jsonrpc":"2.0","method":"account_getKeyStores","params":[], "id": 3}' -H "Content-Type:application/json"
```
**cli:**
```cli
drepClient 127.0.0.1:10085 account_getKeyStores 3
```

- **Response：**

```json
{"jsonrpc":"2.0","id":3,"result":"'path of keystores is: C:\\Users\\Kun\\AppData\\Local\\Drep\\keystore'"}
```

---

## consensus api
Query the consensus node function

### 1. consensus_changeWaitTime
- **Usage：**  

&emsp;&emsp;Modify the waiting time of the leader (ms)

- **Params：**  

&emsp;&emsp; 1. wait time (ms)

- **Return：**

- **Example:**  

**shell:**
```shell
curl http://localhost:10085 -X POST --data '{"jsonrpc":"2.0","method":"consensus_changeWaitTime","params":[100000], "id": 3}' -H "Content-Type:application/json"
```
**cli:**
```cli
drepClient 127.0.0.1:10085 consensus_changeWaitTime 3
```

- **Response：**

```json
{"jsonrpc":"2.0","id":3,"result":null}
```

---


### 2. consensus_getMiners()
- **Usage：**  

&emsp;&emsp;Gets the current mining node

- **Params：**  

&emsp;&emsp;
- **Return：mining nodes's pub key**

- **Example:**  

**shell:**
```shell
curl http://localhost:10085 -X POST --data '{"jsonrpc":"2.0","method":"consensus_getMiners","params":[""], "id": 3}' -H "Content-Type:application/json"
```
**cli:**
```cli
drepClient 127.0.0.1:10085 consensus_getMiners() 3
```

- **Response：**

```json
{"jsonrpc":"2.0","id":3,"result":['0x02c682c9f503465a27d1941d1a25547b5ea879a7145056283599a33869982513df', '0x036a09f9012cb3f73c11ceb2aae4242265c2aa35ebec20dbc28a78712802f457db']
}
```

---

