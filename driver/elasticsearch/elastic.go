package elasticsearch

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/1infras/go-kit/tracing"

	"github.com/olivere/elastic"
)

const (
	// DefaultElasticURL -
	DefaultElasticURL = "http://localhost:9200"
	// DefaultMaxRetries -
	DefaultMaxRetries = 3
	// DefaultRetryAfter -
	DefaultRetryAfter = 5 * time.Second
)

// Connection - Config connection to ElasticSearch
type Connection struct {
	URL         string        `json:"url"`
	Secure      bool          `json:"secure"`
	APIKey      string        `json:"api_key"`
	EnableSniff bool          `json:"enable_sniff"`
	MaxRetries  int           `json:"max_retries"`
	RetryAfter  time.Duration `json:"retry_after"`
}

// RoundTrip - Wrap RoundTrip to adding API Key in header to do authorization
func (_this *Connection) RoundTrip(r *http.Request) (*http.Response, error) {
	if _this.Secure {
		r.Header.Add("Authorization", fmt.Sprintf("ApiKey %v", _this.APIKey))
	}

	return http.DefaultTransport.RoundTrip(r)
}

// Default - Set a default connection
func (_this *Connection) Default() {
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

// NewElasticClient - New a elastic client with connection configured
func NewElasticClient(c *Connection) (*elastic.Client, error) {
	if c.URL == "" {
		return nil, fmt.Errorf("url of elasticsearch must not be empty")
	}

	if c.Secure && c.APIKey == "" {
		return nil, fmt.Errorf("secure was enabled but api_key is empty")
	}

	var (
		client *elastic.Client
		err    error
	)

	if tracing.Enabled {
		client, err = elastic.NewClient(
			elastic.SetHttpClient(&http.Client{Transport: &tracing.WrapTransport{}}),
			elastic.SetURL(c.URL),
			elastic.SetSniff(c.EnableSniff),
		)
	} else {
		client, err = elastic.NewClient(
			elastic.SetHttpClient(&http.Client{Transport: c}),
			elastic.SetURL(c.URL),
			elastic.SetSniff(c.EnableSniff),
		)
	}

	if err != nil {
		return nil, err
	}

	var (
		ctx     = context.Background()
		retries = 0
	)

	for {
		_, _, err = client.Ping(c.URL).Do(ctx)
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

// NewDefaultElasticClient - New a default elastic client with default URL
func NewDefaultElasticClient() (*elastic.Client, error) {
	c := &Connection{}
	// Set default connection
	c.Default()
	return NewElasticClient(c)
}
