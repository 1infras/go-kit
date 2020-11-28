package tracing

import (
	"fmt"
	"net/http"

	"go.elastic.co/apm/module/apmelasticsearch"
)

type WrapTransport struct {
	APIKey   string
	Username string
	Password string
}

func (_this *WrapTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	if _this.APIKey != "" {
		r.Header.Add("Authorization", fmt.Sprintf("ApiKey %v", _this.APIKey))
	} else if _this.Username != "" && _this.Password != "" {
		r.SetBasicAuth(_this.Username, _this.Password)
	}

	return apmelasticsearch.WrapRoundTripper(http.DefaultTransport).RoundTrip(r)
}
