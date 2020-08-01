package config

import (
	"bytes"
	"github.com/spf13/viper"
	"gitlab.id.vin/devops/go-kit/src/cmd/logger"
	"gitlab.id.vin/devops/go-kit/src/cmd/util/file_utils"
	"io"
	"io/ioutil"
	"strings"
)

//Read config by viper (support merge multiple config keys)
func ReadConfigByViper(isMerge bool, values []byte) error {
	reader := bytes.NewReader(values)
	defer io.Copy(ioutil.Discard, reader)

	if isMerge {
		return viper.MergeConfig(reader)
	}
	return viper.ReadConfig(reader)
}

//Load config files by viper
func LoadConfigFilesByViper(configFilePaths []string) error {
	length := len(configFilePaths)
	if length == 0 {
		logger.Warnf("No config file paths have found, ignoring read configs...")
		return nil
	}

	viper.SetConfigType("yaml")

	for i := 0; i < length; i++ {
		isMerge := i != 0
		filePath := configFilePaths[i]

		//ignore empty config file paths
		f := strings.Trim(filePath, " ")
		if f == "" {
			continue
		}
		//read content file
		content, err := file_utils.ReadLocalFile(filePath)
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

	return nil
}