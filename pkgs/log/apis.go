package log

import (
	"github.com/pingcap/errors"
	"github.com/sirupsen/logrus"
	"strconv"
	"strings"
)

type LogApi struct {
  hook *ModuleHook
}

func NewLogApi(hook *ModuleHook) *LogApi {
	return &LogApi{hook}
}

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