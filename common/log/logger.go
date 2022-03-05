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

//**************************//
// 设置一个全局的Logger供使用
//**************************//
var logger Logger = nil

func SetLogger(l Logger) {
	logger = l
}
func GetLogger() Logger {
	return logger
}

func Debug(fmtStr string, args ...interface{}) {
	logger.LogDebug(fmtStr, args...)
}
func Info(fmtStr string, args ...interface{}) {
	logger.LogInfo(fmtStr, args...)
}
func Warn(fmtStr string, args ...interface{}) {
	logger.LogWarn(fmtStr, args...)
}
func Error(fmtStr string, args ...interface{}) {
	logger.LogError(fmtStr, args...)
}
func Fatal(fmtStr string, args ...interface{}) {
	logger.LogFatal(fmtStr, args...)
}
func Flush() {
	logger.Flush()
}
