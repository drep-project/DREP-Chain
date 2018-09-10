package store

import (
    "BlockChainTest/network"
    "BlockChainTest/fetch_block"
)

var (
    sender *fetch_block.BlockSender
    receiver *fetch_block.BlockReceiver
)

func init() {
    sender_ip := network.IP("192.168.3.43")
    sender_port := network.Port(55556)
    sender_peer := &network.Peer{IP: sender_ip, Port: sender_port}
    receiver_ip := network.IP("192.168.3.43")
    receiver_port := network.Port(55555)
    receiver_peer := &network.Peer{IP: receiver_ip, Port: receiver_port}
    sender = fetch_block.NewBlockSender(receiver_peer)
    receiver = fetch_block.NewBlockReceiver(sender_peer)
}

func GetItselfOnSender() *fetch_block.BlockSender {
    return sender
}

func GetItselfOnReceiver() *fetch_block.BlockReceiver {
    return receiver
}