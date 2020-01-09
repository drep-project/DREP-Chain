package service

import "github.com/drep-project/DREP-Chain/network/p2p/enode"

/*
name: p2p网络接口
usage: 设置查询网络状态
prefix:p2p
*/
type P2PApi struct {
	p2pService P2P
}

/*
 name: getPeers
 usage: 获取当前连接的节点
 params:
 return: 和本地建立连接的p2p对端信息
 example:  curl http://127.0.0.1:15645 -X POST --data '{"jsonrpc":"2.0","method":"p2p_getPeers","params":"", "id": 3}' -H "Content-Type:application/json"
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
 name: addPeers
 usage: 添加节点
 params: enode://publickey@ip:p2p端口
 return:nil
 example:  curl http://127.0.0.1:15645 -X POST --data '{"jsonrpc":"2.0","method":"p2p_addPeers","params":["enode://e1b2f83b7b0f5845cc74ca12bb40152e520842bbd0597b7770cb459bd40f109178811ebddd6d640100cdb9b661a3a43a9811d9fdc63770032a3f2524257fb62d@192.168.74.1:55555"], "id": 3}' -H "Content-Type:application/json"
 response:

*/
func (p2pApis *P2PApi) AddPeers(addr string) {
	p2pApis.p2pService.AddPeer(addr)
}

/*
 name: removePeers
 usage: 移除节点
 params:enode://publickey@ip:p2p端口
 return:nil
 example: curl http://127.0.0.1:15645 -X POST --data '{"jsonrpc":"2.0","method":"p2p_addPeers","params":["enode://e1b2f83b7b0f5845cc74ca12bb40152e520842bbd0597b7770cb459bd40f109178811ebddd6d640100cdb9b661a3a43a9811d9fdc63770032a3f2524257fb62d@192.168.74.1:55555"], "id": 3}' -H "Content-Type:application/json"
 response:
*/
func (p2pApis *P2PApi) RemovePeers(addr string) {
	p2pApis.p2pService.RemovePeer(addr)
}

/*
 name: LocalNode
 usage: 需要获取本地的enode，用于P2p链接
 params:""
 return: 本地节点的enode
 example: curl http://127.0.0.1:15645 -X POST --data '{"jsonrpc":"2.0","method":"p2p_localNode","params":"", "id": 3}' -H "Content-Type:application/json"
 response:
*/

func (p2pApis *P2PApi) LocalNode() *enode.Node {
	return p2pApis.p2pService.LocalNode()
}
