package elastic

import "C"
import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/kelseyhightower/envconfig"

	"github.com/1infras/go-kit/tracing"
)

const (
	// DefaultElasticURL -
	DefaultElasticURL = "http://localhost:9200"
	// DefaultMaxRetries -
	DefaultMaxRetries = 3
	// DefaultRetryAfter -
	DefaultRetryAfter = 5 * time.Second
)

// Config - Config connection to ElasticSearch
type Config struct {
	URL         string        `mapstructure:"url" envconfig:"ELASTIC_URL"`
	Secure      bool          `mapstructure:"secure" envconfig:"ELASTIC_SECURE"`
	APIKey      string        `mapstructure:"api_key" envconfig:"ELASTIC_API_KEY"`
	Username    string        `mapstructure:"username" envconfig:"ELASTIC_USERNAME"`
	Password    string        `mapstructure:"password" envconfig:"ELASTIC_PASSWORD"`
	MaxRetries  int           `mapstructure:"max_retries" envconfig:"ELASTIC_MAX_RETRIES"`
	RetryAfter  time.Duration `mapstructure:"retry_after" envconfig:"ELASTIC_RETRY_AFTER"`
}

// Transport -
type Transport struct {
	APIKey   string
	Username string
	Password string
}

// RoundTrip - Wrap RoundTrip to adding API Key/Basic Authenticate in header to do authorization
func (_this *Transport) RoundTrip(r *http.Request) (*http.Response, error) {
	if _this.APIKey != "" {
		r.Header.Add("Authorization", fmt.Sprintf("ApiKey %v", _this.APIKey))
	} else if _this.Username != "" && _this.Password != "" {
		r.SetBasicAuth(_this.Username, _this.Password)
	}

	return http.DefaultTransport.RoundTrip(r)
}

func ProcessConfig(c *Config) (elasticsearch.Config, error) {
	cfg := elasticsearch.Config{}
	if c == nil {
		c = &Config{}
		err := envconfig.Process("elastic", c)
		if err != nil {
			return cfg, err
		}
	}

	c.Default()

	cfg = elasticsearch.Config{
		Addresses: []string{
			c.URL,
		},

		RetryOnStatus: []int{502, 503, 504},
		MaxRetries:    c.MaxRetries,

	}

	if tracing.Enabled {
		cfg.Transport = &tracing.WrapTransport{}
	} else {
		cfg.Transport = &Transport{}
	}

	return cfg, nil
}

// Default - Set a default connection
func (_this *Config) Default() {
	if _this.URL == "" {
		_this.URL = DefaultElasticURL
	}

	if _this.MaxRetries <= 0 {
		_this.MaxRetries = DefaultMaxRetries
	}

	if _this.RetryAfter <= 0 {
		_this.RetryAfter = DefaultRetryAfter
	}
}

// NewElasticClient
func NewElasticClient(c *Config) (*elasticsearch.Client, error) {
	cfg, err := ProcessConfig(c)
	if err != nil {
		return nil, fmt.Errorf("create elasticsearch client has error: %v", err)
	}

	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("create elasticsearch client has error: %v", err)
	}

	retries := 0
	ctx := context.Background()

	for {
		_, err := client.Ping(client.Ping.WithContext(ctx))
		if err == nil {
			break
		}
		if retries >= c.MaxRetries {
			return nil, err
		}
		retries++
		time.Sleep(c.RetryAfter)
	}

	return client, nil
}