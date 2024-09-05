package img

import (
	"os"
	"fmt"
	"image"
	"strings"
	"image/png"
	"image/jpeg"
)

func LoadPng(filePath string) (*image.Image, error) {
	imgFile, openError := os.Open(filePath)
    if openError!= nil {
        return nil, openError
    }
    defer imgFile.Close()

    img, decodeError := png.Decode(imgFile)
    if decodeError!= nil {
        return nil, decodeError
    }

    return &img, nil
}

func LoadJpeg(filePath string) (*image.Image, error) {

	if !strings.Contains(filePath, "jpeg") && !strings.Contains(filePath, "jpg") {
		return nil, fmt.Errorf("error ")
	}

	imgFile, openError := os.Open(filePath)
    if openError != nil {
        return nil, openError
    }
    defer imgFile.Close()

    img, decodeError := jpeg.Decode(imgFile)
    if decodeError != nil {
        return nil, decodeError
    }

    return &img, nil
}

