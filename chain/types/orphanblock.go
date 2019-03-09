package types

import "time"

// orphanBlock represents a block that we don't yet have the parent for.  It
// is a normal block plus an expiration time to prevent caching the orphan
// forever.
type OrphanBlock struct {
	Block      *Block
	Expiration time.Time
}
