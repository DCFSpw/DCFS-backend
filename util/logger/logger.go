package logger

import (
	"fmt"
	"github.com/fatih/color"
	"os"
	"runtime"
	"strings"
)

type logger struct {
	scopes   []string
	logLevel int
}

// GetLogLevel - get current logging level
//
// return type:
//   - int - current logging level
func (logger *logger) GetLogLevel() int {
	return logger.logLevel
}

// SetLogLevel - set current logging level
//
// params:
//   - level int - new logging level
func (logger *logger) SetLogLevel(level int) {
	logger.logLevel = level
}

// SetScopes - set the list of scopes to log
//
// params:
//   - _scopes []string - array of scopes
func (logger *logger) SetScopes(_scopes []string) {
	if logger.scopes == nil {
		logger.scopes = make([]string, 0)
	}

	logger.scopes = append(logger.scopes, _scopes...)
}

// inScope - check if provided scope is in the list of scopes
//
// params:
//   - _scope string - scope of the message
//
// return type:
//   - bool - true if scope is in the list of scopes, false otherwise
func (logger *logger) inScope(_scope string) bool {
	for _, scope := range logger.scopes {
		if scope == _scope || scope == "all" {
			return true
		}
	}

	return false
}

// log - save a message to the logs
//
// params:
//   - msg string - message to log
//   - color *color.Color - color of the message in the console
//   - scope string - scope of the message
func (logger *logger) log(msg string, color *color.Color, scope string) {
	// prepare the file path
	_, file, line, _ := runtime.Caller(2)
	basePath, _ := os.Getwd()
	basePath = strings.ReplaceAll(basePath, "\\", "/")
	file = strings.ReplaceAll(file, basePath, "")

	// remove the first /
	file = strings.Replace(file, "/", "", 1)

	_, err := color.Printf("[%s][%s:%d] %s\n", scope, file, line, msg)
	if err != nil {
		println(fmt.Sprintf("[%s][%s:%d] %s\n", scope, file, line, msg))
	}
}

// Debug - log a debug message
//
// params:
//   - scope string - scope of the message
//   - messages ...string - messages to log
func (logger *logger) Debug(scope string, messages ...string) {
	if logger.logLevel < 2 || !logger.inScope(scope) {
		return
	}

	message := ""
	for _, _msg := range messages {
		message = message + _msg
	}

	logger.log(message, color.New(color.Reset), scope)
}

// Warning - log a warning message
//
// params:
//   - scope string - scope of the message
//   - messages ...string - messages to log
func (logger *logger) Warning(scope string, messages ...string) {
	if logger.logLevel < 1 || !logger.inScope(scope) {
		return
	}

	message := ""
	for _, _msg := range messages {
		message = message + _msg
	}

	logger.log(message, color.New(color.FgYellow), scope)
}

// Error - log an error message
//
// params:
//   - scope string - scope of the message
//   - messages ...string - messages to log
func (logger *logger) Error(scope string, messages ...string) {
	if logger.logLevel < 0 {
		return
	}

	message := ""
	for _, _msg := range messages {
		message = message + _msg
	}

	logger.log(message, color.New(color.FgRed), scope)
}

var Logger logger = logger{
	scopes:   make([]string, 0),
	logLevel: 0,
}
