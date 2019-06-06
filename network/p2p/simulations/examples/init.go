package main

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

func NewLog() *logrus.Logger {
	return dlog.NewLogger(MODULENAME)
}