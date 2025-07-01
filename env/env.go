package env

import (
	"fmt"
	"os"
	"strings"
)

func Load(filePaths []string) error {
	for _, filePath := range filePaths {
		content, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("file %v not found", filePath)
		}
		contentParts := strings.Split(string(content), "\n")
		for _, envRow := range contentParts {
			if strings.TrimSpace(envRow) == "" {
				continue
			}

			envParts := strings.Split(envRow, "=")
			err = os.Setenv(envParts[0], strings.TrimSpace(envParts[1]))
			if err != nil {
				return fmt.Errorf("environment variable %v could not be set", envParts[0])
			}
		}
	}

	return nil
}

func Get(key string, defaultValue string) string {
	res := os.Getenv(key)
	if strings.TrimSpace(res) == "" {
		return defaultValue
	}
	return res
}
