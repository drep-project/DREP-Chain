package filter

type FilterConfig struct {
	Enable		bool	`json:"enable"`
}

var (
	DefaultConfig = &FilterConfig{
		Enable:      true,
	}
)