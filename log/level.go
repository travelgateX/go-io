package log

import (
	"fmt"
	"strings"
)

type Level int

const (
	LvlFatal   Level = iota // critical errors causing complete failure of the application that will lead the application to abort
	LvlError                // Any error which is fatal to the operation, but not the service or application
	LvlWarning              // indicators of possible issues or service/functionality degradation
	LvlInfo                 // events of interest or that have relevance to outside observers
	LvlDebug                // for testing purposes
	LvlVerbose
)

func (lvl Level) String() string {
	switch lvl {
	case LvlFatal:
		return "fatal"
	case LvlError:
		return "error"
	case LvlWarning:
		return "warning"
	case LvlInfo:
		return "info"
	case LvlDebug:
		return "debug"
	case LvlVerbose:
		return "verbose"
	default:
		return "undefined"
	}
}

func ToLevel(description string) (lvl Level, err error) {
	switch strings.ToLower(description) {
	case "fatal":
		lvl = LvlFatal
	case "error":
		lvl = LvlError
	case "warning":
		lvl = LvlWarning
	case "info":
		lvl = LvlInfo
	case "debug":
		lvl = LvlDebug
	case "verbose":
		lvl = LvlVerbose
	default:
		err = fmt.Errorf("Level unrecognized: " + description)
	}
	return
}
