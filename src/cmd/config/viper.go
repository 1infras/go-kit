package config

import (
	"fmt"
	"github.com/spf13/viper"
	"gitlab.id.vin/devops/go-kit/src/cmd/logger"
	"os"
)

//Read config with file
func InitViper(filePath string) error {
	logger.Debug(nil, fmt.Sprintf("reading config at: %v", filePath))
	info, err := os.Stat(filePath)
	if err != nil {
		logger.Warnf("the config file was not found, config with file will be skipped")
		return nil
	}

	//Ensure config file is not a directory
	if info.IsDir() {
		return fmt.Errorf("the config file path is not a file")
	}

	//Set config file
	viper.SetConfigFile(filePath)
	viper.AutomaticEnv()

	//Load config from file
	return viper.MergeInConfig()
}
