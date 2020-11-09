package kafka

import (
	"crypto/tls"
	"fmt"

	"github.com/kelseyhightower/envconfig"

	"github.com/spf13/viper"

	"github.com/1infras/go-kit/util/cert_utils"
	"github.com/1infras/go-kit/util/file_utils"
)

// Config of Kafka Cluster/Standalone
type Config struct {
	Brokers               []string `json:"brokers" envconfig:"KAFKA_BROKERS"`
	TLS                   bool     `json:"tls" envconfig:"KAFKA_TLS"`
	Certificate           string   `json:"tls_client_cert" envconfig:"KAFKA_CERTIFICATE"`
	PrivateKey            string   `json:"tls_client_key" envconfig:"KAFKA_PRIVATE_KEY"`
	CertificateAuthority  string   `json:"tls_client_ca" envconfig:"KAFKA_CERTIFICATE_AUTHORITY"`
	SkipVerifyCertificate bool     `json:"tls_skip_verify" envconfig:"KAFKA_SKIP_VERIFY_CERTIFICATE"`
}

// Connection of Kafka Cluster/Standalone
type Connection struct {
	Brokers []string    `json:"brokers"` // The list of kafka brokers
	TLS     *tls.Config `json:"tls"`     // SSL configuration
}

// NewDefaultKafkaConfig of Kafka Clusters/Standalone
func NewDefaultKafkaConfig() (*Config, error) {
	var (
		cfg *Config
		err error
	)

	err = envconfig.Process("kafka", &cfg)
	if err != nil {
		cfg = &Config{
			Brokers:               viper.GetStringSlice("kafka.servers"),
			Certificate:           viper.GetString("kafka.tls_client_cert"),
			PrivateKey:            viper.GetString("kafka.tls_client_key"),
			CertificateAuthority:  viper.GetString("kafka.tls_client_ca"),
			SkipVerifyCertificate: viper.GetBool("kafka.tls_skip_verify"),
		}
	}

	if cfg.TLS {
		tlsClientCert, err := file_utils.GetAbsolutelyLocalFilePath(cfg.Certificate)
		if err != nil {
			return nil, err
		}

		cfg.Certificate = tlsClientCert

		tlsClientKey, err := file_utils.GetAbsolutelyLocalFilePath(cfg.PrivateKey)
		if err != nil {
			return nil, err
		}

		cfg.PrivateKey = tlsClientKey

		if cfg.CertificateAuthority != "" {
			tlsClientCA, err := file_utils.GetAbsolutelyLocalFilePath(viper.GetString("kafka.tls_client_ca"))
			if err != nil {
				return nil, err
			}

			cfg.CertificateAuthority = tlsClientCA
		}
	}

	return cfg, nil
}

// NewDefaultKafkaConnection with settings from viper
func NewDefaultKafkaConnection(cfg *Config) (*Connection, error) {
	if cfg == nil {
		config, err := NewDefaultKafkaConfig()
		if err != nil {
			return nil, err
		}

		cfg = config
	}

	if len(cfg.Brokers) == 0 {
		return nil, fmt.Errorf("kafka servers must be defined")
	}

	c := &Connection{
		Brokers: cfg.Brokers,
	}

	if cfg.TLS {
		tlsConfig, err := cert_utils.NewTLS(cfg.Certificate, cfg.PrivateKey, cfg.CertificateAuthority, cfg.SkipVerifyCertificate)
		if err != nil {
			return nil, err
		}

		c.TLS = tlsConfig
	}

	return c, nil
}
