package osutil

import (
	"errors"
	"os"
)

/*
*
checks whether a file or directory exists
*/
func FileExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}

	return true
}

func GetSystemLanguage() (string, error) {
	lang, error := os.LookupEnv("LANG")

	if !error {
		return "", errors.New("system language could not be determined")
	}

	return lang, nil
}
