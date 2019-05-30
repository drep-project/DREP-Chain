package trace

// HistoryConfig used to condig history data dir and db message
type HistoryConfig struct {
	HistoryDir string			 `json:"historydir"`
	Url        string			 `json:"url"`
	DbType     string			 `json:"dbtype"`
	Enable     bool				 `json:"enable"`
}