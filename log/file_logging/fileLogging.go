package log

import (
	// "archive/zip"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
	"path/filepath"
)

type FileLogger struct {
	TimeZone   string
	TimeLayout string
	BeforeTime string
	AfterTime  string
	KeepOldest int32
	DirPath    string
	LogFilePrefix  string
	LogMessagePrefix string
	UseTimestamp bool
}

/* Setter */

func (l *FileLogger) SetTimeZone(value string) {
	l.TimeZone = value
}

func (l *FileLogger) SetTimeLayout(value string) {
	l.TimeLayout = value
}

func (l *FileLogger) SetUseTimestamp(value bool) {
	l.UseTimestamp = value
}


/* Getter */

func (l *FileLogger) LogRotate() bool {
	files, readDirError := os.ReadDir(l.DirPath)

	if readDirError != nil {
		panic(readDirError)
	}

	for _, file := range files {
		fmt.Println(file.Name())
		logName := file.Name()
		logTime := strings.Replace(logName, l.LogFilePrefix, "", -1)
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
		return errors.New("x> Log directory path has not been specified")
	}

	if _, fileExistsError := os.Stat(l.DirPath); !os.IsNotExist(fileExistsError) {
		return errors.New("x> Log Directory already exists")
	}

	createDirError := os.Mkdir(l.DirPath, 0755)
	if createDirError != nil {
		return errors.New("x> Log Directory could not be created")
	}

	return nil
}

func (l *FileLogger) InitLogFile() error {
	logFileName, logFileNameError := GetLogFileName(l.LogFilePrefix, l.TimeLayout, l.TimeZone)
	
	if logFileNameError != nil {
		return fmt.Errorf("x> Log file name could not be generated: %s", logFileNameError.Error())
	}

	if strings.TrimSpace(logFileName) == "" {
		return errors.New("x> Log file name could not be generated")
	}

	if _, fileExistsError := os.Stat(l.DirPath); os.IsNotExist(fileExistsError) {
		return errors.New("x> Log Directory does not exist")
	}

	if _, fileExistsError := os.Stat(filepath.Join(l.DirPath, logFileName)); !os.IsNotExist(fileExistsError) {
        return errors.New("x> Log file already exists")
    }

	if _, createFileError := os.Create(filepath.Join(l.DirPath, logFileName)); createFileError != nil {
		return fmt.Errorf("x> Log file could not be created: %v", createFileError)
	}

	return nil
}

func (l *FileLogger) LogMsg(msg string) error {
	formattedMsg, formatError := l.FormatFileLogMsg(msg)
	if formatError!= nil {
        return fmt.Errorf("x> error formatting message: %v", formatError)
    }
	encodedMsg := []byte(formattedMsg)
	logFileName, logFileNameError := GetLogFileName(l.LogFilePrefix, l.TimeLayout, l.TimeZone)
	if logFileNameError != nil {
		return fmt.Errorf("x> undefined log file name")
	}

	logFilePath := filepath.Join(l.DirPath, logFileName)
	
	file, openFileError := os.OpenFile(logFilePath, os.O_APPEND, 0644)
	if openFileError!= nil {
        return fmt.Errorf("x> Error opening log file: %s", openFileError.Error())
    }
	defer file.Close()

	_, writeFileError := file.Write(encodedMsg)
	if writeFileError != nil {
		return fmt.Errorf("x> Error writing to log file: %s", writeFileError.Error())
	}

	return nil
}

func NewFileLogger() *FileLogger {
	var logger FileLogger
	logger.UseTimestamp = false
	logger.LogFilePrefix = "log_"
	return &logger
}


/* Util */
func (l *FileLogger) FormatFileLogMsg(msg string) (string, error) {
	
	if strings.TrimSpace(l.TimeLayout) != "" && strings.TrimSpace(l.TimeZone) != "" {
		timestamp, timestampError := GetCurrentTime(l.TimeZone, l.TimeLayout)

		if timestampError != nil {
			return "", timestampError
		}

		//add a timestamp
		if l.UseTimestamp  {
			msg = fmt.Sprintf("[%s] - %s", timestamp, msg)
		}
	} else {
		return "", fmt.Errorf("x> TimeZone or TimeLayout have not been specified correctly")
	}

	//add a line break
	msg = fmt.Sprintf("%s\n", msg)

	return msg, nil
}