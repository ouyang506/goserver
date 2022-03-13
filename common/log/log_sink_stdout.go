package log

import "fmt"

/// std out log sink
type StdoutLogSink struct {
}

func NewStdLogSink() *StdoutLogSink {
	return &StdoutLogSink{}
}

func (sink *StdoutLogSink) Sink(content *LogContent) {
	output := fmt.Sprintf("[%s][%s][%s]%s", content.logTime.Format("2006-01-02 15:04:05.000"),
		LogLevelName[content.logLvl], content.fileName, content.content)
	fmt.Println(output)
}

func (sink *StdoutLogSink) Flush() {
	//pass
}
