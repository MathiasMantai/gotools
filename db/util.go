package db

import (
	"strings"
)

func RemoveFileExtension(file string) string {
	return strings.Split(file, ".")[0]
}
