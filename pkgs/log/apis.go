package log

import (
	"github.com/pingcap/errors"
	"github.com/sirupsen/logrus"
	"strconv"
	"strings"
)

/*
name: Logging RPC Api
usage: Set the log level
prefix:log
*/
type LogApi struct {
	hook *ModuleHook
}

func NewLogApi(hook *ModuleHook) *LogApi {
	return &LogApi{hook}
}

/*
 name: setLevel
 usage: Set the log level
 params:
	1. log level（"debug","0"）
 return: 无
 example:  curl http://localhost10085 -X POST --data '{"jsonrpc":"2.0","method":"log_setLevel","params":["trace"], "id": 3}' -H "Content-Type:application/json"
 response:
  {"jsonrpc":"2.0","id":3,"result":null}
*/
func (logApi *LogApi) SetLevel(lvl string) error {
	lvInt, err := strconv.ParseInt(lvl, 10, 64)
	if err == nil {
		logApi.hook.SetLevel(logrus.Level(lvInt))
		return nil
	}
	lv, err := parserLevel(lvl)
	if err != nil {
		return nil
	}
	logApi.hook.SetLevel(lv)
	return nil
}

/*
 name: setVmodule
 usage: Set the level by module
 params:
	1. module name (txpool=5)
 return: 无
 example:   curl http://localhost10085 -X POST --data '{"jsonrpc":"2.0","method":"log_setVmodule","params":["txpool=5"], "id": 3}' -H "Content-Type:application/json"
 response:
  {"jsonrpc":"2.0","id":3,"result":null}
*/
func (logApi *LogApi) SetVmodule(module string) error {
	args := []interface{}{}
	pairs := strings.Split(module, ";")
	for _, pair := range pairs {
		k_v := strings.Split(pair, "=")
		if len(k_v) != 2 {
			return errors.New("not correct module format")
		}
		args = append(args, k_v[0])
		args = append(args, k_v[1])
	}
	return logApi.hook.SetModulesLevel(args...)
}
