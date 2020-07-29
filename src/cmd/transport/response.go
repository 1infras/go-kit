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
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	b, _ := json.Marshal(message)
	w.Write(b)
}

// Return HTTP 400
func BadRequest(w http.ResponseWriter, message map[string]interface{}) {
	w.WriteHeader(http.StatusBadRequest)
	w.Header().Set("Content-Type", "application/json")
	b, _ := json.Marshal(message)
	w.Write(b)
}

// Return HTTP 500 with OKJson
func BadRequestJson(w http.ResponseWriter, message interface{}) {
	w.WriteHeader(http.StatusBadRequest)
	w.Header().Set("Content-Type", "application/json")
	b, _ := json.Marshal(message)
	w.Write(b)
}

// Return HTTP 500
func InternalServerError(w http.ResponseWriter) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Header().Set("Content-Type", "application/json")
	b, _ := json.Marshal(map[string]interface{}{
		"Error": "Internal Server Error",
	})
	w.Write(b)
}

// Return HTTP 500 with OKJson
func InternalServerErrorJson(w http.ResponseWriter, message interface{}) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Header().Set("Content-Type", "application/json")
	b, _ := json.Marshal(message)
	w.Write(b)
}
