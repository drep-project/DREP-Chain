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

type Config struct {
	DataDir string
	LogLevel int 
	Vmodule string
	BacktraceAt string
}

func SetUp(cfg *Config) error {
	if cfg.DataDir != "" {
		rfh, err := RotatingFileHandler(
			cfg.DataDir,
			262144,
			JSONFormatOrderedEx(false, true),
		)
		if err != nil {
			return err
		}
		glogger.SetHandler(MultiHandler(ostream, rfh))
	}
	glogger.Verbosity(Lvl(cfg.LogLevel))
	glogger.Vmodule(cfg.Vmodule)
	glogger.BacktraceAt(cfg.BacktraceAt)
	Root().SetHandler(glogger)
	return nil
}
