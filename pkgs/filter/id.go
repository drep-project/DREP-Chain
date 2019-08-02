package filter

import (
	"bufio"
	crand "crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"math/rand"
	"strings"
	"sync"
	"time"
)

var globalGen = randomIDGenerator()

type ID string

// NewID returns a new, random ID.
func NewID() ID {
	return globalGen()
}

// randomIDGenerator returns a function generates a random IDs.
func randomIDGenerator() func() ID {
	seed, err := binary.ReadVarint(bufio.NewReader(crand.Reader))
	if err != nil {
		seed = int64(time.Now().Nanosecond())
	}
	var (
		mu  sync.Mutex
		rng = rand.New(rand.NewSource(seed))
	)
	return func() ID {
		mu.Lock()
		defer mu.Unlock()
		id := make([]byte, 16)
		rng.Read(id)
		return encodeID(id)
	}
}

func encodeID(b []byte) ID {
	id := hex.EncodeToString(b)
	id = strings.TrimLeft(id, "0")
	if id == "" {
		id = "0" // ID's are RPC quantities, no leading zero's and 0 is 0x0.
	}
	return ID("0x" + id)
}
