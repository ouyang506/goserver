package log

import (
	"fmt"
	"os"
	"time"
)

type RotateType int

const (
	RotateByDay RotateType = iota
	RotateByHour
)

type FileLogSink struct {
	prefixFilename string
	logDir      string
	rotateType  RotateType
	curFile     *os.File
	curFileName string
}

func NewFileLogSink(prefixFilename string, logDir string, rotateType RotateType) *FileLogSink {
	if logDir == "" {
		logDir = "./log/"
	}
	_, err := os.Stat(logDir)
	if os.IsNotExist(err) {
		os.Mkdir(logDir, os.FileMode(0770))
	}

	sink := &FileLogSink{
		prefixFilename: prefixFilename,
		logDir:     logDir,
		rotateType: rotateType,
	}
	return sink
}
func (sink *FileLogSink) getFileName(t time.Time) string {
	switch sink.rotateType {
	case RotateByDay:
		return fmt.Sprintf("%s_%s.log", sink.prefixFilename,  t.Format("2006_01_02"))
	case RotateByHour:
		return fmt.Sprintf("%s_%s.log", sink.prefixFilename,  t.Format("2006_01_02_15")) 
	default:
		return fmt.Sprintf("%s_%s.log", sink.prefixFilename,  t.Format("2006_01_02"))
	}
}
func (sink *FileLogSink) openFile(fileName string) (*os.File, error) {
	f, err := os.OpenFile(sink.logDir+fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, os.FileMode(0660))
	return f, err
}

func (sink *FileLogSink) Sink(content *LogContent) {
	t := content.logTime
	fileName := sink.getFileName(t)
	if sink.curFileName != fileName {
		if sink.curFile != nil {
			sink.curFile.Close()
		}
		sink.curFileName = fileName
		sink.curFile, _ = sink.openFile(fileName)
	}

	if sink.curFile == nil {
		return
	}
	output := fmt.Sprintf("[%s][%s][%s]%s\n", content.logTime.Format("2006-01-02 15:04:05.000"),
		LogLevelName[content.logLvl], content.fileName, content.content)
	sink.curFile.WriteString(output)
}

func (sink *FileLogSink) Flush() {
	sink.curFile.Sync()
}
