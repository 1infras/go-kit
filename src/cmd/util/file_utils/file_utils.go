package fileutils

import (
	"fmt"
	"os"
)

//ReadLocalFile - Read a local file with file path
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
