package tlog

import (
	"os"

	"github.com/sirupsen/logrus"
)

const (
	// ProgramName is used in log reports.
	ProgramName = "wizefs"
	wpanicMsg   = "-wpanic turns this warning into a panic: "
)

var logger *logrus.Logger

func SetLevel(level logrus.Level) {
	logger.SetLevel(level)
}

func SetDebugLevel() {
	logger.SetLevel(logrus.DebugLevel)
}

func Warn(args ...interface{}) {
	if logger != nil {
		logger.Warn(args)
	}
}

func Warnf(format string, args ...interface{}) {
	if logger != nil {
		logger.Warnf(format, args)
	}
}

func Debug(args ...interface{}) {
	if logger != nil {
		logger.Debug(args)
	}
}

func Debugf(format string, args ...interface{}) {
	if logger != nil {
		logger.Debugf(format, args)
	}
}

func Info(args ...interface{}) {
	if logger != nil {
		logger.Info(args)
	}
}

func Infof(format string, args ...interface{}) {
	if logger != nil {
		logger.Infof(format, args)
	}
}

func Fatal(args ...interface{}) {
	if logger != nil {
		logger.Fatal(args)
	}
}

func Fatalf(format string, args ...interface{}) {
	if logger != nil {
		logger.Fatalf(format, args)
	}
}

func WithFields(fields logrus.Fields) *logrus.Entry {
	if logger != nil {
		return logger.WithFields(fields)
	}
	return nil
}

func init() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetOutput(os.Stdout)

	logger = logrus.New()

	logger.SetLevel(logrus.InfoLevel)
}

// SwitchToSyslog redirects the output of this logger to syslog.
//func (l *toggledLogger) SwitchToSyslog(p syslog.Priority) {
//	w, err := syslog.New(p, ProgramName)
//	if err != nil {
//		Warn.Printf("SwitchToSyslog: %v", err)
//	} else {
//		l.SetOutput(w)
//	}
//}

// SwitchLoggerToSyslog redirects the default log.Logger that the go-fuse lib uses
// to syslog.
//func SwitchLoggerToSyslog(p syslog.Priority) {
//	w, err := syslog.New(p, ProgramName)
//	if err != nil {
//		Warn.Printf("SwitchLoggerToSyslog: %v", err)
//	} else {
//		log.SetPrefix("go-fuse: ")
//		// Disable printing the timestamp, syslog already provides that
//		log.SetFlags(0)
//		log.SetOutput(w)
//	}
//}
