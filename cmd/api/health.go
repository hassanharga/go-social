package main

import "net/http"

func (app *application) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	// Set the Content-Type header to application/json
	w.Header().Set("Content-Type", "application/json")

	// Write a JSON response with a 200 OK status
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"OK"}`))
}
