package file_utils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExistLocalFilet(t *testing.T) {
	ok := ExistLocalFile("file_utils.go")
	assert.Equal(t, true, ok)
}

func TestGetExtLocalFile(t *testing.T) {
	ext, err := GetExtLocalFile("file_utils.go")
	assert.Nil(t, err)
	assert.Equal(t, "go", ext)
}

func TestGetAbsolutelyLocalFilePath(t *testing.T) {
	wd, err := os.Getwd()
	assert.Nil(t, err)

	p, err := GetAbsolutelyLocalFilePath("file_utils.go")
	assert.Nil(t, err)

	assert.Equal(t, filepath.Join(wd, "file_utils.go"), p)
}

func TestReadLocalFile(t *testing.T) {
	_, err := ReadLocalFile("file_utils.go")
	assert.Nil(t, err)
}
