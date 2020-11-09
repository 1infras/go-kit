package tracing

import (
	"fmt"
	"net/http"

	"go.elastic.co/apm/module/apmelasticsearch"
)

type WrapTransport struct {
	APIKey string `json:"api_key"`
}

func (_this *WrapTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	if _this.APIKey != "" {
		r.Header.Add("Authorization", fmt.Sprintf("ApiKey %v", _this.APIKey))
	}

	return apmelasticsearch.WrapRoundTripper(http.DefaultTransport).RoundTrip(r)
}
