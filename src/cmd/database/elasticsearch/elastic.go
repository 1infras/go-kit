package elasticsearch

import (
	"context"
	"fmt"
	"net/http"

	"github.com/olivere/elastic"
	"go.elastic.co/apm/module/apmelasticsearch"
)

const DefaultElasticURL = "http://localhost:9200"

type Connection struct {
	URL         string `json:"url"`
	Secure      bool   `json:"secure"`
	APIKey      string `json:"api_key"`
	EnableSniff bool   `json:"sniff"`
}

func (c *Connection) RoundTrip(r *http.Request) (*http.Response, error) {
	if c.Secure {
		r.Header.Add("Authorization", fmt.Sprintf("ApiKey %v", c.APIKey))
	}
	return apmelasticsearch.WrapRoundTripper(http.DefaultTransport).RoundTrip(r)
}

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

	info, code, err := client.Ping(c.URL).Do(ctx)
	if err != nil {
		return nil, err
	}

	logger.Infof("Elasticsearch returned with code %d and version %s\n", code, info.Version.Number)
	return client, nil
}

func NewDefaultElasticClient() (*elastic.Client, error) {
	return NewElasticClient(&Connection{
		URL: DefaultElasticURL,
	})
}
