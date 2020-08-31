package kafka

import (
	"crypto/tls"
	"fmt"

	"github.com/1infras/go-kit/util/cert_utils"
	"github.com/1infras/go-kit/util/file_utils"
	"github.com/spf13/viper"
)

//Connection of Kafka Cluster/Standalone
type Connection struct {
	Brokers []string    `json:"brokers"` //The list of kafka brokers
	TLS     *tls.Config `json:"tls"`     //SSL configuration
}

//NewDefaultKafkaConnection with settings from viper
func NewDefaultKafkaConnection() (*Connection, error) {
	servers := viper.GetStringSlice("kafka.servers")
	if len(servers) == 0 {
		return nil, fmt.Errorf("kafka servers must be defined")
	}

	c := &Connection{
		Brokers: servers,
	}

	if viper.GetBool("kafka.tls") {
		tlsClientCert, err := file_utils.GetAbsolutelyLocalFilePath(viper.GetString("kafka.tls_client_cert"))
		if err != nil {
			return nil, err
		}

		tlsClientKey, err := file_utils.GetAbsolutelyLocalFilePath(viper.GetString("kafka.tls_client_key"))
		if err != nil {
			return nil, err
		}

		tlsClientCA := viper.GetString("kafka.tls_client_ca")
		tlsSkipVerify := viper.GetBool("kafka.tls_skip_verify")

		if tlsClientCA != "" {
			tlsClientCA, err = file_utils.GetAbsolutelyLocalFilePath(viper.GetString("kafka.tls_client_ca"))
			if err != nil {
				return nil, err
			}
		}

		tlsConfig, err := cert_utils.NewTLS(tlsClientCert, tlsClientKey, tlsClientCA, tlsSkipVerify)
		if err != nil {
			return nil, err
		}

		c.TLS = tlsConfig
	}

	return c, nil
}