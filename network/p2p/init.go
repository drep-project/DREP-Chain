package p2p

import (
	dlog "github.com/drep-project/drep-chain/pkgs/log"
	"github.com/sirupsen/logrus"
)

const (
	MODULENAME = "p2p"
)

var (
	log = dlog.EnsureLogger(MODULENAME)
)

func NewLog() *logrus.Entry {
	return dlog.EnsureLogger(MODULENAME)
}
