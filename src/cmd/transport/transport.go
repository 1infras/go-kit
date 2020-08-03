package transport

import (
	"github.com/1infras/go-kit/src/cmd/logger"
	"github.com/1infras/go-kit/src/cmd/middleware"
	"github.com/gorilla/mux"
	"github.com/urfave/negroni"
	"go.elastic.co/apm/module/apmgorilla"
	"io"
	"net/http"
)

type Transport struct {
	PathPrefix string
	Routes     []Route
}

type Route struct {
	Path       string
	Method     string
	Handler    http.Handler
	Middleware []http.Handler
}

//Add New HTTP Router
func NewRouter(transport Transport) *mux.Router {
	r := mux.NewRouter().StrictSlash(false)
	apmgorilla.Instrument(r)

	//Add route health check
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		io.WriteString(w, `{"status": "ok"}`)
	})

	//Add routes
	for _, t := range transport.Routes {
		n := negroni.New()
		n.Use(middleware.NewZapLoggerMiddleware())

		for _, m := range t.Middleware {
			n.UseHandler(m)
		}

		n.UseHandler(t.Handler)
		r.PathPrefix(transport.PathPrefix).
			Path(t.Path).
			Methods(t.Method).
			Handler(n)
	}

	//Debug print route was created
	r.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		t, err := route.GetPathTemplate()
		if err != nil {
			return err
		}
		logger.Debugf("Route %s was initialized", t)
		return nil
	})

	return r
}
