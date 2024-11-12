package llog

import (
	"fmt"
	"log"
	"os"
	"strings"
)

const (
	LevelDebug = iota
	LevelInfo
	LevelWarn
	LevelError
)

var (
	debugLogger = log.New(os.Stderr, "DEBUG ", log.LstdFlags|log.Lmsgprefix)
	infoLogger  = log.New(os.Stderr, "INFO  ", log.LstdFlags|log.Lmsgprefix)
	warnLogger  = log.New(os.Stderr, "WARN  ", log.LstdFlags|log.Lmsgprefix)
	errorLogger = log.New(os.Stderr, "ERROR ", log.LstdFlags|log.Lmsgprefix)
	logLevel    = LevelWarn
)

func SetLogLevel(level string) error {
	normalized := strings.ToUpper(level)
	switch normalized {
	case "DEBUG":
		logLevel = LevelDebug
		return nil
	case "INFO":
		logLevel = LevelInfo
		return nil
	case "WARN":
		logLevel = LevelWarn
		return nil
	case "ERROR":
		logLevel = LevelError
		return nil
	default:
		return fmt.Errorf("invalid log level: %s", level)
	}
}

func Debug(v ...interface{}) {
	if logLevel <= LevelDebug {
		debugLogger.Print(v...)
	}
}

func Debugf(format string, v ...interface{}) {
	if logLevel <= LevelDebug {
		debugLogger.Printf(format, v...)
	}
}

func Debugln(v ...interface{}) {
	if logLevel <= LevelDebug {
		debugLogger.Println(v...)
	}
}

func Info(v ...interface{}) {
	if logLevel <= LevelInfo {
		infoLogger.Print(v...)
	}
}

func Infof(format string, v ...interface{}) {
	if logLevel <= LevelInfo {
		infoLogger.Printf(format, v...)
	}
}

func Infoln(v ...interface{}) {
	if logLevel <= LevelInfo {
		infoLogger.Println(v...)
	}
}

func Warn(v ...interface{}) {
	if logLevel <= LevelWarn {
		warnLogger.Print(v...)
	}
}

func Warnf(format string, v ...interface{}) {
	if logLevel <= LevelWarn {
		warnLogger.Printf(format, v...)
	}
}

func Warnln(v ...interface{}) {
	if logLevel <= LevelWarn {
		warnLogger.Println(v...)
	}
}

func Error(v ...interface{}) {
	if logLevel <= LevelError {
		errorLogger.Print(v...)
	}
}

func Errorf(format string, v ...interface{}) {
	if logLevel <= LevelError {
		errorLogger.Printf(format, v...)
	}
}

func Errorln(v ...interface{}) {
	if logLevel <= LevelError {
		errorLogger.Println(v...)
	}
}
