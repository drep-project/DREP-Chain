package log

type LogApi struct {

}

func (logApi *LogApi) SetLevel(lvl string) error {
	lv, err := LvlFromString(lvl)
	if err != nil {
		return err
	}
	glogger.Verbosity(lv)
	return nil
}

func (logApi *LogApi) SetVmodule(module string) error {
	return glogger.Vmodule(module)
}

func (logApi *LogApi) SetBackTraceAt(backTraceAt string) error {
	return glogger.BacktraceAt(backTraceAt)
}