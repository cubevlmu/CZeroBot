package log

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
)

type ILogger interface {
	Info(tmp string)
	Error(tmp string)
	Warn(tmp string)
	Debug(tmp string)
	Trace(tmp string)
}

type DefaultLogger struct{}

func (d DefaultLogger) Info(tmp string) {
	log.Info(tmp)
}

func (d DefaultLogger) Error(tmp string) {
	log.Error(tmp)
}

func (d DefaultLogger) Warn(tmp string) {
	log.Warn(tmp)
}

func (d DefaultLogger) Debug(tmp string) {
	log.Debug(tmp)
}

func (d DefaultLogger) Trace(tmp string) {
	log.Trace(tmp)
}

var sharedLogger ILogger = DefaultLogger{}

func SetDefaultLogger(logger ILogger) {
	sharedLogger = logger
}

func GetDefaultLogger() *ILogger {
	if sharedLogger == nil {
		sharedLogger = DefaultLogger{}
	}
	return &sharedLogger
}

func Infof(msg string, args ...interface{}) {
	sharedLogger.Info(fmt.Sprintf(msg, args...))
}

func Warningf(msg string, args ...interface{}) {
	sharedLogger.Warn(fmt.Sprintf(msg, args...))
}

func Errorf(msg string, args ...interface{}) {
	sharedLogger.Error(fmt.Sprintf(msg, args...))
}

func Debugf(msg string, args ...interface{}) {
	sharedLogger.Debug(fmt.Sprintf(msg, args...))
}

func Tracef(msg string, args ...interface{}) {
	sharedLogger.Trace(fmt.Sprintf(msg, args...))
}

func Info(msg string) {
	sharedLogger.Info(msg)
}

func Warning(msg string) {
	sharedLogger.Warn(msg)
}

func Error(msg string) {
	sharedLogger.Error(msg)
}

func Trace(msg string) {
	sharedLogger.Trace(msg)
}

func Debug(msg string) {
	sharedLogger.Debug(msg)
}

func Fatal(err error) {
	Error(err.Error())
	os.Exit(1)
}
