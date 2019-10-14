package bft

/*
name: 共识rpc接口
usage: 查询共识节点功能
prefix:consensus
*/
type ConsensusApi struct {
	consensusService *BftConsensusService
}

/*
	 name: changeWaitTime
	 usage: 修改leader等待时间 (ms)
	 params:
		1.等待时间(ms)
	 return: 私钥
	 example:
		curl http://localhost:15645 -X POST --data '{"jsonrpc":"2.0","method":"consensus_changeWaitTime","params":[100000], "id": 3}' -H "Content-Type:application/json"

	response:
		 {"jsonrpc":"2.0","id":3,"result":null}
*/
func (consensusApi *ConsensusApi) ChangeWaitTime(waitTime int) {
}
