/*
 * Copyright (C) 2018 The ontology Authors
 * This file is part of The ontology library.
 *
 * The ontology is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The ontology is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The ontology.  If not, see <http://www.gnu.org/licenses/>.
 */

package types

import (
	"fmt"
	"github.com/AsynkronIT/protoactor-go/actor"
)

type RouteIn struct {
	Type int
	Peer *Peer
	Detail interface{}
}
// MessageRouter mostly route different message type-based to the
// related message handler
type MessageRouter struct {
	msgHandlers  map[int]*actor.PID // Msg handler mapped to msg type
	recMsgCh chan *RouteIn    // The channel to handle consensus msg
	stopCh   chan bool                 // To stop consensus channel
	pid *actor.PID

	consensusPid *actor.PID
	chainPid *actor.PID
}

// NewMsgRouter returns a message router object
func NewMsgRouter(recMsgCh chan *RouteIn) *MessageRouter {
	msgRouter := &MessageRouter{}
	msgRouter.init(recMsgCh)
	go msgRouter.handleP2pMessage()
	return msgRouter
}

// init initializes the message router's attributes
func (this *MessageRouter) init(recMsgCh chan *RouteIn) {
	this.msgHandlers = make(map[int]*actor.PID)
	this.recMsgCh = recMsgCh
	this.stopCh = make(chan bool)
}

// RegisterMsgHandler registers msg handler with the msg type
func (this *MessageRouter) RegisterMsgHandler(key int, handler *actor.PID) {
	this.msgHandlers[key] = handler
}

// UnRegisterMsgHandler un-registers the msg handler with
// the msg type
func (this *MessageRouter) UnRegisterMsgHandler(key int) {
	delete(this.msgHandlers, key)
}

// SetPID sets p2p actor
func (this *MessageRouter) SetPID(pid *actor.PID) {
	this.pid = pid
}

// Start starts the loop to handle the message from the network
func (this *MessageRouter) Start() {

}

func (this *MessageRouter) handleP2pMessage(){
	for {
		select {
		case data, ok := <-this.recMsgCh:
			if ok {
				handler, ok := this.msgHandlers[data.Type]
				if ok {
					 handler.Tell(data)
				} else {
					fmt.Println("unknown message handler for the msg: ",  data.Type)
				}
			}
		case <-this.stopCh:
			return
		}
	}
}


// Stop stops the message router's loop
func (this *MessageRouter) Stop() {
	if this.stopCh != nil {
		this.stopCh <- true
	}
}
