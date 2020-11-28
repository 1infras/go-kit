package transport

import (
	"io"
	"net/http"
)

type HealthCheckHandler struct{}

func (*HealthCheckHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	_, _ = io.WriteString(w, `{"status": "ok"}`)
}
