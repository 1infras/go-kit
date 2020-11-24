package config

import (
	"os"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"github.com/1infras/go-kit/driver/consul"
	"github.com/1infras/go-kit/util"
)

func TestReadLocalConfigFilesWithViper(t *testing.T) {
	err := ReadLocalConfigFilesWithViper([]string{"config_test.yaml"}, false)
	assert.Nil(t, err)
	err = ReadLocalConfigFilesWithViper([]string{"config_test.json"}, true)
	assert.Nil(t, err)

	assert.Equal(t, "http://localhost:9200", viper.GetString("elasticsearch.url"))
	assert.Equal(t, false, viper.GetBool("elasticsearch.secure"))
	assert.Equal(t, "123456", viper.GetString("elasticsearch.api_key"))
	assert.Equal(t, "localhost:3306", viper.GetString("mysql.address"))
	assert.Equal(t, "root", viper.GetString("mysql.username"))
	assert.Equal(t, "123456", viper.GetString("mysql.password"))
}

func TestReadRemoteConfigFilesWithConsul(t *testing.T) {
	kv, err := consul.NewConsul(&consul.Config{
		Endpoint: "http://localhost:8500",
	})
	assert.Nil(t, err)

	f1, err := util.Read("config_test.yaml")
	assert.Nil(t, err)

	f2, err := util.Read("config_test.json")
	assert.Nil(t, err)

	keys, err := kv.GetKeys("app")
	assert.Nil(t, err)

	if keys == nil {
		err = kv.PutKV("app/config_test.yaml", f1)
		assert.Nil(t, err)
		err = kv.PutKV("app/config_test.json", f2)
		assert.Nil(t, err)
	}

	err = ReadRemoteConfigFilesWithConsul(kv, []string{"app/config_test.yaml", "app/config_test.json"}, true)
	assert.Nil(t, err)

	assert.Equal(t, "http://localhost:9200", viper.GetString("elasticsearch.url"))
	assert.Equal(t, false, viper.GetBool("elasticsearch.secure"))
	assert.Equal(t, "123456", viper.GetString("elasticsearch.api_key"))
	assert.Equal(t, "localhost:3306", viper.GetString("mysql.address"))
	assert.Equal(t, "root", viper.GetString("mysql.username"))
	assert.Equal(t, "123456", viper.GetString("mysql.password"))
}

func TestAutomateReadConfig(t *testing.T) {
	var err error

	cfg := &Config{
		ConfigType: "skip",
	}
	err = AutomateReadConfig(cfg)
	assert.Nil(t, err)

	cfg.ConfigType = "default"
	err = AutomateReadConfig(cfg)
	assert.NotNil(t, err)

	cfg.ConfigType = "local"
	cfg.Names = "config_test.json,config_test.yaml"
	err = AutomateReadConfig(cfg)
	assert.Nil(t, err)
	assert.Equal(t, "http://localhost:9200", viper.GetString("elasticsearch.url"))
	assert.Equal(t, false, viper.GetBool("elasticsearch.secure"))
	assert.Equal(t, "123456", viper.GetString("elasticsearch.api_key"))
	assert.Equal(t, "localhost:3306", viper.GetString("mysql.address"))
	assert.Equal(t, "root", viper.GetString("mysql.username"))
	assert.Equal(t, "123456", viper.GetString("mysql.password"))

	kv, err := consul.NewConsul(&consul.Config{
		Endpoint: "http://localhost:8500",
	})
	assert.Nil(t, err)

	cfg.ConfigType = "remote"
	cfg.ConsulKV = kv
	cfg.Names = "app/config_test.yaml,app/config_test.json"
	f1, err := util.Read("config_test.yaml")
	assert.Nil(t, err)

	f2, err := util.Read("config_test.json")
	assert.Nil(t, err)

	keys, err := kv.GetKeys("app")
	assert.Nil(t, err)

	if keys == nil {
		err = kv.PutKV("app/config_test.yaml", f1)
		assert.Nil(t, err)
		err = kv.PutKV("app/config_test.json", f2)
		assert.Nil(t, err)
	}
	err = AutomateReadConfig(cfg)
	assert.Nil(t, err)
	assert.Equal(t, "http://localhost:9200", viper.GetString("elasticsearch.url"))
	assert.Equal(t, false, viper.GetBool("elasticsearch.secure"))
	assert.Equal(t, "123456", viper.GetString("elasticsearch.api_key"))
	assert.Equal(t, "localhost:3306", viper.GetString("mysql.address"))
	assert.Equal(t, "root", viper.GetString("mysql.username"))
	assert.Equal(t, "123456", viper.GetString("mysql.password"))

	cfg.ConfigType = "environment"
	cfg.PrefixEnvironment = "my"
	cfg.Names = "foo,bar"
	_ = os.Setenv("MY_FOO", "FOO")
	_ = os.Setenv("MY_BAR", "BAR")

	err = AutomateReadConfig(cfg)
	assert.Nil(t, err)
	assert.Equal(t, "FOO", viper.GetString("FOO"))
	assert.Equal(t, "BAR", viper.GetString("BAR"))
}
