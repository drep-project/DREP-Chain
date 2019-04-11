package common

import (
	"math/big"
)

func CalcMemSize(off, l *big.Int) *big.Int {
	if l.Sign() == 0 {
		return Big0
	}
	return new(big.Int).Add(off, l)
}

func ToWordSize(size uint64) uint64 {
	if size > MaxUint64-31 {
		return MaxUint64/32 + 1
	}
	return (size + 31) / 32
}

func GetData(data []byte, start uint64, size uint64) []byte {
	length := uint64(len(data))
	if start > length {
		start = length
	}
	end := start + size
	if end > length {
		end = length
	}
	return RightPadBytes(data[start:end], int(size))
}

func AllZero(b []byte) bool {
	for _, byt := range b {
		if byt != 0 {
			return false
		}
	}
	return true
}

func GetDataBig(data []byte, start *big.Int, size *big.Int) []byte {
	dlen := big.NewInt(int64(len(data)))

	s := BigMin(start, dlen)
	e := BigMin(new(big.Int).Add(s, size), dlen)
	return RightPadBytes(data[s.Uint64():e.Uint64()], int(size.Uint64()))
}