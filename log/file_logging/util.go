package log

import (
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
