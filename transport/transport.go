package transport

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/urfave/negroni"

	"github.com/1infras/go-kit/logger"
	"github.com/1infras/go-kit/middleware"
	"github.com/1infras/go-kit/tracing"
)

// Route -
type Route struct {
	Path       string
	Method     string
	Handler    http.Handler
	Middleware []http.Handler
}

// NewRouter -
func NewRouter(pathPrefix string, strictSlash bool, routes []*Route) *mux.Router {
	r := mux.NewRouter().StrictSlash(strictSlash)
	if tracing.Enabled {
		tracing.WrapGorillaMux(r)
	}
	// Add route health check
	r.Handle("/health", &HealthCheckHandler{})

	// Add routes
	for _, t := range routes {
		n := negroni.New()
		n.Use(middleware.NewZapLoggerMiddleware())

		for _, m := range t.Middleware {
			n.UseHandler(m)
		}

		n.UseHandler(t.Handler)
		r.PathPrefix(pathPrefix).
			Path(t.Path).
			Methods(t.Method).
			Handler(n)
	}

	// Validate routes
	_ = r.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		t, err := route.GetPathTemplate()
		if err != nil {
			return err
		}
		logger.Debugf("Route %s was initialized", t)
		return nil
	})

	return r
}
