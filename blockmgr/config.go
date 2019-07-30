package blockmgr

type BlockMgrConfig struct {
	GasPrice    OracleConfig `json:"gasprice"`
	JournalFile string       `json:"journalFile"`
}

type OracleConfig struct {
	Blocks     int    `json:"blocks"`
	Percentile int    `json:"percentile"`
	Default    uint64 `json:"default"`
	MaxPrice   uint64 `json:"maxPrice"`
}
