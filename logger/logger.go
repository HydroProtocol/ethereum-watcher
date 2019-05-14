package logger

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
)

type JSONFormatter struct {
	// TimestampFormat sets the format used for marshaling timestamps.
	TimestampFormat string
}

func (f *JSONFormatter) Format(entry *log.Entry) ([]byte, error) {
	data := make(map[string]interface{})
	for k, v := range entry.Data {
		switch v := v.(type) {
		case error:
			// Otherwise errors are ignored by `encoding/json`
			// https://github.com/Sirupsen/logrus/issues/137
			data[k] = v.Error()
		default:
			data[k] = v
		}
	}

	_, ok := data["time"]
	if ok {
		data["fields.time"] = data["time"]
	}

	_, ok = data["msg"]
	if ok {
		data["fields.msg"] = data["msg"]
	}

	_, ok = data["level"]
	if ok {
		data["fields.level"] = data["level"]
	}

	timestampFormat := f.TimestampFormat
	if timestampFormat == "" {
		timestampFormat = log.DefaultTimestampFormat
	}

	data["time"] = entry.Time.Format(timestampFormat)
	data["msg"] = entry.Message
	data["level"] = entry.Level.String()

	logData := make(map[string]interface{})
	logData["logName"] = "default"
	logData["logPayload"] = data

	serialized, err := json.Marshal(logData)

	if err != nil {
		return nil, fmt.Errorf("Failed to marshal fields to JSON, %v", err)
	}

	return append(serialized, '\n'), nil
}

var ReloadLogger *log.Entry
var CommonTransactionCancelLogger *log.Entry
var BlockWatcherLogger *log.Entry

func init() {
	if os.Getenv("DEBUG") == "true" {
		log.SetLevel(log.DebugLevel)
	}

	if os.Getenv("DDEX_NAMESPACE") == "production" || os.Getenv("DDEX_NAMESPACE") == "ropsten" {
		log.SetFormatter(&JSONFormatter{})
	} else {
		log.SetFormatter(&log.TextFormatter{})
	}

	log.SetOutput(os.Stdout)

	ReloadLogger = log.WithFields(log.Fields{
		"component": "[RELOAD]",
	})

	CommonTransactionCancelLogger = log.WithFields(log.Fields{
		"component": "[CANCELER]",
	})

	BlockWatcherLogger = log.WithFields(log.Fields{
		"component": "[BLOCK]",
	})
}
