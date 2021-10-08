package log

import "time"

type LogLevel int32

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
	LogLevelFatal
)

var LogLevelName = []string{
	LogLevelDebug: "DEBUG",
	LogLevelInfo:  "INFO",
	LogLevelWarn:  "WARN",
	LogLevelError: "ERROR",
	LogLevelFatal: "FATAL",
}

type Logger interface {
	SetLogLevel(LogLevel)
	LogDebug(string, ...interface{})
	LogInfo(string, ...interface{})
	LogWarn(string, ...interface{})
	LogError(string, ...interface{})
	LogFatal(string, ...interface{})
	Flush()
}

type LogSink interface {
	Sink(*LogContent)
	Flush()
}

type LogContent struct {
	logTime time.Time
	logLvl  LogLevel
	content string
}
