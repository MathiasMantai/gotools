package osutil

import (
	"os"
)


// move a file from the source path to the destination path
// if fSync is true, file.Sync() will be called, which means the change will be immediate (meaning not written to the buffer first) (caution: uses more ressources)
// if truncate is false and the file already exist in the destination path, an error will be returned
func MoveFile(sourcePath string, destPath string, fSync bool) error {

	file, err := os.ReadFile(sourcePath)
	if err != nil {
		return err
	}

	dest, err := os.Open(destPath)
	if err != nil {
		return err
	}

	defer dest.Close()

	_, err = dest.Write([]byte(file))
	if err != nil {
		return err
	}

	if fSync {
		err = dest.Sync()
		if err != nil {
			return err
		}
	}

	return nil
}