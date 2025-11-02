package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

// Init initializes the logger with the specified environment
func Init(env string) {
	log = logrus.New()

	// Set logger format based on environment
	if env == "production" {
		log.SetFormatter(&logrus.JSONFormatter{})
		log.SetLevel(logrus.InfoLevel)
	} else {
		log.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
		})
		log.SetLevel(logrus.DebugLevel)
	}

	// Set output to stdout
	log.SetOutput(os.Stdout)
}

// Debug logs a debug message
func Debug(args ...interface{}) {
	if log != nil {
		log.Debug(args...)
	}
}

// Debugf logs a debug message with formatting
func Debugf(format string, args ...interface{}) {
	if log != nil {
		log.Debugf(format, args...)
	}
}

// Info logs an info message
func Info(args ...interface{}) {
	if log != nil {
		log.Info(args...)
	}
}

// Infof logs an info message with formatting
func Infof(format string, args ...interface{}) {
	if log != nil {
		log.Infof(format, args...)
	}
}

// Warn logs a warning message
func Warn(args ...interface{}) {
	if log != nil {
		log.Warn(args...)
	}
}

// Warnf logs a warning message with formatting
func Warnf(format string, args ...interface{}) {
	if log != nil {
		log.Warnf(format, args...)
	}
}

// Error logs an error message
func Error(args ...interface{}) {
	if log != nil {
		log.Error(args...)
	}
}

// Errorf logs an error message with formatting
func Errorf(format string, args ...interface{}) {
	if log != nil {
		log.Errorf(format, args...)
	}
}

// Fatal logs a fatal message and exits
func Fatal(args ...interface{}) {
	if log != nil {
		log.Fatal(args...)
	}
}

// Fatalf logs a fatal message with formatting and exits
func Fatalf(format string, args ...interface{}) {
	if log != nil {
		log.Fatalf(format, args...)
	}
}

// WithField returns a logger entry with a single field
func WithField(key string, value interface{}) *logrus.Entry {
	if log != nil {
		return log.WithField(key, value)
	}
	return nil
}

// WithFields returns a logger entry with multiple fields
func WithFields(fields logrus.Fields) *logrus.Entry {
	if log != nil {
		return log.WithFields(fields)
	}
	return nil
}

// GetLogger returns the underlying logger instance
func GetLogger() *logrus.Logger {
	return log
}