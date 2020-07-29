package middleware

import (
	"go.elastic.co/apm"
	"go.uber.org/zap"
	"net/http"
	"time"
)

const timeFormat = time.RFC3339

type ZapLoggerMiddleware struct{}

// New Zap Logger Middleware to trace your HTTP request
func NewZapLoggerMiddleware() *ZapLoggerMiddleware {
	return &ZapLoggerMiddleware{}
}

func (m *ZapLoggerMiddleware) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	//Start time
	start := time.Now()
	//Next
	next(rw, r)
	//End time
	end := time.Now()
	//Calculate Latency
	latency := end.Sub(start)
	span, _ := apm.StartSpan(r.Context(), "logger", "custom")
	defer span.End()
	//Log the result
	zap.L().Info(r.URL.Path,
		zap.String("host", r.Host),
		zap.String("method", r.Method),
		zap.String("query", r.URL.RawQuery),
		zap.String("start", start.Format(timeFormat)),
		zap.String("end", end.Format(timeFormat)),
		zap.String("latency", latency.String()))
}
