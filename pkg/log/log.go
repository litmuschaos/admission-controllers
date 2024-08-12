package log

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

type Log struct {
	*logrus.Logger
}

var Logger Log

func InitLogger(formatter, level string) {
	logr := logrus.New()
	setFormatter(logr, formatter)
	setLoggingLevel(logr, level)
	Logger = Log{logr}
}

func setFormatter(logger *logrus.Logger, formatter string) {
	switch strings.ToLower(formatter) {
	case "json":
		logger.SetFormatter(&CustomJSONFormatter{})
	default:
		logger.SetFormatter(&logrus.TextFormatter{})
	}
}

func setLoggingLevel(logger *logrus.Logger, level string) {
	switch strings.ToLower(level) {
	case "trace":
		logger.SetLevel(logrus.TraceLevel)
	case "debug":
		logger.SetLevel(logrus.DebugLevel)
	case "warn":
		logger.SetLevel(logrus.WarnLevel)
	case "error":
		logger.SetLevel(logrus.ErrorLevel)
	case "fatal":
		logger.SetLevel(logrus.FatalLevel)
	case "panic":
		logger.SetLevel(logrus.PanicLevel)
	default:
		logger.SetLevel(logrus.InfoLevel)
	}
}

// Fatalf Logs first and then calls `logger.Exit(1)`
// logging level is set to Panic.
func (l *Log) Fatalf(msg string, err ...interface{}) {
	l.WithFields(logrus.Fields{}).Fatalf(msg, err...)
}

// Fatal Logs first and then calls `logger.Exit(1)`
// logging level is set to Panic.
func (l *Log) Fatal(msg string) {
	l.WithFields(logrus.Fields{}).Fatal(msg)
}

// Infof log the General operational entries about what's going on inside the application
func (l *Log) Infof(msg string, val ...interface{}) {
	l.WithFields(logrus.Fields{}).Infof(msg, val...)
}

// Info log the General operational entries about what's going on inside the application
func (l *Log) Info(msg string) {
	l.WithFields(logrus.Fields{}).Infof(msg)
}

// InfoWithValues log the General operational entries about what's going on inside the application
// It also print the extra key values pairs
func (l *Log) InfoWithValues(msg string, val map[string]interface{}) {
	l.WithFields(val).Info(msg)
}

// ErrorWithValues log the Error entries happening inside the code
// It also print the extra key values pairs
func (l *Log) ErrorWithValues(msg string, val map[string]interface{}) {
	l.WithFields(val).Error(msg)
}

// Warn log the Non-critical entries that deserve eyes.
func (l *Log) Warn(msg string) {
	l.WithFields(logrus.Fields{}).Warn(msg)
}

// Warnf log the Non-critical entries that deserve eyes.
func (l *Log) Warnf(msg string, val ...interface{}) {
	l.WithFields(logrus.Fields{}).Warnf(msg, val...)
}

// Errorf used for errors that should definitely be noted.
// Commonly used for hooks to send errors to an error tracking service.
func (l *Log) Errorf(msg string, err ...interface{}) {
	l.WithFields(logrus.Fields{}).Errorf(msg, err...)
}

// Error used for errors that should definitely be noted.
// Commonly used for hooks to send errors to an error tracking service
func (l *Log) Error(msg string) {
	l.WithFields(logrus.Fields{}).Error(msg)
}

// Debugf Usually only enabled when debugging. Very verbose logging
func (l *Log) Debugf(msg string, val ...interface{}) {
	l.WithFields(logrus.Fields{}).Debugf(msg, val...)
}

type CustomJSONFormatter struct {
	FieldMap map[string]string
}

func (f *CustomJSONFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	data := make(logrus.Fields, len(entry.Data)+4)
	args := map[string]string{}

	for key, value := range entry.Data {
		if key != "time" && key != "level" && key != "out" {
			bytes, err := json.Marshal(value)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal fields to JSON, %w", err)
			}
			args[key] = string(bytes)
		} else {
			data[key] = value
		}
	}

	data["time"] = entry.Time.Format(time.RFC3339)
	data["level"] = entry.Level.String()
	data["out"] = entry.Message
	if len(args) != 0 {
		data["args"] = args
	}

	var b *bytes.Buffer
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}

	encoder := json.NewEncoder(b)
	if err := encoder.Encode(data); err != nil {
		return nil, fmt.Errorf("failed to marshal fields to JSON, %w", err)
	}

	return b.Bytes(), nil
}
