package utils

import (
	log "github.com/sirupsen/logrus"
	"os"
)

func init() {
	switch os.Getenv("HSK_LOG_LEVEL") {
	case "FATAL":
		log.SetLevel(log.FatalLevel)
	case "ERROR":
		log.SetLevel(log.ErrorLevel)
	case "WARN":
		log.SetLevel(log.WarnLevel)
	case "INFO":
		log.SetLevel(log.InfoLevel)
	case "DEBUG":
		log.SetLevel(log.DebugLevel)
	default:
		log.SetLevel(log.InfoLevel)
	}

	formatter := &log.TextFormatter{
		FullTimestamp: true,
	}

	log.SetFormatter(formatter)
}

func Debugf(format string, v ...interface{}) {
	log.Debugf(format, v...)
}

func Infof(format string, v ...interface{}) {
	log.Infof(format, v...)
}

func Errorf(format string, v ...interface{}) {
	log.Errorf(format, v...)
}
