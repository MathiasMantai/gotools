package json

import (
	"encoding/json"
	"os"
)

type Config struct {
	Data map[string]interface{}
}

func (c *Config) Fetch(filePath string) error {
	fileContent, readError := os.ReadFile(filePath)
	if readError != nil {
		return readError
	}

	unmarshalError := json.Unmarshal([]byte(fileContent), &c.Data)
	if unmarshalError != nil {
		return unmarshalError
	}

	return nil
}

func (c *Config) Get(key string) interface{} {
	return c.Data[key]
}
