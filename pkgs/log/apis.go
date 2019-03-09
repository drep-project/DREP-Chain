package log

import "github.com/drep-project/dlog"

type LogApi struct {

}

func (logApi *LogApi) SetLevel(lvl string) error {
	lv, err := dlog.LvlFromString(lvl)
	if err != nil {
		return err
	}
	dlog.SetVerbosity(lv)
	return nil
}

func (logApi *LogApi) SetVmodule(module string) error {
	return dlog.SetVmodule(module)
}

func (logApi *LogApi) SetBackTraceAt(backTraceAt string) error {
	return dlog.SetBacktraceAt(backTraceAt)
}