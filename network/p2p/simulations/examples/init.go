package main

import (
	dlog "github.com/drep-project/DREP-Chain/pkgs/log"
	"github.com/sirupsen/logrus"
)

const (
	MODULENAME = "p2p"
)

var (
	log = dlog.EnsureLogger(MODULENAME)
)

func NewLog() *logrus.Logger {
	return dlog.EnsureLogger(MODULENAME)
}
