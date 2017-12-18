package logging

import (
	"io"
	"log"
)

type Logger interface {
	Print(v ...interface{})
	Printf(format string, v ...interface{})
	Println(v ...interface{})
}

type NopLogger struct{}

func (NopLogger) Print(v ...interface{})                 {}
func (NopLogger) Printf(format string, v ...interface{}) {}
func (NopLogger) Println(v ...interface{})               {}

type Level int

const (
	LQuiet Level = iota
	LError
	LWarn
	LInfo
	LDebug
)

type Log struct {
	err, warn, info, debug Logger
}

func (l Log) Error() Logger { return l.err }
func (l Log) Warn() Logger  { return l.warn }
func (l Log) Info() Logger  { return l.info }
func (l Log) Debug() Logger { return l.debug }

func buildLogger(out io.Writer, prefix string, lWant, lHave Level) Logger {
	if lWant <= lHave {
		return log.New(out, prefix, log.LstdFlags)
	} else {
		return NopLogger{}
	}
}

func NewLog(out io.Writer, level Level) *Log {
	l := new(Log)

	l.err = buildLogger(out, "[Err]", LError, level)
	l.warn = buildLogger(out, "[Warn]", LWarn, level)
	l.info = buildLogger(out, "[Info]", LInfo, level)
	l.debug = buildLogger(out, "[Debug]", LDebug, level)

	return l
}
