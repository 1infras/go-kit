package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"

	"github.com/1infras/go-kit/driver/config"
	"github.com/1infras/go-kit/driver/consul"
	"github.com/1infras/go-kit/util"
)

func ReadRemoteConfigFilesWithConsul(kv consul.KV, keys []string, isMerge bool) error {
	length := len(keys)
	if length == 0 {
		return nil
	}

	for i := 0; i < length; i++ {
		key := strings.Trim(keys[i], " ")
		if key == "" {
			continue
		}

		v, err := kv.GetKV(key)
		if err != nil {
			return fmt.Errorf("read key %s with error: %s", key, err.Error())
		}

		ext, err := util.GetExtension(key)
		if err != nil {
			return fmt.Errorf("key: %s is not valid file", key)
		}

		if ok := config.IsExtFileViperSupported(ext); !ok {
			return fmt.Errorf("key: %s is not support with ext: %s", key, ext)
		}

		viper.SetConfigType(ext)

		if err := config.ReadConfigWithViper(isMerge, v); err != nil {
			return fmt.Errorf("read key %s with error: %s", key, err.Error())
		}
	}

	return nil
}
