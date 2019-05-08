package trace

import (
	"gopkg.in/urfave/cli.v1"
)

var (
	HistoryDirFlag = cli.StringFlag{
		Name:  "historydir",
		Usage: "directory to save history data",
		Value: "",
	}
)