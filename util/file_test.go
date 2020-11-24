package util

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExistLocalFilet(t *testing.T) {
	ok := Exist("file.go")
	assert.Equal(t, true, ok)
}

func TestGetExtLocalFile(t *testing.T) {
	ext, err := GetExtension("file.go")
	assert.Nil(t, err)
	assert.Equal(t, "go", ext)
}

func TestGetAbsolutelyLocalFilePath(t *testing.T) {
	wd, err := os.Getwd()
	assert.Nil(t, err)

	p, err := GetAbsolutelyPath("file.go")
	assert.Nil(t, err)

	assert.Equal(t, filepath.Join(wd, "file.go"), p)
}

func TestReadLocalFile(t *testing.T) {
	_, err := Read("file.go")
	assert.Nil(t, err)
}
