package tracing

import (
	"github.com/gorilla/mux"
	"go.elastic.co/apm/module/apmgorilla"
)

func WrapGorillaMux(r *mux.Router) {
	apmgorilla.Instrument(r)
}
