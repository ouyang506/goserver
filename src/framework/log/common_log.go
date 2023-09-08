package log

import (
	"fmt"
	"path/filepath"
	"runtime"
	"time"
	"utility/queue"
)

type CommonLogger struct {
	logLvl LogLevel
	sinks  []LogSink
	queue  *queue.LockFreeQueue
}

func NewCommonLogger() *CommonLogger {
	cl := &CommonLogger{}
	cl.queue = queue.NewLockFreeQueue()
	return cl
}

func (cl *CommonLogger) SetLogLevel(logLvl LogLevel) {
	cl.logLvl = logLvl
}

func (cl *CommonLogger) AddSink(sink LogSink) {
	cl.sinks = append(cl.sinks, sink)
}

func (cl *CommonLogger) Start() {
	go cl.loopSink()
}

func (cl *CommonLogger) loopSink() {
	for {
		v := cl.queue.Dequeue()
		if v == nil {
			time.Sleep(500 * time.Millisecond)
			continue
		}
		content := v.(*LogContent)
		for _, sink := range cl.sinks {
			sink.Sink(content)
		}
	}
}

func (cl *CommonLogger) levelLog(depth int, lvl LogLevel, fmtStr string, args ...interface{}) {
	if cl.logLvl > lvl {
		return
	}
	content := &LogContent{}
	content.logLvl = lvl
	content.logTime = time.Now()

	_, fullname, line, ok := runtime.Caller(depth + 1)
	if !ok {
		content.fileName = "???.go:0"
	} else {
		_, fileName := filepath.Split(fullname)
		content.fileName = fmt.Sprintf("%s:%d", fileName, line)
	}

	content.content = fmt.Sprintf(fmtStr, args...)
	cl.queue.Enqueue(content)
}

func (cl *CommonLogger) LogDebug(depth int, fmtStr string, args ...interface{}) {
	cl.levelLog(depth+1, LogLevelDebug, fmtStr, args...)
}

func (cl *CommonLogger) LogInfo(depth int, fmtStr string, args ...interface{}) {
	cl.levelLog(depth+1, LogLevelInfo, fmtStr, args...)
}

func (cl *CommonLogger) LogWarn(depth int, fmtStr string, args ...interface{}) {
	cl.levelLog(depth+1, LogLevelWarn, fmtStr, args...)
}

func (cl *CommonLogger) LogError(depth int, fmtStr string, args ...interface{}) {
	cl.levelLog(depth+1, LogLevelError, fmtStr, args...)
}

func (cl *CommonLogger) LogFatal(depth int, fmtStr string, args ...interface{}) {
	cl.levelLog(depth+1, LogLevelFatal, fmtStr, args...)
}

func (cl *CommonLogger) Flush() {
	for _, sink := range cl.sinks {
		sink.Flush()
	}
}
