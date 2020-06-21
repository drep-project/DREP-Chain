package bft

import (
	"github.com/drep-project/DREP-Chain/crypto/secp256k1"
	"time"
)

/*
name: consensus api
usage: Query the consensus node function
prefix:consensus
*/
type ConsensusApi struct {
	consensusService *BftConsensusService
}

/*
 name: changeWaitTime
 usage: Modify the waiting time of the leader (ms)
 params:
	1.wait time (ms)
 return:
 example:
	curl http://localhost10085 -X POST --data '{"jsonrpc":"2.0","method":"consensus_changeWaitTime","params":[100000], "id": 3}' -H "Content-Type:application/json"

response:
	 {"jsonrpc":"2.0","id":3,"result":null}
*/
func (consensusApi *ConsensusApi) ChangeWaitTime(waitTime int) {
	consensusApi.consensusService.BftConsensus.ChangeTime(time.Duration(waitTime))
}

/*
 name: getMiners()
 usage: Gets the current mining node
 params:
 return: mining nodes's pub key
 example:
	curl http://localhost10085 -X POST --data '{"jsonrpc":"2.0","method":"consensus_getMiners","params":[""], "id": 3}' -H "Content-Type:application/json"

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
