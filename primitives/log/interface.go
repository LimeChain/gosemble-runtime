package log

type logLevel int

const target = "runtime"

const (
	CriticalLevel logLevel = iota
	WarnLevel
	InfoLevel
	DebugLevel
	TraceLevel
)

type RuntimeLogger interface {
	Info(message string)
	Infof(message string, a ...any)
	Trace(message string)
	Tracef(message string, a ...any)
	Debug(message string)
	Debugf(message string, a ...any)
	Warn(message string)
	Warnf(message string, a ...any)
	Critical(message string)
	Criticalf(message string, a ...any)
}
