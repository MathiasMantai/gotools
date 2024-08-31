package language

import (
	"errors"
	"os"
)

func GetSystemLanguage() (string, error) {
	lang, error := os.LookupEnv("LANG")

	if !error {
		return "", errors.New("system language could not be determined")
	}

	return lang, nil
}
