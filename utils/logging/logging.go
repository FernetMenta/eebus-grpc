// SPDX-License-Identifier: Apache-2.0
// Copyright Pionix GmbH and Contributors to EVerest

package logging

import (
	"fmt"
	"strings"
	"sync/atomic"
	"time"
)

type Logger struct{}

type LogLevel int32

const (
	LogLevelTrace LogLevel = iota
	LogLevelDebug
	LogLevelInfo
	LogLevelError
)

// currentLevel controls the minimum severity that gets printed.
var currentLevel atomic.Int32

func init() {
	currentLevel.Store(int32(LogLevelTrace))
}

// SetLevel sets the minimum log level. Messages below this level are discarded.
func SetLevel(level LogLevel) {
	currentLevel.Store(int32(level))
}

// GetLevel returns the current minimum log level.
func GetLevel() LogLevel {
	return LogLevel(currentLevel.Load())
}

// ParseLevel converts a string (case-insensitive) to a LogLevel.
// Returns LogLevelInfo if the string is not recognized.
func ParseLevel(s string) LogLevel {
	switch strings.ToLower(s) {
	case "trace":
		return LogLevelTrace
	case "debug":
		return LogLevelDebug
	case "info":
		return LogLevelInfo
	case "error":
		return LogLevelError
	default:
		return LogLevelInfo
	}
}

func (log_level LogLevel) String() string {
	switch log_level {
	case LogLevelTrace:
		return "TRACE"
	case LogLevelDebug:
		return "DEBUG"
	case LogLevelInfo:
		return "INFO"
	case LogLevelError:
		return "ERROR"
	default:
		fmt.Print("Unknown log level")
		return "UNKNOWN"
	}
}

func currentTimestamp() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func log(log_level LogLevel, args ...interface{}) {
	if log_level < LogLevel(currentLevel.Load()) {
		return
	}
	value := fmt.Sprintln(args...)
	fmt.Printf("%s %s %s", currentTimestamp(), log_level.String(), value)
}

func logf(log_level LogLevel, format string, args ...interface{}) {
	if log_level < LogLevel(currentLevel.Load()) {
		return
	}
	value := fmt.Sprintf(format, args...)
	fmt.Printf("%s %s %s\n", currentTimestamp(), log_level.String(), value)
}

func (logger *Logger) Trace(args ...interface{}) {
	Trace(args...)
}

func (logger *Logger) Tracef(format string, args ...interface{}) {
	Tracef(format, args...)
}

func (logger *Logger) Debug(args ...interface{}) {
	Debug(args...)
}

func (logger *Logger) Debugf(format string, args ...interface{}) {
	Debugf(format, args...)
}

func (logger *Logger) Info(args ...interface{}) {
	Info(args...)
}

func (logger *Logger) Infof(format string, args ...interface{}) {
	Infof(format, args...)
}

func (logger *Logger) Error(args ...interface{}) {
	Error(args...)
}

func (logger *Logger) Errorf(format string, args ...interface{}) {
	Errorf(format, args...)
}

func Trace(args ...interface{}) {
	log(LogLevelTrace, args...)
}

func Tracef(format string, args ...interface{}) {
	logf(LogLevelTrace, format, args...)
}

func Debug(args ...interface{}) {
	log(LogLevelDebug, args...)
}

func Debugf(format string, args ...interface{}) {
	logf(LogLevelDebug, format, args...)
}

func Info(args ...interface{}) {
	log(LogLevelInfo, args...)
}

func Infof(format string, args ...interface{}) {
	logf(LogLevelInfo, format, args...)
}

func Error(args ...interface{}) {
	log(LogLevelError, args...)
}

func Errorf(format string, args ...interface{}) {
	logf(LogLevelError, format, args...)
}
