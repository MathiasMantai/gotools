package log

import (
	// "archive/zip"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
)

type FileLogger struct {
	TimeZone   string
	TimeLayout string
	BeforeTime string
	AfterTime  string
	KeepOldest int32
	DirPath    string
	LogPrefix  string
}

func (l *FileLogger) LogMsg(msg string) bool {
	encodedMsg := []byte(msg)
	writeError := os.WriteFile(l.DirPath, encodedMsg, 0644)

	return writeError == nil
}

func (l *FileLogger) LogRotate() bool {
	files, readDirError := os.ReadDir(l.DirPath)

	if readDirError != nil {
		panic(readDirError)
	}

	for _, file := range files {
		fmt.Println(file.Name())
		logName := file.Name()
		logTime := strings.Replace(logName, l.LogPrefix, "", -1)
		tmpTime, parseTimeError := time.Parse("01.12.2006 13:25", logTime)

		if parseTimeError != nil {
			panic(parseTimeError)
		}

		since := time.Since(tmpTime)
		sinceInDays := (since.Hours() / 24)

		if sinceInDays >= float64(l.KeepOldest) {
			//zip log
		}
	}

	return true
}

func (l *FileLogger) InitLogDir() error {

	if l.DirPath == "" {
		return errors.New("logging directory has not been specified")
	}

	createDirError := os.Mkdir(l.DirPath, 0755)
	if createDirError != nil {
		fmt.Println("=> Directory exists already. Skipping...")
	} else {
		fmt.Println("=> Log Directory created...")
	}

	return nil
}

func NewFileLogger() *FileLogger {
	var logger FileLogger
	logger.LogPrefix = "log_"
	return &logger
}
