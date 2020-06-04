package blockmgr

// BlockMgrConfig defines gasprice & journal file type.
type BlockMgrConfig struct {
	GasPrice    OracleConfig `json:"gasprice"`
	JournalFile string       `json:"journalFile"`
}

// OracleConfig manages gas price of block.
type OracleConfig struct {
	Blocks     int    `json:"blocks"`
	Percentile int    `json:"percentile"`
	Default    uint64 `json:"default"`
	MaxPrice   uint64 `json:"maxPrice"`
}
