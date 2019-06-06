package nat

import (
	dlog "github.com/drep-project/drep-chain/pkgs/log"
	"github.com/sirupsen/logrus"
)

const (
	MODULENAME = "p2p"
)

var (
	log = dlog.NewLogger(MODULENAME)
)

func NewLog() *logrus.Entry {
	return dlog.NewLogger(MODULENAME)
}