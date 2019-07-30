package chain_indexer

import (
	"gopkg.in/urfave/cli.v1"
)

var (
	EnableChainIndexerFlag = cli.BoolFlag{
		Name:  "enableChainIndexer",
		Usage: "enable ChainIndexer",
	}
)
