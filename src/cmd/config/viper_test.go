package config

import (
	"github.com/1infras/go-kit/src/cmd/logger"
	"github.com/spf13/viper"
	"testing"
)

func TestInitViper(t *testing.T) {
	logger.InitLogger(logger.DebugLevel)
	err := LoadConfigFilesByViper([]string{"config.yml"})
	if err != nil {
		t.Fatal(err)
	}

	keys := viper.AllKeys()
	for _, k := range keys {
		logger.Debug(viper.GetString(k))
	}
}
