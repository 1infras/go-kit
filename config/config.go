package config

import (
	"fmt"
	"strings"

	"github.com/1infras/go-kit/driver/consul"
)

type Config struct {
	ConfigType        string
	PrefixEnvironment string
	Names             string
	ConsulKV          consul.KV
}

func AutomateReadConfig(cfg *Config) error {
	if cfg.ConfigType == "skip" {
		return nil
	}

	var (
		names = strings.Split(cfg.Names, ",")
		err   error
	)

	if len(names) == 0 {
		return fmt.Errorf("at least one of name of config must be defined")
	}

	switch cfg.ConfigType {
	case "local":
		err = ReadLocalConfigFilesWithViper(names, true)
	case "remote":
		err = ReadRemoteConfigFilesWithConsul(cfg.ConsulKV, names, true)
	case "environment":
		err = ReadLocalConfigEnvironmentsWithViper(cfg.PrefixEnvironment, names)
	default:
		err = fmt.Errorf("invalid config type, please choose one of optionals skip | local | remote")
	}

	return err
}
