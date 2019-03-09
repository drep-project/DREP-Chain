package app

import (
	"encoding/hex"
)

const (
	ChainIdSize = 64
)

type ChainIdType [ChainIdSize]byte

func (c ChainIdType) Hex() string {
	return hex.EncodeToString(c[:])
}

func (c *ChainIdType) SetBytes(b []byte) {
	bytes := []byte{}
	hex.Decode(b, bytes)
	if len(bytes) > len(c) {
		copy(c[:], b[len(bytes)-ChainIdSize:])
	} else {
		copy(c[ChainIdSize-len(bytes):], bytes)
	}
}

func (c ChainIdType) MarshalText() ([]byte, error) {
	return []byte(c.Hex()), nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *ChainIdType) UnmarshalJSON(input []byte) error {
	return c.UnmarshalText(input[1 : len(input)-1])
}

// UnmarshalText implements encoding.TextUnmarshaler
func (c *ChainIdType) UnmarshalText(input []byte) error {
	c.SetBytes(input)
	return nil
}
