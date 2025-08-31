package log

import (
	"fmt"
	standardLog "log"
	"os"
)

var (
	logger Logger = DefaultLogger()
)

type Level int

const (
	LevelDebug Level = -4
	LevelInfo  Level = 0
	LevelWarn  Level = 4
	LevelError Level = 8
)

type Logger interface {
	Log(level Level, msg string)
}

type LoggerFunc func(level Level, msg string)

func (f LoggerFunc) Log(level Level, msg string) {
	f(level, msg)
}

func Debug(msg string) {
	logger.Log(LevelDebug, msg)
}

func Info(msg string) {
	logger.Log(LevelInfo, msg)
}

func Warn(msg string) {
	logger.Log(LevelWarn, msg)
}

func Error(msg string) {
	logger.Log(LevelError, msg)
}

func DefaultLogger() Logger {
	l := standardLog.New(os.Stderr, "", standardLog.Ldate|standardLog.Ltime)

	return LoggerFunc(func(level Level, msg string) {
		switch level {
		case LevelDebug:
			msg = fmt.Sprintf("DEBUG :%v", msg)
		case LevelInfo:
			msg = fmt.Sprintf("INFO :%v", msg)
		case LevelWarn:
			msg = fmt.Sprintf("WARN :%v", msg)
		case LevelError:
			msg = fmt.Sprintf("ERROR :%v", msg)
		default:
			msg = fmt.Sprintf("UNKNOWN(%d): %v", level, msg)
		}

		l.Println(msg)
	})
}
