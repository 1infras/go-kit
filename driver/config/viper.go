package config

import (
	"bytes"
	"io"
	"io/ioutil"

	"github.com/spf13/viper"
)

var (
	SupportedExts = []string{"json", "hcl", "toml", "yaml"}
)

func IsExtFileViperSupported(ext string) bool {
	for _, v := range SupportedExts {
		if ext == v {
			return true
		}
	}
	return false
}

func ReadConfigWithViper(isMerge bool, values []byte) error {
	reader := bytes.NewReader(values)
	defer func() {
		_, _ = io.Copy(ioutil.Discard, reader)
	}()

	if isMerge {
		return viper.MergeConfig(reader)
	}
	return viper.ReadConfig(reader)
}

func ReadEnvironmentWithViper(prefix string, keys ...string) error {
	viper.SetEnvPrefix(prefix)
	viper.AutomaticEnv()
	return viper.BindEnv(keys...)
}