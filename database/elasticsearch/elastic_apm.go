package elasticsearch

import (
	"os"

	"github.com/spf13/cast"
)

const (
	//DefaultAPMServer -
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

//DefaultAPMConnection -
func DefaultAPMConnection() *APMConnection {
	return &APMConnection{
		URL: DefaultAPMServer,
	}
}

//AutoBindEnvironment -
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
