package rpc

import (
	"github.com/drep-project/drep-chain/app"
	"github.com/drep-project/rpc"
)

type HTTPControl interface {
	StartHTTP(endpoint string, apis []app.API, modules []string, cors []string, vhosts []string, timeouts *rpc.HTTPTimeouts)
	StopHTTP()
}

type RESTControl interface {
	StartRest(endpoint string, restApi rpc.RestDescription) error
	StopREST()
}

type InProcControl interface {
	StartInProc(apis []app.API) error
	StopInProc()
}

type IPCControl interface {
	StartIPC(apis []app.API)
	StopIPC()
}

type WSControl interface {
	StartWS(endpoint string, apis []app.API, modules []string, wsOrigins []string, exposeAll bool) error
	StopWS()
}

type Rpc interface {
	app.Service
	HTTPControl
	RESTControl
	InProcControl
	IPCControl
	WSControl
}
