package msg

import (
	"fmt"
	"os"
	"runtime"
)

// Fatal is just like func l.Critical logger except that it is followed by exit to program
func (l *Logger) Fatal(message string) {
	l.logInternal(LCrit, message, 2)
	os.Exit(1)
}

// FatalF is just like func l.CriticalF logger except that it is followed by exit to program
func (l *Logger) FatalF(format string, a ...interface{}) {
	l.logInternal(LCrit, fmt.Sprintf(format, a...), 2)
	os.Exit(1)
}

// Fatalf is just like func l.CriticalF logger except that it is followed by exit to program
func (l *Logger) Fatalf(format string, a ...interface{}) {
	l.logInternal(LCrit, fmt.Sprintf(format, a...), 2)
	os.Exit(1)
}

// Panic is just like func l.Critical except that it is followed by a call to panic
func (l *Logger) Panic(message string) {
	l.logInternal(LCrit, message, 2)
	panic(message)
}

// PanicF is just like func l.CriticalF except that it is followed by a call to panic
func (l *Logger) PanicF(format string, a ...interface{}) {
	l.logInternal(LCrit, fmt.Sprintf(format, a...), 2)
	panic(fmt.Sprintf(format, a...))
}

// Panicf is just like func l.CriticalF except that it is followed by a call to panic
func (l *Logger) Panicf(format string, a ...interface{}) {
	l.logInternal(LCrit, fmt.Sprintf(format, a...), 2)
	panic(fmt.Sprintf(format, a...))
}

// Critical logs a message at a Critical Level
func (l *Logger) Critical(message string) {
	l.logInternal(LCrit, message, 2)
}

// CriticalF logs a message at Critical level using the same syntax and options as fmt.Printf
func (l *Logger) CriticalF(format string, a ...interface{}) {
	l.logInternal(LCrit, fmt.Sprintf(format, a...), 2)
}

// Criticalf logs a message at Critical level using the same syntax and options as fmt.Printf
func (l *Logger) Criticalf(format string, a ...interface{}) {
	l.logInternal(LCrit, fmt.Sprintf(format, a...), 2)
}

// Error logs a message at Error level
func (l *Logger) Error(message string) {
	l.logInternal(LErr, message, 2)
}

// ErrorF logs a message at Error level using the same syntax and options as fmt.Printf
func (l *Logger) ErrorF(format string, a ...interface{}) {
	l.logInternal(LErr, fmt.Sprintf(format, a...), 2)
}

// Errorf logs a message at Error level using the same syntax and options as fmt.Printf
func (l *Logger) Errorf(format string, a ...interface{}) {
	l.logInternal(LErr, fmt.Sprintf(format, a...), 2)
}

// Warning logs a message at Warning level
func (l *Logger) Warning(message string) {
	l.logInternal(LWarn, message, 2)
}

// WarningF logs a message at Warning level using the same syntax and options as fmt.Printf
func (l *Logger) WarningF(format string, a ...interface{}) {
	l.logInternal(LWarn, fmt.Sprintf(format, a...), 2)
}

// Warningf logs a message at Warning level using the same syntax and options as fmt.Printf
func (l *Logger) Warningf(format string, a ...interface{}) {
	l.logInternal(LWarn, fmt.Sprintf(format, a...), 2)
}

// Notice logs a message at Notice level
func (l *Logger) Notice(message string) {
	l.logInternal(LNotice, message, 2)
}

// NoticeF logs a message at Notice level using the same syntax and options as fmt.Printf
func (l *Logger) NoticeF(format string, a ...interface{}) {
	l.logInternal(LNotice, fmt.Sprintf(format, a...), 2)
}

// Noticef logs a message at Notice level using the same syntax and options as fmt.Printf
func (l *Logger) Noticef(format string, a ...interface{}) {
	l.logInternal(LNotice, fmt.Sprintf(format, a...), 2)
}

// Info logs a message at Info level
func (l *Logger) Info(message string) {
	l.logInternal(LInfo, message, 2)
}

// InfoF logs a message at Info level using the same syntax and options as fmt.Printf
func (l *Logger) InfoF(format string, a ...interface{}) {
	l.logInternal(LInfo, fmt.Sprintf(format, a...), 2)
}

// Infof logs a message at Info level using the same syntax and options as fmt.Printf
func (l *Logger) Infof(format string, a ...interface{}) {
	l.logInternal(LInfo, fmt.Sprintf(format, a...), 2)
}

// Debug logs a message at Debug level
func (l *Logger) Debug(message string) {
	l.logInternal(LDebug, message, 2)
}

// DebugF logs a message at Debug level using the same syntax and options as fmt.Printf
func (l *Logger) DebugF(format string, a ...interface{}) {
	l.logInternal(LDebug, fmt.Sprintf(format, a...), 2)
}

// Debugf logs a message at Debug level using the same syntax and options as fmt.Printf
func (l *Logger) Debugf(format string, a ...interface{}) {
	l.logInternal(LDebug, fmt.Sprintf(format, a...), 2)
}

// StackAsError Prints a goroutine's execution stack as an error with an optional message at the begining
func (l *Logger) StackAsError(message string) {
	l.logInternal(LErr, stack(message), 2)
}

// StackAsCritical Prints a goroutine's execution stack as critical with an optional message at the begining
func (l *Logger) StackAsCritical(message string) {
	l.logInternal(LCrit, stack(message), 2)
}

func stack(s string) string {
	if s == "" {
		s = "Stack info\n"
	}
	buf := make([]byte, 1<<16)
	runtime.Stack(buf, false)
	return fmt.Sprintf("%s\n%s", s, buf)
}
