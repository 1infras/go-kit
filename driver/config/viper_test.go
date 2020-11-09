package config

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestReadConfigWithViper(t *testing.T) {
	f1 := struct {
		Foo string `json:"foo"`
	}{
		Foo: "bar",
	}

	b1, err := json.Marshal(f1)
	assert.Nil(t, err)

	viper.SetConfigType("json")
	err = ReadConfigWithViper(false, b1)
	assert.Nil(t, err)

	assert.Equal(t, "bar", viper.GetString("foo"))

	f2 := struct {
		Bar string `json:"bar"`
	}{
		Bar: "foo",
	}

	b2, err := json.Marshal(f2)
	assert.Nil(t, err)

	viper.SetConfigType("json")
	err = ReadConfigWithViper(true, b2)
	assert.Nil(t, err)

	assert.Equal(t, "bar", viper.GetString("foo"))
	assert.Equal(t, "foo", viper.GetString("bar"))
}

func TestReadEnvironmentWithViper(t *testing.T) {
	os.Setenv("MY_FOO", "foo")
	os.Setenv("MY_BAR", "bar")

	err := ReadEnvironmentWithViper("my", "foo", "bar")
	assert.Nil(t, err)

	assert.Equal(t, "bar", viper.GetString("bar"))
	assert.Equal(t, "foo", viper.GetString("foo"))
}

func TestIsExtFileViperSupported(t *testing.T) {
	assert.Equal(t, true, IsExtFileViperSupported("yaml"))
	assert.Equal(t, true, IsExtFileViperSupported("json"))
	assert.Equal(t, true, IsExtFileViperSupported("hcl"))
	assert.Equal(t, true, IsExtFileViperSupported("toml"))
	assert.Equal(t, false, IsExtFileViperSupported("java"))
	assert.Equal(t, false, IsExtFileViperSupported("go"))
}
