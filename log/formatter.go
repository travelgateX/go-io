package log

import (
	"bytes"
)

type Formatter interface {
	Format(*bytes.Buffer, string, Level)
}

type TextFormatter struct{}

func (TextFormatter) Format(buf *bytes.Buffer, m string, lvl Level) {
	const (
		red    = "\x1b[31;1m"
		yellow = "\x1b[33;1m"
		blue   = "\x1b[34;1m"
		white  = "\x1b[37;1m"
		cyan   = "\x1b[36;1m"
		reset  = "\x1b[0m"
	)

	var prefix string
	switch lvl {
	case LvlFatal:
		prefix = red + "FATAL "
	case LvlError:
		prefix = red + "ERROR " + reset
	case LvlWarning:
		prefix = yellow + "WARNING " + reset
	case LvlInfo:
		prefix = blue + "INFO " + reset
	case LvlDebug:
		prefix = white + "DEBUG " + reset
	case LvlVerbose:
		prefix = cyan + "VERBOSE " + reset
	}
	buf.WriteString(prefix + m + "\n")
}
