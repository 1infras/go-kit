package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"

	"github.com/1infras/go-kit/driver/config"
	"github.com/1infras/go-kit/util/file_utils"
)

func ReadLocalConfigFilesWithViper(filePaths []string, isMerge bool) error {
	length := len(filePaths)
	if length == 0 {
		return fmt.Errorf("at least one config's file path must be defined")
	}

	for i := 0; i < length; i++ {
		filePath := filePaths[i]

		if !file_utils.ExistLocalFile(filePath) {
			return fmt.Errorf("the file path: %s is not exist", filePath)
		}

		f := strings.Trim(filePath, " ")
		if f == "" {
			continue
		}

		ext, err := file_utils.GetExtLocalFile(filePath)
		if err != nil {
			return err
		}

		if ok := config.IsExtFileViperSupported(ext); !ok {
			return fmt.Errorf("the file path: %s is not support", filePath)
		}

		content, err := file_utils.ReadLocalFile(filePath)
		if err != nil {
			return err
		}

		viper.SetConfigType(ext)

		err = config.ReadConfigWithViper(isMerge, content)
		if err != nil {
			return err
		}
	}

	return nil
}

func ReadLocalConfigEnvironmentsWithViper(prefix string, keys []string) error {
	if prefix == "" {
		return fmt.Errorf("prefix must not empty")
	}

	if len(keys) == 0 {
		return fmt.Errorf("at least one configs's environment must be defined")
	}

	return config.ReadEnvironmentWithViper(prefix, keys...)
}
