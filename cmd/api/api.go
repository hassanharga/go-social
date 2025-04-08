package main

import (
	"fmt"
	"github/hassanharga/go-social/internal/mailer"
	"github/hassanharga/go-social/internal/store"
	"github/hassanharga/go-social/utils"
	"net/http"
	"time"

	"github/hassanharga/go-social/docs" // this is required for swagger

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.uber.org/zap"
)

type dbConfig struct {
	addr         string
	maxOpenConns int
	maxIdleConns int
	maxIdleTime  string
}

type sendGridConfig struct {
	apiKey string
}

type mailConfig struct {
	expiry    time.Duration
	fromEmail string
	sendGrid  sendGridConfig
}

type config struct {
	addr        string
	db          dbConfig
	env         string
	version     string
	apiURL      string
	frontendURL string
	mail        mailConfig
}

type application struct {
	config
	store  store.Storage
	logger *zap.SugaredLogger
	mailer mailer.Client
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

	r.Route("/v1", func(r chi.Router) {

		// swagger
		docsUrl := fmt.Sprintf("%s/swagger/doc.json", app.config.addr)
		r.Get("/swagger/*", httpSwagger.Handler(
			httpSwagger.URL(docsUrl), //The url pointing to API definition
		))

		// check health
		r.Get("/health", app.healthCheckHandler)

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
		// user routers
		r.Route("/users", func(r chi.Router) {
			// activate user
			r.Put("/activate/{token}", app.activateUserHandler)

			r.Route("/{id}", func(r chi.Router) {
				// user middleware
				r.Use(app.userContextMiddleware)

				r.Get("/", app.getUserHandler)
				r.Put("/follow", app.followUserHandler)
				r.Put("/unfollow", app.unfollowUserHandler)
			})

			// user feed
			r.Group(func(r chi.Router) {
				r.Get("/feed", app.getUserFeedHandler)
			})

		})

		// auth routers
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", app.registerUserHandler)
		})
	})

	return r
}

// initializes the server chi and starts the HTTP server
func (app *application) run() {
	// docs
	docs.SwaggerInfo.Version = app.config.version
	docs.SwaggerInfo.Host = app.config.apiURL
	docs.SwaggerInfo.BasePath = "/v1"

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

	app.logger.Infof("Starting server on %s", app.config.addr)
	// Start the server and log any errors
	app.logger.Fatal(srv.ListenAndServe())
}

func (app *application) jsonResponse(w http.ResponseWriter, status int, data any) error {
	type envelope struct {
		Data any `json:"data"`
	}
	return utils.WriteJson(w, status, &envelope{Data: data})
}
