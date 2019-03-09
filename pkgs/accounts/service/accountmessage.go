package service

import (
	"github.com/AsynkronIT/protoactor-go/actor"
)

func (accountService *AccountService) Receive(context actor.Context) {
	/*
	routeMsg, ok := context.Message().(*p2pTypes.RouteIn)
	if !ok {
		return
	}
	switch msg := routeMsg.Detail.(type) {
	}
	*/
}