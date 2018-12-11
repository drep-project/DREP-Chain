
package config

type LogConfig struct {
	DataDir string 		`json:"-"`
	LogLevel int 		`json:"logLevel"`
	Vmodule string		`json:"vmodule"`
	BacktraceAt string	`json:"backtraceAt"`
}