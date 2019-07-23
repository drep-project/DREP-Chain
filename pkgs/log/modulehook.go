package log

import (
	"github.com/pingcap/errors"
	log "github.com/sirupsen/logrus"
	"io"
	"strconv"
	"sync"
)


// ModuleHook use custom hooks to redefine log input and output, module control, level control
type ModuleHook struct {
	writer        io.Writer
	globalLevel   log.Level
	moduleLevel   map[string]log.Level
	saveFormatter log.Formatter
	printFormat   log.Formatter
	lock          sync.RWMutex
}

func NewMyHook(writer io.Writer, formatter log.Formatter, printFormat log.Formatter) *ModuleHook {
	return &ModuleHook{
		writer:        writer,
		saveFormatter: formatter,
		printFormat:   printFormat,
		moduleLevel:   make(map[string]log.Level),
	}
}

func (hook *ModuleHook) Fire(entry *log.Entry) error {
	hook.lock.RLock()
	defer hook.lock.RUnlock()
	if val, ok := entry.Data[MODULE]; ok {
		if lv, ok2 := hook.moduleLevel[val.(string)]; ok2 {
			if lv < entry.Level {
				return nil
			}
		} else {
			if log.GetLevel() < entry.Level {
				return nil
			}
		}
	}
	hook.saveLog(entry)
	hook.printLog(entry)
	return nil
}

//saveLog use saveformater format save log
func (hook *ModuleHook) saveLog(entry *log.Entry) {
	msg, _ := hook.saveFormatter.Format(entry)
	hook.writer.Write(msg)
}

//printLog use printFormat format log output
func (hook *ModuleHook) printLog(entry *log.Entry) {
	var msg []byte
	_, ok := entry.Data[MODULE]
	if ok {
		delete(entry.Data, MODULE)
		msg, _ = hook.printFormat.Format(entry)
	} else {
		msg, _ = hook.printFormat.Format(entry)
	}
	entry.Logger.Out.Write(msg)
}

func (hook *ModuleHook) Levels() []log.Level {
	return log.AllLevels
}

func (hook *ModuleHook) SetLevel(lvInt log.Level) {
	log.SetLevel(lvInt)
	for key, _ := range hook.moduleLevel {
		hook.moduleLevel[key] = lvInt
	}
}

func (hook *ModuleHook) SetModulesLevel(moduleLevel ...interface{}) error {
	hook.lock.Lock()
	defer hook.lock.Unlock()
	if len(moduleLevel)%2 != 0 {
		return errors.New("err format for SetModulesLevel, eg: key lv")
	}
	for i := 0; i < len(moduleLevel); i++ {
		module := moduleLevel[i].(string)
		i++
		var lv log.Level
		switch t := moduleLevel[i].(type) {
		case log.Level:
			lv = t
		case int:
			lv = log.Level(t)
		case string:
			var err error
			lvInt, err := strconv.ParseInt(t, 10, 64)
			if err != nil {
				lv, err = log.ParseLevel(t)
				if err != nil {
					return err
				}
				return err
			} else {
				lv = log.Level(lvInt)
			}

		default:
			return errors.New("unsport lvl type")
		}
		hook.moduleLevel[module] = lv
	}
	return nil
}

func parserLevel(lvAny interface{}) (log.Level, error) {
	var lv log.Level
	switch t := lvAny.(type) {
	case log.Level:
		lv = t
	case int:
		lv = log.Level(t)
	case string:
		var err error
		lvInt, err := strconv.ParseInt(t, 10, 64)
		if err != nil {
			lv, err = log.ParseLevel(t)
			if err != nil {
				return 0, err
			}
		} else {
			lv = log.Level(lvInt)
		}

	default:
		lv = log.InfoLevel
	}
	return lv, nil
}
