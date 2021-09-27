package log

import "fmt"

/// std out log sink
type StdoutLogSink struct {
}

func NewStdLogSink() *StdoutLogSink {
	return &StdoutLogSink{}
}

func (sink *StdoutLogSink) Sink(str string) {
	fmt.Println(str)
}

func (sink *StdoutLogSink) Flush() {

}
