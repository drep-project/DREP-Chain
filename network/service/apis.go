package service

import (
	"github.com/drep-project/drep-chain/network/p2p"
)

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
 return: 交易字节信息
 example:  curl http://localhost:15645 -X POST --data '{"jsonrpc":"2.0","method":"trace_getRawTransaction","params":["0x00001c9b8c8fdb1f53faf02321f76253704123e2b56cce065852bab93e526ae2"], "id": 3}' -H "Content-Type:application/json"
 response:
   {"jsonrpc":"2.0","id":3,"result":[{},{},{},{}]}
*/
func (p2pApis *P2PApi) GetPeers() []string {
	peersInfo := make([]string,0)
	peers := p2pApis.p2pService.Peers()

	for _,peer := range peers{
		peersInfo = append(peersInfo, peer.ID().String() + "@"+ peer.IP())
	}
	return peersInfo
} 

/*
 name: addPeers（未实现）
 usage: 添加节点
 params:
 return:
 example:
 response:

*/
func (p2pApis *P2PApi) AddPeers(addr string) {
	p2pApis.p2pService.AddPeer(addr)
}

/*
 name: removePeers（未实现）
 usage: 移除节点
 params:
 return:
 example:
 response:
*/
func (p2pApis *P2PApi) RemovePeers(addr string) {
	p2pApis.p2pService.RemovePeer(addr)
}
