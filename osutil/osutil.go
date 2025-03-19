package osutil

import (
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
