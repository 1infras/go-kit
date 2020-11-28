package kafka

import (
	"crypto/tls"
	"fmt"

	"github.com/Shopify/sarama"
	"github.com/kelseyhightower/envconfig"

	"github.com/1infras/go-kit/util"
)

// Config
type Config struct {
	Brokers               []string `mapstructure:"brokers" envconfig:"KAFKA_BROKERS"`
	Version               string   `mapstructure:"version" envconfig:"KAFKA_VERSION"`
	TLS                   bool     `mapstructure:"tls" envconfig:"KAFKA_TLS"`
	Certificate           string   `mapstructure:"tls_client_cert" envconfig:"KAFKA_CERTIFICATE"`
	PrivateKey            string   `mapstructure:"tls_client_key" envconfig:"KAFKA_PRIVATE_KEY"`
	CertificateAuthority  string   `mapstructure:"tls_client_ca" envconfig:"KAFKA_CERTIFICATE_AUTHORITY"`
	SkipVerifyCertificate bool     `mapstructure:"tls_skip_verify" envconfig:"KAFKA_SKIP_VERIFY_CERTIFICATE"`
}

// Kafka
type Kafka struct {
	Brokers []string            // The list of kafka brokers
	TLS     *tls.Config         // SSL configuration
	Version sarama.KafkaVersion // Kafka version
}

// ProcessConfig
func ProcessConfig(cfg *Config) (*Config, error) {
	if cfg == nil {
		cfg = &Config{}
		err := envconfig.Process("kafka", cfg)
		if err != nil {
			return nil, err
		}
	}

	if cfg.Version == "" {
		cfg.Version = sarama.V2_6_0_0.String()
	}

	return cfg, nil
}

// NewKafka
func NewKafka(c *Config) (*Kafka, error) {
	cfg, err := ProcessConfig(c)
	if err != nil {
		return nil, err
	}

	connection := &Kafka{
		Brokers: cfg.Brokers,
	}

	if cfg.TLS {
		tlsConfig, err := util.NewTLS(&util.TLS{
			CertificateFile:          cfg.Certificate,
			PrivateKeyFile:           cfg.PrivateKey,
			CertificateAuthorityFile: cfg.CertificateAuthority,
			SkipVerifyCertificate:    cfg.SkipVerifyCertificate,
		})
		if err != nil {
			return nil, err
		}

		connection.TLS = tlsConfig
	}

	version, err := sarama.ParseKafkaVersion(cfg.Version)
	if err != nil {
		return nil, fmt.Errorf("parse kafka verison has error: %v", err)
	}

	connection.Version = version

	return connection, nil
}
