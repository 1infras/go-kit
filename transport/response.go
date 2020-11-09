package transport

import (
	"encoding/json"
	"net/http"
)

// OK - Return HTTP 200
func OK(w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
}

// OKJson - Return HTTP 200 with Json encoding
func OKJson(w http.ResponseWriter, message interface{}) {
	w.Header().Set("Content-Type", "application/json")
	b, _ := json.Marshal(message)
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}

// BadRequest - Return HTTP 400
func BadRequest(w http.ResponseWriter, message map[string]interface{}) {
	w.Header().Set("Content-Type", "application/json")
	b, _ := json.Marshal(message)
	w.WriteHeader(http.StatusBadRequest)
	w.Write(b)
}

// BadRequestJson - Return HTTP 500 with Json
func BadRequestJson(w http.ResponseWriter, message interface{}) {
	w.Header().Set("Content-Type", "application/json")
	b, _ := json.Marshal(message)
	w.WriteHeader(http.StatusBadRequest)
	w.Write(b)
}

// InternalServerError - Return HTTP 500
func InternalServerError(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	b, _ := json.Marshal(map[string]interface{}{
		"Error": "Internal Server Error",
	})
	w.WriteHeader(http.StatusInternalServerError)
	w.Write(b)
}

// InternalServerErrorJson - Return HTTP 500 with Json
func InternalServerErrorJson(w http.ResponseWriter, message interface{}) {
	w.Header().Set("Content-Type", "application/json")
	b, _ := json.Marshal(message)
	w.WriteHeader(http.StatusInternalServerError)
	w.Write(b)
}

// CustomJson - Return a HTTP Status Code with Json
func CustomJson(w http.ResponseWriter, statusCode int, message interface{}) {
	w.Header().Set("Content-Type", "application/json")
	b, _ := json.Marshal(message)
	w.WriteHeader(statusCode)
	w.Write(b)
}
