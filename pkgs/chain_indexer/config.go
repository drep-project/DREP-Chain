package chain_indexer

import (
	"time"
)

type ChainIndexerConfig struct {
	Enable      bool			`json:"enable"`
	SectionSize uint64			`json:"sectionsize"` // Number of blocks in a single chain segment to process
	ConfirmsReq uint64			`json:"confirmsreq"` // Number of confirmations before processing a completed segment
	Throttling	time.Duration	`json:"throttling"` // Disk throttling to prevent a heavy upgrade from hogging resources
}

var (
	DefaultConfig = &ChainIndexerConfig{
		Enable: true,
		SectionSize: 4096,
		ConfirmsReq: 256,
		Throttling: 100 * time.Millisecond,
	}
)
