package log

import (
	"fmt"
	"time"
)

type LoggerInterface interface {
	LogDebug(string, ...interface{})
	LogInfo(string, ...interface{})
	LogError(string, ...interface{})
	LogFatal(string, ...interface{})
}

type Logger struct {
	sinks []LoggerInterface
}

func NewLogger() *Logger {
	return &Logger{}
}

func (logger *Logger) AddSink(sink LoggerInterface) {
	logger.sinks = append(logger.sinks, sink)
}

func (logger *Logger) LogDebug(fmtStr string, args ...interface{}) {
	for _, sink := range logger.sinks {
		sink.LogDebug(fmtStr, args...)
	}
}

func (logger *Logger) LogInfo(fmtStr string, args ...interface{}) {
	for _, sink := range logger.sinks {
		sink.LogInfo(fmtStr, args...)
	}
}

func (logger *Logger) LogError(fmtStr string, args ...interface{}) {
	for _, sink := range logger.sinks {
		sink.LogError(fmtStr, args...)
	}
}

func (logger *Logger) LogFatal(fmtStr string, args ...interface{}) {
	for _, sink := range logger.sinks {
		sink.LogFatal(fmtStr, args...)
	}
}

/// std out log sink
type StdLogSink struct {
}

func NewStdLogSink() *StdLogSink {
	return &StdLogSink{}
}

func (sink *StdLogSink) makePrefix(flag string) string {
	return fmt.Sprintf("[%s][%s]", time.Now().Format("2006-01-02 15:04:05.000"), flag)
}

func (sink *StdLogSink) LogDebug(fmtStr string, args ...interface{}) {
	str := fmt.Sprintf(sink.makePrefix("DEBUG")+fmtStr, args...)
	fmt.Println(str)
}

func (sink *StdLogSink) LogInfo(fmtStr string, args ...interface{}) {
	str := fmt.Sprintf(sink.makePrefix("INFO")+fmtStr, args...)
	fmt.Println(str)
}

func (sink *StdLogSink) LogError(fmtStr string, args ...interface{}) {
	str := fmt.Sprintf(sink.makePrefix("ERROR")+fmtStr, args...)
	fmt.Println(str)
}

func (sink *StdLogSink) LogFatal(fmtStr string, args ...interface{}) {
	str := fmt.Sprintf(sink.makePrefix("FATAL")+fmtStr, args...)
	fmt.Println(str)
}
