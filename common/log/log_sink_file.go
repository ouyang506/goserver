package log

import (
	"fmt"
	"os"
	"time"
)

type FileLogSink struct {
	logDir      string
	curFile     *os.File
	curFileName string
}

func NewFileLogSink(logDir string) *FileLogSink {
	if logDir == "" {
		logDir = "./log/"
	}

	_, err := os.Stat(logDir)
	if os.IsNotExist(err) {
		os.Mkdir(logDir, os.ModePerm)
	}

	sink := &FileLogSink{
		logDir: logDir,
	}
	return sink
}
func (sink *FileLogSink) getFileName(t time.Time) string {
	return t.Format("2006_01_02") + ".log"
}
func (sink *FileLogSink) getFile(fileName string) (*os.File, error) {
	f, err := os.OpenFile(sink.logDir+fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 660)
	return f, err
}

func (sink *FileLogSink) Sink(content *LogContent) {
	t := content.logTime
	fileName := sink.getFileName(t)
	if sink.curFile == nil || sink.curFileName == fileName {
		sink.curFileName = fileName
		sink.curFile, _ = sink.getFile(fileName)
	}

	if sink.curFile == nil {
		return
	}
	output := fmt.Sprintf("[%s][%s]%s\n", content.logTime.Format("2006-01-02 15:04:05.000"),
		LogLevelName[content.logLvl], content.content)
	sink.curFile.WriteString(output)
}

func (sink *FileLogSink) Flush() {
	sink.curFile.Sync()
}
