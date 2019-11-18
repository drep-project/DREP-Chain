package filter

import (
	"context"
	"github.com/drep-project/DREP-Chain/types"
)

/*
name: 过滤器rpc接口
usage: 侦听链上特定的交易，区块，日志信息
prefix: filter
*/

type FilterApi struct {
	filterService *FilterService
}

/*
 name: newPendingTransactionFilter
 usage: Creates a filter in the node, to notify when new pending transactions arrive. To check if the state has changed, call filter_getFilterChanges.
 params:
	None
 return:
	QUANTITY - A filter id.
 example: curl http://localhost:15645 -X POST --data '{"jsonrpc":"2.0","method":"filter_newPendingTransactionFilter","params":[], "id": 3}' -H "Content-Type:application/json"
 response:
{
  "jsonrpc": "2.0",
  "id": 3,
  "result": "0x1"
  }
}
*/
func (filter *FilterApi) NewPendingTransactionFilter() ID {
	return filter.filterService.NewPendingTransactionFilter()
}

/*
 name: newBlockFilter
 usage: Creates a filter in the node, to notify when a new block arrives. To check if the state has changed, call filter_getFilterChanges.
 params:
	None
 return:
	QUANTITY - A filter id.
 example: curl http://localhost:15645 -X POST --data '{"jsonrpc":"2.0","method":"filter_newBlockFilter","params":[], "id": 3}' -H "Content-Type:application/json"
 response:
{
  "jsonrpc": "2.0",
  "id": 3,
  "result": "0x1"
  }
}
*/
func (filter *FilterApi) NewBlockFilter() ID {
	return filter.filterService.NewBlockFilter()
}

/*
 name: newFilter
 usage: Creates a filter object, based on filter options, to notify when the state changes (logs). To check if the state has changed, call filter_getFilterChanges.
 params:
	1. Object - The filter options:
		fromBlock: QUANTITY|TAG - (optional, default: "latest") Integer block number, or "latest" for the last mined block or "pending", "earliest" for not yet mined transactions.
		toBlock: QUANTITY|TAG - (optional, default: "latest") Integer block number, or "latest" for the last mined block or "pending", "earliest" for not yet mined transactions.
		address: DATA|Array, 20 Bytes - (optional) Contract address or a list of addresses from which logs should originate.
		topics: Array of DATA, - (optional) Array of 32 Bytes DATA topics. Topics are order-dependent. Each topic can also be an array of DATA with "or" options.
 return:
	QUANTITY - A filter id.
 example: curl http://localhost:15645 -X POST --data '{"jsonrpc":"2.0","method":"filter_newFilter","params":[{"topics":["0x0000000000000000000000000000000000000000000000000000000012341234"]}], "id": 3}' -H "Content-Type:application/json"
 response:
{
  "jsonrpc": "2.0",
  "id": 3,
  "result": "0x1"
  }
}
*/
func (filter *FilterApi) NewFilter(crit FilterQuery) (ID, error) {
	return filter.filterService.NewFilter(crit)
}

/*
 name: uninstallFilter
 usage: Uninstalls a filter with given id. Should always be called when watch is no longer needed. Additonally Filters timeout when they aren't requested with filter_getFilterChanges for a period of time.
 params:
	1. QUANTITY - The filter id.
 return:
	Boolean - true if the filter was successfully uninstalled, otherwise false.
 example: curl http://localhost:15645 -X POST --data '{"jsonrpc":"2.0","method":"filter_uninstallFilter","params":["0xb"], "id": 3}' -H "Content-Type:application/json"
 response:
{
  "jsonrpc": "2.0",
  "id": 3,
  "result": "0x1"
  }
}
*/
func (filter *FilterApi) UninstallFilter(id ID) bool {
	return filter.filterService.UninstallFilter(id)
}

/*
 name: getLogs
 usage: Returns an array of all logs matching a given filter object.
 params:
	1. Object - The filter options:
		fromBlock: QUANTITY|TAG - (optional, default: "latest") Integer block number, or "latest" for the last mined block or "pending", "earliest" for not yet mined transactions.
		toBlock: QUANTITY|TAG - (optional, default: "latest") Integer block number, or "latest" for the last mined block or "pending", "earliest" for not yet mined transactions.
		address: DATA|Array, 20 Bytes - (optional) Contract address or a list of addresses from which logs should originate.
		topics: Array of DATA, - (optional) Array of 32 Bytes DATA topics. Topics are order-dependent. Each topic can also be an array of DATA with "or" options.
		blockhash: DATA, 32 Bytes - (optional) , blockHash is a new filter option which restricts the logs returned to the single block with the 32-byte hash blockHash. Using blockHash is equivalent to fromBlock = toBlock = the block number with hash blockHash. If blockHash is present in the filter criteria, then neither fromBlock nor toBlock are allowed.
 return:
	Array - Array of log objects, or an empty array if nothing has changed since last poll.
 example: curl http://localhost:15645 -X POST --data '{"jsonrpc":"2.0","method":"filter_getLogs","params":[{"topics":["0x000000000000000000000000a94f5374fce5edbc8e2a8697c15331677e6ebf0b"]}], "id": 3}' -H "Content-Type:application/json"
 response:
{
  "jsonrpc": "2.0",
  "id": 3,
  "result": [{
    "logIndex": "0x1", // 1
    "blockNumber":"0x1b4", // 436
    "blockHash": "0x8216c5785ac562ff41e2dcfdf5785ac562ff41e2dcfdf829c5a142f1fccd7d",
    "transactionHash":  "0xdf829c5a142f1fccd7d8216c5785ac562ff41e2dcfdf5785ac562ff41e2dcf",
    "transactionIndex": "0x0", // 0
    "address": "0x16c5785ac562ff41e2dcfdf829c5a142f1fccd7d",
    "data":"0x0000000000000000000000000000000000000000000000000000000000000000",
    "topics": ["0x59ebeb90bc63057b6515673c3ecf9438e5058bca0f92585014eced636878c9a5"]
    },{
      ...
    }]
  }
}
*/
func (filter *FilterApi) GetLogs(crit FilterQuery) ([]*types.Log, error) {
	return filter.filterService.GetLogs(context.Background(), crit)
}

/*
 name: getFilterLogs
 usage: Returns an array of all logs matching filter with given id.
 params:
	1. QUANTITY - The filter id.
 return:
	Array - Array of log objects, or an empty array if nothing has changed since last poll.
 example: curl http://localhost:15645 -X POST --data '{"jsonrpc":"2.0","method":"filter_getFilterLogs","params":["0x16"], "id": 3}' -H "Content-Type:application/json"
 response:
{
  "jsonrpc": "2.0",
  "id": 3,
  "result": [{
    "logIndex": "0x1", // 1
    "blockNumber":"0x1b4", // 436
    "blockHash": "0x8216c5785ac562ff41e2dcfdf5785ac562ff41e2dcfdf829c5a142f1fccd7d",
    "transactionHash":  "0xdf829c5a142f1fccd7d8216c5785ac562ff41e2dcfdf5785ac562ff41e2dcf",
    "transactionIndex": "0x0", // 0
    "address": "0x16c5785ac562ff41e2dcfdf829c5a142f1fccd7d",
    "data":"0x0000000000000000000000000000000000000000000000000000000000000000",
    "topics": ["0x59ebeb90bc63057b6515673c3ecf9438e5058bca0f92585014eced636878c9a5"]
    },{
      ...
    }]
  }
}
*/
func (filter *FilterApi) GetFilterLogs(id ID) ([]*types.Log, error) {
	return filter.filterService.GetFilterLogs(context.Background(), id)
}

/*
 name: getFilterChanges
 usage: Polling method for a filter, which returns an array of logs which occurred since last poll.
 params:
	1. QUANTITY - the filter id.
 return:
	Array - Array of log objects, or an empty array if nothing has changed since last poll.
 example: curl http://localhost:15645 -X POST --data '{"jsonrpc":"2.0","method":"filter_getFilterChanges","params":["0x16"], "id": 3}' -H "Content-Type:application/json"
 response:
{
  "jsonrpc": "2.0",
  "id": 3,
  "result": [{
    "logIndex": "0x1", // 1
    "blockNumber":"0x1b4", // 436
    "blockHash": "0x8216c5785ac562ff41e2dcfdf5785ac562ff41e2dcfdf829c5a142f1fccd7d",
    "transactionHash":  "0xdf829c5a142f1fccd7d8216c5785ac562ff41e2dcfdf5785ac562ff41e2dcf",
    "transactionIndex": "0x0", // 0
    "address": "0x16c5785ac562ff41e2dcfdf829c5a142f1fccd7d",
    "data":"0x0000000000000000000000000000000000000000000000000000000000000000",
    "topics": ["0x59ebeb90bc63057b6515673c3ecf9438e5058bca0f92585014eced636878c9a5"]
    },{
      ...
    }]
  }
}
*/
func (filter *FilterApi) GetFilterChanges(id ID) (interface{}, error) {
	return filter.filterService.GetFilterChanges(id)
}
