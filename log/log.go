package log

// package level logger, initialized with as StdLogger by default
var std = NewStdLogger()

func Fatal(m string) {
	std.Log(LvlFatal, m)
}

func Error(m string) {
	std.Log(LvlError, m)
}

func Warning(m string) {
	std.Log(LvlWarning, m)
}

func Info(m string) {
	std.Log(LvlInfo, m)
}

func Debug(m string) {
	std.Log(LvlDebug, m)
}

func Verbose(m string) {
	std.Log(LvlVerbose, m)
}

func Fatalf(format string, a ...interface{}) {
	std.Logf(LvlFatal, format, a)
}

func Errorf(format string, a ...interface{}) {
	std.Logf(LvlError, format, a)
}

func Warningf(format string, a ...interface{}) {
	std.Logf(LvlWarning, format, a)
}

func Infof(format string, a ...interface{}) {
	std.Logf(LvlInfo, format, a)
}

func Debugf(format string, a ...interface{}) {
	std.Logf(LvlDebug, format, a)
}

func Verbosef(format string, a ...interface{}) {
	std.Logf(LvlVerbose, format, a)
}
