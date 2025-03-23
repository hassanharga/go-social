package main

import (
	"github/hassanharga/go-social/utils"
	"log"
	"net/http"
)

// internalServerError is a helper function to handle internal server errors
func (app *application) internalServerError(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("Internal Server Error: %s path: %s error: %s", r.Method, r.URL.Path, err.Error())
	utils.WriteJsonError(w, http.StatusInternalServerError, "Internal Server Error")

}

// badRequestError is a helper function to handle bad request errors
func (app *application) badRequestError(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("Bad Request Error: %s path: %s error: %s", r.Method, r.URL.Path, err.Error())
	utils.WriteJsonError(w, http.StatusBadRequest, err.Error())

}

// notFoundError is a helper function to handle not found errors
func (app *application) notFoundError(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("Not Found Error: %s path: %s error: %s", r.Method, r.URL.Path, err.Error())
	utils.WriteJsonError(w, http.StatusNotFound, "Resource not found")
}
