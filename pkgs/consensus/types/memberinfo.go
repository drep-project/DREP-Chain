package types

type MemberInfo struct {
	Peer     *PeerInfo
	Producer *Producer
	Status   int
	IsMe     bool
	IsLeader bool
	IsOnline bool
}
