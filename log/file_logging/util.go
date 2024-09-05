package log

import (
	"fmt"
	"time"
)

func GetCurrentTime(timeZone string, layout string) (string, error) {

	location, err := time.LoadLocation(timeZone)
	if err != nil {
		return "", err
	}

	curTime := time.Now().In(location).Format(layout)
	return curTime, nil
}

func GetLogFileName(fileNamePrefix string, TimeLayout string, TimeZone string) (string, error) {
	curTime, curTimeError := GetCurrentTime(TimeZone, "2006-01-02")
	if curTimeError != nil {
		return "", fmt.Errorf("x> error parsing the current time for creating the log file name: %s", curTimeError.Error())
	}

	return fmt.Sprintf("%s%s.log", fileNamePrefix, curTime), nil
}
