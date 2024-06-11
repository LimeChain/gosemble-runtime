//go:build !nonwasmenv

package log

import (
	"fmt"

	"github.com/LimeChain/gosemble/env"
	"github.com/LimeChain/gosemble/utils"
)

type logger struct {
	memUtils utils.WasmMemoryTranslator
}

func NewLogger() RuntimeLogger {
	return logger{
		memUtils: utils.NewMemoryTranslator(),
	}
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
	targetOffsetSize := l.memUtils.BytesToOffsetAndSize(target)
	messageOffsetSize := l.memUtils.BytesToOffsetAndSize(message)
	env.ExtLoggingLogVersion1(int32(level), targetOffsetSize, messageOffsetSize)
}
