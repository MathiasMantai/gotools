package log

import (
	"archive/zip"
	"errors"
	"fmt"
	// "io/fs"
	"os"
	"path/filepath"
	"strings"
	// "time"
	"io"
	"strconv"

)

type FileLogger struct {
	TimeZone   string
	TimeLayout string
	BeforeTime string
	AfterTime  string
	KeepOldest int
	DirPath    string
	LogFilePrefix  string
	LogMessagePrefix string
	UseTimestamp bool
	LatestLogFileName string
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

func (l *FileLogger) SetKeepOldest(value int) {
	l.KeepOldest = value
}


/* Getter */

/* Log file Rotation */
func (l *FileLogger) LogRotate() error {
	files, readDirError := l.GetLogFilesNamesInDir(l.DirPath)

	if readDirError != nil {
		panic(readDirError)
	}

	logCnt := len(files)

	if logCnt > int(l.KeepOldest) {
		archiveName, getArchiveNameError := l.GetLatestArchiveFileName(l.DirPath)
		if getArchiveNameError!= nil {
            return fmt.Errorf("x> error while getting archive name: %v", getArchiveNameError)
        }

		if strings.TrimSpace(archiveName) == "" {
			return errors.New("x> archive name is empty")
		}

		archive, err:=os.Create(archiveName)
		if err!=nil{
			return err
		}

		defer archive.Close()

		for _, file := range files {
			fmt.Println(file)
			logFilePath := filepath.Join(l.DirPath, file)
			zipWriter := zip.NewWriter(archive)
			defer zipWriter.Close()

			logFile, logFileOpenError := os.Open(logFilePath)
			if logFileOpenError!= nil {
                return logFileOpenError
            }
			defer logFile.Close()

			writer, createWriterError := zipWriter.Create(file)
			if createWriterError!= nil {
                return createWriterError
            }
			
			_, copyError := io.Copy(writer, logFile)
			if copyError!= nil {
                return copyError
            }

			//remove file
			deleteFileError := os.Remove(logFilePath)
			if deleteFileError != nil {
				return deleteFileError
			}
		}
	}

	return nil
}

func (l *FileLogger) GetLogFilesNamesInDir(dir string) ([]string, error) {
	files, readDirError := os.ReadDir(dir)

    if readDirError!= nil {
        return nil, readDirError
    }

	var rs []string

	for _, file := range files {

		if strings.Contains(file.Name(), ".log") {
			rs = append(rs, file.Name())
		}
	}

    return rs, nil
}

func (l *FileLogger) GetArchiveCount(dir string) (int, error) {
	files, readDirError := os.ReadDir(dir)

    if readDirError!= nil {
        return 0, readDirError
    }

	var rs []string

	for _, file := range files {
        if strings.Contains(file.Name(), ".zip") {
            rs = append(rs, file.Name())
        }
    }

    return len(files), nil
}

func (l *FileLogger) GetLatestArchiveFileName(dir string) (string, error) {
	archiveCnt, archiveCountError := l.GetArchiveCount(dir)
	fmt.Println(strconv.Itoa(archiveCnt))
	if archiveCountError!= nil {
        return "", archiveCountError
    }

	return fmt.Sprintf("archive_%v.zip", archiveCnt + 1), nil
}

func (l *FileLogger) getRootDirError() error {
	if l.DirPath == "" {
		return errors.New("x> ")
	}

	return nil
}

/* Init */

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