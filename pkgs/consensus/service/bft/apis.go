package bft

import (
	"github.com/drep-project/DREP-Chain/crypto/secp256k1"
	"time"
)

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
	consensusApi.consensusService.BftConsensus.ChangeTime(time.Duration(waitTime))
}

/*
	 name: getMiners()
	 usage: 获取当前出块节点
	 params:
		无
	 return: 出块节点信息
	 example:
		curl http://localhost:15645 -X POST --data '{"jsonrpc":"2.0","method":"consensus_getMiners","params":[""], "id": 3}' -H "Content-Type:application/json"

	response:
		 {"jsonrpc":"2.0","id":3,"result":['0x02c682c9f503465a27d1941d1a25547b5ea879a7145056283599a33869982513df', '0x036a09f9012cb3f73c11ceb2aae4242265c2aa35ebec20dbc28a78712802f457db']
}
*/
func (consensusApi *ConsensusApi) GetMiners() []*secp256k1.PublicKey {
	miners := consensusApi.consensusService.BftConsensus.producer
	pk := make([]*secp256k1.PublicKey, 0)
	for _, p := range miners {
		pk = append(pk, p.Pubkey)
	}

	return pk
}
