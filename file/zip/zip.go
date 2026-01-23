package zip

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Zip struct {
	FileName      string
	FilesToCreate map[string][]byte
	FilesToCopy   []string
}

func (z *Zip) AddFile(fileName string, contents []byte) error {

	if z.FilesToCreate == nil {
		z.FilesToCreate = make(map[string][]byte)
	}

	if strings.TrimSpace(fileName) == "" {
		return errors.New("parameter fileName is empty")
	}
	z.FilesToCreate[fileName] = contents
	return nil
}

func (z *Zip) CopyFile(fileName string) error {
	if strings.TrimSpace(fileName) == "" {
		return errors.New("parameter fileName is empty")
	}
	z.FilesToCopy = append(z.FilesToCopy, fileName)
	return nil
}

func (z *Zip) Create() error {

	archive, err := os.Create(z.FileName)
	if err != nil {
		return err
	}
	defer archive.Close()

	writer := zip.NewWriter(archive)
	defer writer.Close()

	//create new files into zip
	for fName, content := range z.FilesToCreate {
		file, err := writer.Create(fName)
		if err != nil {
			return err
		}
		_, err = file.Write(content)
		if err != nil {
			return err
		}
	}

	//copy files into zip
	for _, path := range z.FilesToCopy {
		fileInfo, err := os.Stat(path)
		if err != nil {
			return err
		}

		if fileInfo.IsDir() {
			z.addDirToZip(writer, path)
		} else {
			z.addFileToZip(writer, path, "")
		}
	}

	return nil
}

func (z *Zip) addDirToZip(writer *zip.Writer, rootDir string) error {
	return filepath.WalkDir(rootDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(filepath.Dir(rootDir), path)
		if err != nil {
			return err
		}

		relPath = filepath.ToSlash(relPath)

		if d.IsDir() {
			_, err := writer.Create(relPath + "/")
			return err
		}

		return z.addFileToZip(writer, path, relPath)
	})
}

func (z *Zip) addFileToZip(writer *zip.Writer, systemPath string, zipPath string) error {
	if zipPath == "" {
		zipPath = filepath.Base(systemPath)
	}

	file, err := os.Open(systemPath)
	if err != nil {
		return err
	}

	defer file.Close()

	fInfo, err := file.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(fInfo)
	if err != nil {
		return err
	}

	header.Name = zipPath
	header.Method = zip.Deflate

	writerFile, err := writer.CreateHeader(header)
	if err != nil {
		return err
	}

	_, err = io.Copy(writerFile, file)
	return err
}

func (z *Zip) Load(fileName string) error {
	z.FileName = filepath.Base(fileName)

	archive, err := zip.OpenReader(fileName)
	if err != nil {
		return err
	}

	defer archive.Close()

	for _, file := range archive.File {
		fmt.Println(file.Name)

		if file.FileInfo().IsDir() {
			continue
		}

		reader, err := file.Open()
		if err != nil {
			return err
		}

		contents, err := io.ReadAll(reader)
		if err != nil {
			return err
		}

		z.FilesToCreate[file.Name] = contents
	}

	return nil
}

// unzips the loaded zip into the destination directory
// destination directory has to be an absolute path
func (z *Zip) Unzip(destination string) error {
	//check if destination exists and return error if it does
	_, err := os.Stat(destination)

	if !os.IsNotExist(err) {
		return errors.New("target directory already exists")
	}

	if len(z.FilesToCreate) == 0 {
		return errors.New("no files to unzip")
	}

	for fName, content := range z.FilesToCreate {
		destFilePath := filepath.Join(destination, fName)

		dirPath := filepath.Dir(destFilePath)

		fmt.Println("DESTINATION DIR: " + dirPath + " | DESTINATION PATH: " + destFilePath)

		err = os.MkdirAll(dirPath, 0755)
		if err != nil {
			return err
		}

		err = os.WriteFile(destFilePath, content, 0755)
		if err != nil {
			return err
		}
	}

	return nil
}
