package role

import (
    "BlockChainTest/bean"
    "BlockChainTest/network"
    "BlockChainTest/crypto"
    "BlockChainTest/log"
)

type User struct {
    PubKey  *bean.Point
    Address bean.Address
    peers   []*network.Peer
}

func NewUser(pubKey *crypto.Point, peers []*network.Peer) *User {
    log.Println("ready for process newcomers")
    m := &User{}
    m.PubKey = pubKey
    m.peers = peers
    return m
}

func (n *User) ProcessNewComers(newcomer *bean.Newcomer)  {
    log.Println("user starting process a newcomer")
    pubKey := newcomer.Pk
    address := bean.Addr(pubKey)

    log.Println("starting generate a new peer from the newcomer")
    newPeer := &network.Peer{}
    newPeer.IP = network.IP(newcomer.Ip)
    newPeer.Port = network.Port(newcomer.Port)

    newPeer.PubKey = pubKey
    newPeer.Address = address


    log.Println(" adding the newcomer to a map table.")
    // add newcomer to the map table.

    peerStore := network.GetStore()
    peerStore.AddPeer(newPeer)

    list := make([]*bean.Newcomer, 0)
    peers := make([]*network.Peer, 0)
    for _, value := range peerStore.Store {
        msg := &bean.Newcomer{}
        msg.Pk = value.PubKey
        list = append(list, msg)

        peer := &network.Peer{}
        peer.PubKey = value.PubKey
        peer.Address = address
        peer.IP = value.IP
        peer.Port = value.Port
        peers = append(peers, peer)
    }

    listOfPeer := &bean.ListOfPeer{}
    listOfPeer.List = list

    // return the list to newcomer
    log.Println("send a list of peer to the newcomer")
    task := network.Task{newPeer,listOfPeer}

    task.SendMessageCore()

    // broadcast the new comer msg
    network.SendMessage(n.peers, newcomer)
}
