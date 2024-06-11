//go:build nonwasmenv

package log

import "fmt"

type logger struct{}

func NewLogger() RuntimeLogger {
	return logger{}
}

func (l logger) Critical(message string) {
	l.log(CriticalLevel, []byte(target), []byte(message))
	panic(message)
}

func (l logger) Criticalf(message string, a ...any) {
	l.Critical(fmt.Sprintf(message, a...))
}

func (l logger) Warn(message string) {
	l.log(WarnLevel, []byte(target), []byte(message))
}

func (l logger) Warnf(message string, a ...any) {
	l.Warn(fmt.Sprintf(message, a...))
}

func (l logger) Info(message string) {
	l.log(InfoLevel, []byte(target), []byte(message))
}

func (l logger) Infof(message string, a ...any) {
	l.Info(fmt.Sprintf(message, a...))
}

func (l logger) Debug(message string) {
	l.log(DebugLevel, []byte(target), []byte(message))
}

func (l logger) Debugf(message string, a ...any) {
	l.Debug(fmt.Sprintf(message, a...))
}

func (l logger) Trace(message string) {
	l.log(TraceLevel, []byte(target), []byte(message))
}

func (l logger) Tracef(message string, a ...any) {
	l.Trace(fmt.Sprintf(message, a...))
}

func (l logger) log(level logLevel, target []byte, message []byte) {
	fmt.Println(fmt.Sprintf("%s  target=%s  message=%s", level.string(), string(target), string(message)))
}

func (level logLevel) string() string {
	switch level {
	case CriticalLevel:
		return "CRITICAL"
	case WarnLevel:
		return "WARN"
	case InfoLevel:
		return "INFO"
	case DebugLevel:
		return "DEBUG"
	case TraceLevel:
		return "TRACE"
	default:
		return ""
	}
}
