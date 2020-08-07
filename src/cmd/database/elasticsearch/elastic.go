package elasticsearch

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/spf13/viper"

	"github.com/1infras/go-kit/src/cmd/logger"

	"github.com/olivere/elastic"
	"go.elastic.co/apm/module/apmelasticsearch"
)

const (
	//DefaultElasticURL -
	DefaultElasticURL = "http://localhost:9200"
	//DefaultMaxRetries -
	DefaultMaxRetries = 3
	//DefaultRetryAfter -
	DefaultRetryAfter = 5 * time.Second
)

//Connection - Config connection to ElasticSearch
type Connection struct {
	URL         string        `json:"url"`
	Secure      bool          `json:"secure"`
	APIKey      string        `json:"api_key"`
	EnableSniff bool          `json:"enable_sniff"`
	MaxRetries  int           `json:"max_retries"`
	RetryAfter  time.Duration `json:"retry_after"`
}

//RoundTrip - Wrap RoundTrip with APM ElasticSearch and adding API Key in header to authorization
func (c *Connection) RoundTrip(r *http.Request) (*http.Response, error) {
	if c.Secure {
		r.Header.Add("Authorization", fmt.Sprintf("ApiKey %v", c.APIKey))
	}

	return apmelasticsearch.WrapRoundTripper(http.DefaultTransport).RoundTrip(r)
}

//DefaultConnection - Set a default connection
func (c *Connection) Default() {
	if c.URL == "" {
		c.URL = DefaultElasticURL
	}

	if c.MaxRetries <= 0 {
		c.MaxRetries = DefaultMaxRetries
	}

	if c.RetryAfter <= 0 {
		c.RetryAfter = DefaultRetryAfter
	}
}

//ConnectionWithViper - Read connection with viper
func ConnectionWithViper() *Connection {
	c := &Connection{
		URL:         viper.GetString("elasticsearch.url"),
		Secure:      viper.GetBool("elasticsearch.secure"),
		APIKey:      viper.GetString("elasticsearch.api_key"),
		EnableSniff: viper.GetBool("elasticsearch.enable_sniff"),
		MaxRetries:  viper.GetInt("elasticsearch.max_retries"),
		RetryAfter:  viper.GetDuration("elasticsearch.retry_after"),
	}

	c.Default()

	return c
}

//NewElasticClient - New a elastic client with connection configured
func NewElasticClient(c *Connection) (*elastic.Client, error) {
	if c.URL == "" {
		return nil, fmt.Errorf("url of elasticsearch must not be empty")
	}

	if c.Secure && c.APIKey == "" {
		return nil, fmt.Errorf("secure was enabled but api_key is empty")
	}

	client, err := elastic.NewClient(
		elastic.SetHttpClient(&http.Client{Transport: c}),
		elastic.SetURL(c.URL),
		elastic.SetSniff(c.EnableSniff),
	)

	if err != nil {
		return nil, err
	}

	ctx := context.Background()

	var (
		info    *elastic.PingResult
		code    int
		retries = 0
	)

	for {
		info, code, err = client.Ping(c.URL).Do(ctx)
		if err == nil {
			break
		}
		if retries >= c.MaxRetries {
			return nil, err
		}
		retries++
		time.Sleep(c.RetryAfter)
	}

	logger.Infof("Elasticsearch returned with code %d and version %s\n", code, info.Version.Number)
	return client, nil
}

//NewDefaultElasticClient - New a default elastic client with default URL
func NewDefaultElasticClient() (*elastic.Client, error) {
	c := &Connection{}
	//Set default connection
	c.Default()
	return NewElasticClient(c)
}
