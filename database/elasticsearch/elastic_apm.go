package elasticsearch

import (
	"github.com/spf13/cast"
	"github.com/spf13/viper"
	"os"
)

const (
	//DefaultAPMServer
	DefaultAPMServer = "http://localhost:8200"
)

//APMConnection - Connection to APM Server
type APMConnection struct {
	URL         string `json:"url"`
	SecretToken string `json:"secret_token"`
	ServiceName string `json:"service_name"`
	Environment string `json:"environment"`
	Active      bool   `json:"active"`
}

func APMConnectionWithViper() *APMConnection {
	return &APMConnection{
		URL:         viper.GetString("elasticsearch_apm.url"),
		SecretToken: viper.GetString("elasticsearch_apm.secret_token"),
		ServiceName: viper.GetString("elasticsearch_apm.service_name"),
		Environment: viper.GetString("elasticsearch_apm.environment"),
		Active:      viper.GetBool("elasticsearch_apm.active"),
	}
}

func DefaultAPMConnection() *APMConnection {
	return &APMConnection{
		URL: DefaultAPMServer,
	}
}

func (c *APMConnection) AutoBindEnvironment() {
	if c.URL != "" {
		os.Setenv("ELASTIC_APM_SERVER_URL", c.URL)
	}

	if c.SecretToken != "" {
		os.Setenv("ELASTIC_APM_SECRET_TOKEN", c.SecretToken)
	}

	if c.ServiceName != "" {
		os.Setenv("ELASTIC_APM_SERVICE_NAME", c.ServiceName)
	}

	if c.Environment != "" {
		os.Setenv("ELASTIC_APM_ENVIRONMENT", c.Environment)
	}

	os.Setenv("ELASTIC_APM_ACTIVE", cast.ToString(c.Active))
}
