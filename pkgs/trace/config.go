package trace

// HistoryConfig used to condig history data dir and db message
type HistoryConfig struct {
	HistoryDir string
	Url        string
	DbType     string
	Enable     bool				 `json:"enableTrace"`
}