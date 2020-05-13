package app

import (
	"gopkg.in/urfave/cli.v1"
	"reflect"
)

var (
	TServiceService = reflect.TypeOf((*Service)(nil)).Elem()
	TOrService      = reflect.TypeOf((*OrService)(nil)).Elem()
)

// Services can customize their own configuration, command parameters, interfaces, services
type Service interface {
	Name() string                              // service  name must be unique
	Api() []API                                // Interfaces required for services
	CommandFlags() ([]cli.Command, []cli.Flag) // flags required for services
	//P2pMessages() map[int]interface{}
	//Receive(context actor.Context)
	Init(executeContext *ExecuteContext) error
	Start(executeContext *ExecuteContext) error
	Stop(executeContext *ExecuteContext) error
}

type OrService interface {
	Service
	SelectService() Service
}
