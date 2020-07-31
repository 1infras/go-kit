package transport

import (
	"encoding/json"
	"net/http"
)

// Return HTTP 200 with OKJson encoding for a mapping
func OK(w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
}

// Return HTTP 200 with OKJson encoding
func OKJson(w http.ResponseWriter, message interface{}) {
	w.Header().Set("Content-Type", "application/json")
	b, _ := json.Marshal(message)
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}

// Return HTTP 400
func BadRequest(w http.ResponseWriter, message map[string]interface{}) {
	w.Header().Set("Content-Type", "application/json")
	b, _ := json.Marshal(message)
	w.WriteHeader(http.StatusBadRequest)
	w.Write(b)
}

// Return HTTP 500 with OKJson
func BadRequestJson(w http.ResponseWriter, message interface{}) {
	w.Header().Set("Content-Type", "application/json")
	b, _ := json.Marshal(message)
	w.WriteHeader(http.StatusBadRequest)
	w.Write(b)
}

// Return HTTP 500
func InternalServerError(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	b, _ := json.Marshal(map[string]interface{}{
		"Error": "Internal Server Error",
	})
	w.WriteHeader(http.StatusInternalServerError)
	w.Write(b)
}

// Return HTTP 500 with OKJson
func InternalServerErrorJson(w http.ResponseWriter, message interface{}) {
	w.Header().Set("Content-Type", "application/json")
	b, _ := json.Marshal(message)
	w.WriteHeader(http.StatusInternalServerError)
	w.Write(b)
}
