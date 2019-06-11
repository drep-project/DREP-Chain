# GenDoc 用于从代码中提取出api文档

## example

```txt
/*
 name: getblock
 usage: 用于获取区块信息
 params:
	1. height  usage: 当前区块高度
 return: 区块明细信息
 example: curl http://localhost:15645 -X POST --data '{"jsonrpc":"2.0","method":"chain_getBlock","params":[1], "id": 3}' -H "Content-Type:application/json"
 response:
   response here
*/
```