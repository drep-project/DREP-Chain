package filter

import (
	"gopkg.in/urfave/cli.v1"
)

var (
	EnableFilterFlag = cli.BoolFlag{
		Name:  "enableFilter",
		Usage: "enable Filter",
	}
)