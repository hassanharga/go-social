package main

import (
	"github/hassanharga/go-social/internal/store"
	"github/hassanharga/go-social/utils"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type dbConfig struct {
	addr         string
	maxOpenConns int
	maxIdleConns int
	maxIdleTime  string
}

type config struct {
	addr    string
	db      dbConfig
	env     string
	version string
}

type application struct {
	config
	store store.Storage
}

// initialize the server chi and create routes
func (app *application) mount() http.Handler {
	r := chi.NewRouter()

	// A good base middleware stack
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	// Set a timeout value on the request context (ctx), that will signal
	// through ctx.Done() that the request has timed out and further
	// processing should be stopped.
	r.Use(middleware.Timeout(60 * time.Second))

	r.Get("/", app.healthCheckHandler)

	r.Route("/v1", func(r chi.Router) {
		// post routers
		r.Route("/posts", func(r chi.Router) {
			r.Post("/", app.createPostHandler)
			// r.Get("/", app.getPostsHandler)
			r.Route("/{id}", func(r chi.Router) {
				// post middleware
				r.Use(app.postContextMiddleware)

				r.Get("/", app.getPostHandler)
				r.Patch("/", app.updatePostHandler)
				r.Delete("/", app.deletePostHandler)
				r.Post("/comments", app.createCommentHandler)
			})
		})
		r.Route("/users", func(r chi.Router) {
			// r.Post("/", app.createUserHandler)
			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", app.getUserHandler)
				// r.Patch("/", app.updateUserHandler)
				// r.Delete("/", app.deleteUserHandler)
			})
		})
		// user routers

	})

	return r
}

// initializes the server chi and starts the HTTP server
func (app *application) run() {
	// initialize the server mux
	mux := app.mount()

	// Create a new HTTP server
	srv := &http.Server{
		Addr:         app.config.addr,
		Handler:      mux,
		WriteTimeout: time.Second * 30,
		ReadTimeout:  time.Second * 10,
		IdleTimeout:  time.Second * 60,
	}

	log.Printf("Starting server on %s", app.config.addr)
	// Start the server and log any errors
	log.Fatal(srv.ListenAndServe())
}

func (app *application) jsonResponse(w http.ResponseWriter, status int, data any) error {
	type envelope struct {
		Data any `json:"data"`
	}
	return utils.WriteJson(w, status, &envelope{Data: data})
}
