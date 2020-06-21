package service

import "github.com/drep-project/DREP-Chain/network/p2p/enode"

/*
name: p2p network interface
usage: Set or query network status
prefix:p2p
*/
type P2PApi struct {
	p2pService P2P
}

/*
 name: getPeers
 usage: Get currently connected nodes
 params:
 return: P2P peer-to-peer information connecting with local
 example:  curl http://127.0.0.1:10085 -X POST --data '{"jsonrpc":"2.0","method":"p2p_getPeers","params":"", "id": 3}' -H "Content-Type:application/json"
 response:
   {"jsonrpc":"2.0","id":3,"result":[{},{},{},{}]}
*/
func (p2pApis *P2PApi) GetPeers() []string {
	peersInfo := make([]string, 0)
	peers := p2pApis.p2pService.Peers()

	for _, peer := range peers {
		peersInfo = append(peersInfo, peer.ID().String()+"@"+peer.IP())
	}
	return peersInfo
}

/*
 name: addPeer
 usage: Add peer node
 params: enode://publickey@ip:p2p-Port
 return:nil
 example:  curl http://127.0.0.1:10085 -X POST --data '{"jsonrpc":"2.0","method":"p2p_addPeer","params":["enode://e1b2f83b7b0f5845cc74ca12bb40152e520842bbd0597b7770cb459bd40f109178811ebddd6d640100cdb9b661a3a43a9811d9fdc63770032a3f2524257fb62d@192.168.74.1:55555"], "id": 3}' -H "Content-Type:application/json"
 response:

*/
func (p2pApis *P2PApi) AddPeer(addr string) {
	p2pApis.p2pService.AddPeer(addr)
}

/*
 name: removePeer
 usage: remove peer node
 params:enode://publickey@ip:p2p-port
 return:nil
 example: curl http://127.0.0.1:10085 -X POST --data '{"jsonrpc":"2.0","method":"p2p_removePeer","params":["enode://e1b2f83b7b0f5845cc74ca12bb40152e520842bbd0597b7770cb459bd40f109178811ebddd6d640100cdb9b661a3a43a9811d9fdc63770032a3f2524257fb62d@192.168.74.1:55555"], "id": 3}' -H "Content-Type:application/json"
 response:
*/
func (p2pApis *P2PApi) RemovePeer(addr string) {
	p2pApis.p2pService.RemovePeer(addr)
}

/*
 name: LocalNode
 usage: Need to get local eNode for P2P link
 params:""
 return: local enode
 example: curl http://127.0.0.1:10085 -X POST --data '{"jsonrpc":"2.0","method":"p2p_localNode","params":"", "id": 3}' -H "Content-Type:application/json"
 response:
*/

func (p2pApis *P2PApi) LocalNode() *enode.Node {
	return p2pApis.p2pService.LocalNode()
}
