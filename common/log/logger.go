package log

import (
	"fmt"
	"time"
)

type Logger interface {
	LogDebug(string, ...interface{})
	LogInfo(string, ...interface{})
	LogWarn(string, ...interface{})
	LogError(string, ...interface{})
	LogFatal(string, ...interface{})
}

type LoggerSink interface {
	Log(string)
}

type DefaultLogger struct {
	sinks []LoggerSink
}

func NewDefaultLogger() *DefaultLogger {
	return &DefaultLogger{}
}

func (logger *DefaultLogger) AddSink(sink LoggerSink) {
	logger.sinks = append(logger.sinks, sink)
}

func (logger *DefaultLogger) makePrefix(flag string) string {
	return fmt.Sprintf("[%s][%s]", time.Now().Format("2006-01-02 15:04:05.000"), flag)
}

func (logger *DefaultLogger) LogDebug(fmtStr string, args ...interface{}) {
	output := fmt.Sprintf(logger.makePrefix("DEBUG")+fmtStr, args...)
	for _, sink := range logger.sinks {
		sink.Log(output)
	}
}

func (logger *DefaultLogger) LogInfo(fmtStr string, args ...interface{}) {
	output := fmt.Sprintf(logger.makePrefix("INFO")+fmtStr, args...)
	for _, sink := range logger.sinks {
		sink.Log(output)
	}
}

func (logger *DefaultLogger) LogWarn(fmtStr string, args ...interface{}) {
	output := fmt.Sprintf(logger.makePrefix("WARN")+fmtStr, args...)
	for _, sink := range logger.sinks {
		sink.Log(output)
	}
}

func (logger *DefaultLogger) LogError(fmtStr string, args ...interface{}) {
	output := fmt.Sprintf(logger.makePrefix("ERROR")+fmtStr, args...)
	for _, sink := range logger.sinks {
		sink.Log(output)
	}
}

func (logger *DefaultLogger) LogFatal(fmtStr string, args ...interface{}) {
	output := fmt.Sprintf(logger.makePrefix("FATAL")+fmtStr, args...)
	for _, sink := range logger.sinks {
		sink.Log(output)
	}
}

/// std out log sink
type StdLogSink struct {
}

func NewStdLogSink() *StdLogSink {
	return &StdLogSink{}
}

func (sink *StdLogSink) Log(str string) {
	fmt.Println(str)
}
