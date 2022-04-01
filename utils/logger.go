package utils

import (
	"os"

	"github.com/sirupsen/logrus"
)

// Unless we add custom logging output, only output log levels at or
// above the global log level.

var (
	logLevel  = logrus.DebugLevel
	logPrefix = "ethereumWatcher: "
)

func init() {
	switch os.Getenv("ETHEREUM_WATCHER_LOG_LEVEL") {
	case "FATAL":
		logLevel = logrus.FatalLevel
	case "ERROR":
		logLevel = logrus.ErrorLevel
	case "WARN":
		logLevel = logrus.WarnLevel
	case "INFO":
		logLevel = logrus.InfoLevel
	case "DEBUG":
		logLevel = logrus.DebugLevel
	default:
		logLevel = logrus.TraceLevel
	}
}

func SetCategoryLogLevel(l logrus.Level) {
	logLevel = l
}

func Debugf(format string, v ...interface{}) {
	if logLevel <= logrus.DebugLevel {
		logrus.Debugf(logPrefix+format, v...)
	}
}

func Errorf(format string, v ...interface{}) {
	if logLevel <= logrus.ErrorLevel {
		logrus.Errorf(logPrefix+format, v...)
	}
}

func Infof(format string, v ...interface{}) {
	if logLevel <= logrus.InfoLevel {
		logrus.Infof(logPrefix+format, v...)
	}
}

func Tracef(format string, v ...interface{}) {
	if logLevel <= logrus.TraceLevel {
		logrus.Tracef(logPrefix+format, v...)
	}
}

func Warnf(format string, v ...interface{}) {
	if logLevel <= logrus.WarnLevel {
		logrus.Warnf(logPrefix+format, v...)
	}
}
