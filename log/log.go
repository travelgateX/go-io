package log

// Output is the package level logger, initialized as StdLogger by default
var Output = NewStdLogger()

func Fatal(m string) {
	Output.Log(LvlFatal, m)
}

func Error(m string) {
	Output.Log(LvlError, m)
}

func Warning(m string) {
	Output.Log(LvlWarning, m)
}

func Info(m string) {
	Output.Log(LvlInfo, m)
}

func Debug(m string) {
	Output.Log(LvlDebug, m)
}

func Verbose(m string) {
	Output.Log(LvlVerbose, m)
}

func Fatalf(format string, a ...interface{}) {
	Output.Logf(LvlFatal, format, a)
}

func Errorf(format string, a ...interface{}) {
	Output.Logf(LvlError, format, a)
}

func Warningf(format string, a ...interface{}) {
	Output.Logf(LvlWarning, format, a)
}

func Infof(format string, a ...interface{}) {
	Output.Logf(LvlInfo, format, a)
}

func Debugf(format string, a ...interface{}) {
	Output.Logf(LvlDebug, format, a)
}

func Verbosef(format string, a ...interface{}) {
	Output.Logf(LvlVerbose, format, a)
}
