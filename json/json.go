package json

import (
	"encoding/json"
	"os"
)

func IntoStruct(filePath string, target interface{}) error {
	data, readFileError := os.ReadFile(filePath)

	if readFileError != nil {
		return readFileError
	}

	unmarshalError := json.Unmarshal(data, target)
	if unmarshalError != nil {
		return unmarshalError
	}

	return nil
}
