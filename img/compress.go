package img

import (
	"errors"
	"fmt"
	"golang.org/x/image/draw"
	"image"
	"image/png"
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"strings"
)

//walk dir and compress all files recursively
func WalkDirAndCompressFiles(dirPath string, fileExtensions []string, fileExtensionsNotAllowed []string, dimensionMax int) []error {
	var errorContainer []error
	filepath.WalkDir(dirPath, func(path string, fileInfo fs.DirEntry, walkFuncError error) error {
		if walkFuncError != nil {
			return walkFuncError
		}
		fileName := fileInfo.Name()

		if !fileInfo.IsDir() && FileHasAnyFileExtension(fileName, fileExtensions) && !FileHasAnyFileExtension(fileName, fileExtensionsNotAllowed) {
			//compress file
			compressError := CompressPng(path, dimensionMax)

			if compressError != nil {
				fmt.Printf("=> %v\n", compressError)
				errorContainer = append(errorContainer, compressError)
			}
		}

		return nil
	})

	return errorContainer
}

//checks whether a filename contains one of a list of file extensions
func FileHasAnyFileExtension(fileName string, fileExtensions []string) bool {
	for _, fileExtension := range fileExtensions {
		if strings.Contains(fileName, fileExtension) {
			return true
		}
	}

	return false
}

func CompressPng(filePath string, dimensionMax int) error {
	fmt.Println("=> Attempting to compress png file: " + filePath)

	file, openFileError := os.Open(filePath)
	if openFileError != nil {
		return fmt.Errorf("error opening png file: %v", openFileError)
	}

	newImage, _, decodeToImageError := image.Decode(file)
	if decodeToImageError != nil {
		return fmt.Errorf("error decoding image from file: %v", decodeToImageError)
	}

	//get pixel dimension of image
	bounds := newImage.Bounds()
	fmt.Printf("=> Pixel size of image: %dx%d\n", bounds.Dx(), bounds.Dy())
	var width, height int

	if bounds.Dx() > dimensionMax || bounds.Dy() > dimensionMax {
		//get scaled with and height for image
		scaleFactor := float64(dimensionMax) / math.Max(float64(bounds.Dx()), float64(bounds.Dy()))
		width = int(float64(bounds.Dx()) * scaleFactor)
		height = int(float64(bounds.Dy()) * scaleFactor)

		if width > 0 && height > 0 {
			newImageRGBA := image.NewRGBA(image.Rect(0, 0, width, height))

			draw.CatmullRom.Scale(newImageRGBA, newImageRGBA.Bounds(), newImage, bounds, draw.Over, nil)

			encoder := new(png.Encoder)

			//set to highest compression level
			encoder.CompressionLevel = -2

			//create new file
			newFile, newFileError := os.Create(filePath)

			if newFileError != nil {
				return newFileError
			}

			encodePngError := encoder.Encode(newFile, newImageRGBA)

			if encodePngError != nil {
				return encodePngError
			}
		} else {
			return errors.New("width or height of new image is 0")
		}

		fmt.Printf("=> file %s has been compressed\n", filePath)
	} else {
		fmt.Println("=> Image is already small enough. no action needed...")
	}

	return nil
}
