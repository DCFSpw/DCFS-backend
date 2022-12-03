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

func (logger *logger) GetLogLevel() int {
	return logger.logLevel
}

func (logger *logger) SetLogLevel(level int) {
	logger.logLevel = level
}

func (logger *logger) SetScopes(_scopes []string) {
	if logger.scopes == nil {
		logger.scopes = make([]string, 0)
	}

	logger.scopes = append(logger.scopes, _scopes...)
}

func (logger *logger) inScope(_scope string) bool {
	for _, scope := range logger.scopes {
		if scope == _scope || scope == "all" {
			return true
		}
	}

	return false
}

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

	//fmt.Print(color)
	//fmt.Printf("[%s][%d] %s%s%s\n", scope, log.Llongfile, color, msg, colorReset)
	// fmt.Print(colorReset)
}

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
