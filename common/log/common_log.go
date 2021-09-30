package log

import (
	"common/utility/queue"
	"fmt"
	"time"
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

		output := fmt.Sprintf("[%s][%s]%s", content.logTime.Format("2006-01-02 15:04:05.000"),
			LogLevelName[content.logLvl], content.content)
		for _, sink := range cl.sinks {
			sink.Sink(output)
		}
	}
}

func (cl *CommonLogger) levelLog(lvl LogLevel, fmtStr string, args ...interface{}) {
	content := &LogContent{}
	content.logLvl = lvl
	content.logTime = time.Now()
	content.content = fmt.Sprintf(fmtStr, args...)
	cl.queue.Enqueue(content)
}

func (cl *CommonLogger) LogDebug(fmtStr string, args ...interface{}) {
	if cl.logLvl > LogLevelDebug {
		return
	}
	cl.levelLog(LogLevelDebug, fmtStr, args...)
}

func (cl *CommonLogger) LogInfo(fmtStr string, args ...interface{}) {
	if cl.logLvl > LogLevelInfo {
		return
	}
	cl.levelLog(LogLevelInfo, fmtStr, args...)
}

func (cl *CommonLogger) LogWarn(fmtStr string, args ...interface{}) {
	if cl.logLvl > LogLevelWarn {
		return
	}
	cl.levelLog(LogLevelWarn, fmtStr, args...)
}

func (cl *CommonLogger) LogError(fmtStr string, args ...interface{}) {
	if cl.logLvl > LogLevelError {
		return
	}
	cl.levelLog(LogLevelError, fmtStr, args...)
}

func (cl *CommonLogger) LogFatal(fmtStr string, args ...interface{}) {
	if cl.logLvl > LogLevelFatal {
		return
	}
	cl.levelLog(LogLevelFatal, fmtStr, args...)
}

func (cl *CommonLogger) Flush() {
	for _, sink := range cl.sinks {
		sink.Flush()
	}
}
