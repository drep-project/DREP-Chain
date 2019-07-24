package app

import "encoding/binary"

type ChainIdType uint32

func (chainId ChainIdType) Bytes() []byte{
	bytes := make([]byte,4)
	binary.BigEndian.PutUint32(bytes, uint32(chainId))
	return bytes
}