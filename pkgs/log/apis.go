package log

import (
	"github.com/pingcap/errors"
	"github.com/sirupsen/logrus"
	"strconv"
	"strings"
)

/*
name: 日志rpc接口
usage: 设置日志级别
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
 usage: 设置日志级别
 params:
	1. 日志级别（"debug","0"）
 return: 无
 example:  curl http://localhost:15645 -X POST --data '{"jsonrpc":"2.0","method":"log_setLevel","params":["trace"], "id": 3}' -H "Content-Type:application/json"
 response:
  {"jsonrpc":"2.0","id":3,"result":null}
*/
func (logApi *LogApi) SetLevel(lvl string) error {
	lvInt, err := strconv.ParseInt(lvl,10,64);
	if err == nil {
		logrus.SetLevel(logrus.Level(lvInt))
		return nil
	}
	lv, err := logrus.ParseLevel(lvl)
	if err == nil {
		logrus.SetLevel(lv)
		return nil
	}
	return errors.New("not support lvl type ,eg:1, debug")
}

/*
 name: setVmodule
 usage: 分模块设置级别
 params:
	1. 模块日志级别(txpool=5)
 return: 无
 example:   curl http://localhost:15645 -X POST --data '{"jsonrpc":"2.0","method":"log_setVmodule","params":["txpool=5"], "id": 3}' -H "Content-Type:application/json"
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