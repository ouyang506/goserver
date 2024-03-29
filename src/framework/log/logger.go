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
	LogDebug(int, string, ...interface{})
	LogInfo(int, string, ...interface{})
	LogWarn(int, string, ...interface{})
	LogError(int, string, ...interface{})
	LogFatal(int, string, ...interface{})
	Flush()
}

type LogSink interface {
	Sink(*LogContent)
	Flush()
}

type LogContent struct {
	logTime  time.Time
	logLvl   LogLevel
	fileName string
	content  string
}

// **************************//
// 设置一个全局的Logger供使用
// **************************//
var logger Logger = nil

func SetLogger(l Logger) {
	logger = l
}
func GetLogger() Logger {
	return logger
}

func Debug(fmtStr string, args ...interface{}) {
	if logger == nil {
		return
	}
	logger.LogDebug(1, fmtStr, args...)
}
func Info(fmtStr string, args ...interface{}) {
	if logger == nil {
		return
	}
	logger.LogInfo(1, fmtStr, args...)
}
func Warn(fmtStr string, args ...interface{}) {
	if logger == nil {
		return
	}
	logger.LogWarn(1, fmtStr, args...)
}
func Error(fmtStr string, args ...interface{}) {
	if logger == nil {
		return
	}
	logger.LogError(1, fmtStr, args...)
}
func Fatal(fmtStr string, args ...interface{}) {
	if logger == nil {
		return
	}
	logger.LogFatal(1, fmtStr, args...)
}
func Flush() {
	if logger == nil {
		return
	}
	logger.Flush()
}
