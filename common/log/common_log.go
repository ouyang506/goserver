package log

type CommonLogger struct {
	logLvl LogLevel
	sinks  []LogSink
}

func NewCommonLogger() *CommonLogger {
	cl := &CommonLogger{}
	return cl
}

func (cl *CommonLogger) SetLogLevel(logLvl LogLevel) {
	cl.logLvl = logLvl
}

func (cl *CommonLogger) AddSink(sink LogSink) {
	cl.sinks = append(logger.sinks, sink)
}

func (cl *CommonLogger) levelLog(lvl LogLevel, fmtStr string, args ...interface{}) {
	// output := fmt.Sprintf("[%s][%s]%s", time.Now().Format("2006-01-02 15:04:05.000"), LogLevelName[lvl], fmtStr, args...)
	// for _, sink := range l.sinks {
	// 	sink.Sink(output)
	// }
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
