package log

type FileLogSink struct {
}

func NewFileLogSink() *FileLogSink {
	return &FileLogSink{}
}

func (sink *FileLogSink) Sink(str string) {

}

func (sink *FileLogSink) Flush() {

}
