package log

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sync"
)

var bufferPool = &sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

type Setup struct {
	F Formatter
	W io.Writer
}

type Fields map[string]interface{}

type Logger struct {
	Setups      []Setup
	FatalSetups []Setup
	f           Fields
	// only those logs with level lower or equal than this will be registered
	MaxLvl Level
}

func NewStdLogger() *Logger {
	s := []Setup{
		{
			F: TextFormatter{},
			W: os.Stdout,
		},
	}
	return &Logger{
		Setups:      s,
		FatalSetups: s,
		MaxLvl:      LvlVerbose,
	}
}

// WithFields returns a new instance of logger with fields
func (l *Logger) WithFields(f Fields) *Logger {
	newl := *l
	newl.f = f
	return &newl
}

func (l *Logger) Log(lvl Level, m string) {
	if lvl > l.MaxLvl {
		return
	}

	var setups []Setup
	if lvl == LvlFatal {
		setups = l.FatalSetups
	} else {
		setups = l.Setups
	}

	for _, s := range setups {
		b := bufferPool.Get().(*bytes.Buffer)
		s.F.Format(b, m, lvl, l.f)
		s.W.Write(b.Bytes())
		b.Reset()
		bufferPool.Put(b)
	}
}

func (l *Logger) Logf(lvl Level, format string, a ...interface{}) {
	if lvl > l.MaxLvl {
		return
	}

	m := fmt.Sprintf(format, a...)
	l.Log(lvl, m)
}

func (l *Logger) Fatal(m string) {
	l.Log(LvlFatal, m)
}

func (l *Logger) Error(m string) {
	l.Log(LvlError, m)
}

func (l *Logger) Warning(m string) {
	l.Log(LvlWarning, m)
}

func (l *Logger) Info(m string) {
	l.Log(LvlInfo, m)
}

func (l *Logger) Debug(m string) {
	l.Log(LvlDebug, m)
}

func (l *Logger) Verbose(m string) {
	l.Log(LvlVerbose, m)
}

func (l *Logger) Fatalf(format string, a ...interface{}) {
	l.Logf(LvlFatal, format, a...)
}

func (l *Logger) Errorf(format string, a ...interface{}) {
	l.Logf(LvlError, format, a...)
}

func (l *Logger) Warningf(format string, a ...interface{}) {
	l.Logf(LvlWarning, format, a...)
}

func (l *Logger) Infof(format string, a ...interface{}) {
	l.Logf(LvlInfo, format, a...)
}

func (l *Logger) Debugf(format string, a ...interface{}) {
	l.Logf(LvlDebug, format, a...)
}

func (l *Logger) Verbosef(format string, a ...interface{}) {
	l.Logf(LvlVerbose, format, a...)
}
