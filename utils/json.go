package utils

import (
	"encoding/json"
	"net/http"
)

func WriteJson(w http.ResponseWriter, status int, data any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(data)
}

func ReadJson(w http.ResponseWriter, r *http.Request, data any) error {
	// set maximum bytes to read from body
	maxBytes := 1_048_578                                    // 1 MB
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes)) // limit the size of the request body
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	decoder.DisallowUnknownFields() // disallow unknown fields in the request body

	return decoder.Decode(data)
}

func WriteJsonError(w http.ResponseWriter, status int, message string) error {
	type ErrorResponse struct {
		Message string `json:"message"`
	}
	data := ErrorResponse{Message: message}
	return WriteJson(w, status, &data)
}
