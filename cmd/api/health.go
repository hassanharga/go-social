package main

import (
	"github/hassanharga/go-social/utils"
	"net/http"
)

func (app *application) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{
		"status":  "ok",
		"env":     app.config.env,
		"version": app.config.version,
	}
	if err := utils.WriteJson(w, http.StatusOK, data); err != nil {
		app.internalServerError(w, r, err)
	}
}
