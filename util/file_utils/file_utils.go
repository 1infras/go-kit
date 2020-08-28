package file_utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

//ReadLocalFile
func ReadLocalFile(fileName string) ([]byte, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}

	defer f.Close()
	// validate file's size
	fileInfo, _ := f.Stat()
	size := fileInfo.Size()
	// throw error if it's a empty file
	if size == 0 {
		return nil, fmt.Errorf("file %v is empty", fileName)
	}
	// load file content to buffer
	buffer := make([]byte, size)
	f.Read(buffer)

	return buffer, nil
}

//GetAbsolutelyLocalFilePath
func GetAbsolutelyLocalFilePath(path string) (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// If input file path is not absolutely filepath, let's join it with current working directory
	if !filepath.IsAbs(path) {
		path = filepath.Join(wd, strings.Trim(path, "."))
	}

	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}

	//
	if info.IsDir() {
		return "", fmt.Errorf("%v is is not a file", path)
	}

	return path, nil
}
