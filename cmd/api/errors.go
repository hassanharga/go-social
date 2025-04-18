package main

import (
	"github/hassanharga/go-social/utils"
	"net/http"
)

func (app *application) internalServerError(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Error("internal error", "method", r.Method, "path", r.URL.Path, "error", err.Error())

	utils.WriteJsonError(w, http.StatusInternalServerError, "the server encountered a problem")
}

func (app *application) badRequestError(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Warn("bad request", "method", r.Method, "path", r.URL.Path, "error", err.Error())

	utils.WriteJsonError(w, http.StatusBadRequest, err.Error())
}

func (app *application) notFoundError(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Warn("not found error", "method", r.Method, "path", r.URL.Path, "error", err.Error())

	utils.WriteJsonError(w, http.StatusNotFound, "not found")
}

func (app *application) conflictError(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Error("conflict response", "method", r.Method, "path", r.URL.Path, "error", err.Error())
	utils.WriteJsonError(w, http.StatusConflict, err.Error())
}

func (app *application) forbiddenError(w http.ResponseWriter, r *http.Request) {
	app.logger.Warn("forbidden", "method", r.Method, "path", r.URL.Path, "error")

	utils.WriteJsonError(w, http.StatusForbidden, "forbidden")
}

func (app *application) unauthorizedError(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Warn("unauthorized error", "method", r.Method, "path", r.URL.Path, "error", err.Error())

	utils.WriteJsonError(w, http.StatusUnauthorized, "unauthorized")
}

func (app *application) unauthorizedBasicError(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Warn("unauthorized basic error", "method", r.Method, "path", r.URL.Path, "error", err.Error())

	w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)

	utils.WriteJsonError(w, http.StatusUnauthorized, "unauthorized")
}

func (app *application) rateLimitExceededError(w http.ResponseWriter, r *http.Request, retryAfter string) {
	app.logger.Warn("rate limit exceeded", "method", r.Method, "path", r.URL.Path)

	w.Header().Set("Retry-After", retryAfter)

	utils.WriteJsonError(w, http.StatusTooManyRequests, "rate limit exceeded, retry after: "+retryAfter)
}
