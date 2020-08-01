package config

import (
	"github.com/spf13/viper"
	"gitlab.id.vin/devops/go-kit/src/cmd/logger"
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
