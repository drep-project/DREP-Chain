package log

import (
	"io"
    "os"
    
	colorable "github.com/mattn/go-colorable"
	"github.com/mattn/go-isatty"
)
var DEBUG = false


var (
	ostream Handler
	glogger *GlogHandler
)

func init() {
	usecolor := (isatty.IsTerminal(os.Stderr.Fd()) || isatty.IsCygwinTerminal(os.Stderr.Fd())) && os.Getenv("TERM") != "dumb"
	output := io.Writer(os.Stderr)
	if usecolor {
		output = colorable.NewColorableStderr()
	}
	ostream = StreamHandler(output, TerminalFormat(usecolor))
	glogger = NewGlogHandler(ostream)
}


func SetUp(logdir string) error {
	if logdir != "" {
		rfh, err := RotatingFileHandler(
			logdir,
			262144,
			JSONFormatOrderedEx(false, true),
		)
		if err != nil {
			return err
		}
		glogger.SetHandler(MultiHandler(ostream, rfh))
	}
	glogger.Verbosity(Lvl(3))
	//glogger.Vmodule(ctx.GlobalString(vmoduleFlag.Name))
	//glogger.BacktraceAt(ctx.GlobalString(backtraceAtFlag.Name))
	Root().SetHandler(glogger)
	return nil
}
