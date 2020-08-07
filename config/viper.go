package config

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	"github.com/1infras/go-kit/logger"
	fileutils "github.com/1infras/go-kit/util/file_utils"
	"github.com/spf13/viper"
)

//ReadConfigByViper - Read config by viper (support merge multiple config files)
func ReadConfigByViper(isMerge bool, values []byte) error {
	reader := bytes.NewReader(values)
	defer io.Copy(ioutil.Discard, reader)

	if isMerge {
		return viper.MergeConfig(reader)
	}
	return viper.ReadConfig(reader)
}

//LoadConfigFilesByViper - Load config files by viper
func LoadConfigFilesByViper(configFilePaths []string, configType string) error {
	length := len(configFilePaths)
	if length == 0 {
		logger.Warnf("No config file paths have found, ignoring read configs...")
		return nil
	}

	if configType != "json" && configType != "toml" && configType != "yaml" {
		return fmt.Errorf("config type must be is JSON/TOML/YAML")
	}

	viper.SetConfigType(configType)

	for i := 0; i < length; i++ {
		isMerge := i != 0
		filePath := configFilePaths[i]

		//ignore empty config file paths
		f := strings.Trim(filePath, " ")
		if f == "" {
			continue
		}
		//read content file
		content, err := fileutils.ReadLocalFile(filePath)
		if err != nil {
			return err
		}

		//load config by viper
		err = ReadConfigByViper(isMerge, content)
		if err != nil {
			return err
		}

		logger.Infof("Loaded config from config file: %v", filePath)
	}

	viper.AutomaticEnv()

	return nil
}
