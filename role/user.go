package role

import (
    "BlockChainTest/bean"
    "BlockChainTest/network"
    "BlockChainTest/crypto"
)

type User struct {
    PubKey  *bean.Point
    Address bean.Address
    peers   []*network.Peer
}

func NewUser(pubKey *crypto.Point, peers []*network.Peer) *User {
    m := &User{}
    m.PubKey = pubKey
    m.peers = peers
    return m
}

func (n *User) ProcessNewComers(newcomer *bean.Newcomer)  {
    pubKey := newcomer.Pk
    address := bean.Addr(pubKey)

    newPeer := &network.Peer{}
    newPeer.PubKey = pubKey
    newPeer.Address = address

    peerStore := network.GetStore()

    list := make([]*bean.Newcomer, 0)

    for _, value := range peerStore.Store {
        msg := &bean.Newcomer{}
        msg.Pk = value.PubKey
        list = append(list, msg)
    }

    listOfPeer := &bean.ListOfPeer{}
    listOfPeer.List = list

    // return the list to newcomer
    task := network.Task{newPeer,listOfPeer}
    task.SendMessageCore()

    // add newcomer to the map table.
    peerStore.Store[address] = newPeer

    // broadcast the new comer msg
    //peers := store.GetPeers()
    //network.SendMessage(peers, newcomer)
}
