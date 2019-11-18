package service

import (
	"github.com/drep-project/DREP-Chain/pkgs/consensus/service/bft"
	"time"
)

/*
name: 共识rpc接口
usage: 查询共识节点功能
prefix:consensus
*/
type ConsensusApi struct {
	consensusService *ConsensusService
}

/*
 name: minning
 usage: 查询是否在出块状态 (需开启共识模块)
 params:
 return: true/false
 example:  curl http://localhost:15645 -X POST --data '{"jsonrpc":"2.0","method":"consensus_minning, "params":[], "id": 3}' -H "Content-Type:application/json"
 response:
  {"jsonrpc":"2.0","id":3,"result":false}
*/
func (consensusApi *ConsensusApi) Minning() bool {
	switch consensusApi.consensusService.Config.ConsensusMode {
	case "solo":
		return true
	case "bft":
		return true
	default:
		return false
	}
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
	if consensusApi.consensusService.Config.ConsensusMode == "bft" {

	} else {
		consensusApi.consensusService.ConsensusEngine.(*bft.BftConsensus).ChangeTime(time.Duration(int64(time.Millisecond) * int64(waitTime)))
	}
}
