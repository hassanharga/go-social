package main

import (
	"fmt"
	"github/hassanharga/go-social/internal/auth"
	"github/hassanharga/go-social/internal/mailer"
	"github/hassanharga/go-social/internal/store"
	"github/hassanharga/go-social/utils"
	"net/http"
	"time"

	"github/hassanharga/go-social/docs" // this is required for swagger

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
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

type mailTrapConfig struct {
	apiKey string
}

type mailConfig struct {
	expiry    time.Duration
	fromEmail string
	sendGrid  sendGridConfig
	mailTrap  mailTrapConfig
}

type basicConfig struct {
	user     string
	password string
}

type jwtConfig struct {
	secret string
	aud    string
	iss    string
	exp    time.Duration
}

type authConfig struct {
	basic basicConfig
	jwt   jwtConfig
}

type config struct {
	addr        string
	db          dbConfig
	env         string
	version     string
	apiURL      string
	frontendURL string
	mail        mailConfig
	auth        authConfig
}

type application struct {
	config
	store         store.Storage
	logger        *zap.SugaredLogger
	mailer        mailer.Client
	authenticator auth.Authenticator
}

// initialize the server chi and create routes
func (app *application) mount() http.Handler {
	r := chi.NewRouter()

	// Basic CORS
	// for more ideas, see: https://developer.github.com/v3/#cross-origin-resource-sharing
	r.Use(cors.Handler(cors.Options{
		// AllowedOrigins:   []string{"https://foo.com"}, // Use this to allow specific origin hosts
		AllowedOrigins: []string{"https://*", "http://*"},
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

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

		// check health
		r.With(app.basicMiddleware()).Get("/health", app.healthCheckHandler)

		// swagger
		docsUrl := fmt.Sprintf("%s/swagger/doc.json", app.config.addr)
		r.Get("/swagger/*", httpSwagger.Handler(
			httpSwagger.URL(docsUrl), //The url pointing to API definition
		))

		// post routers
		r.Route("/posts", func(r chi.Router) {
			r.Use(app.authTokenMiddleware)

			r.Post("/", app.createPostHandler)
			// r.Get("/", app.getPostsHandler)
			r.Route("/{id}", func(r chi.Router) {
				// post middleware
				r.Use(app.postContextMiddleware)

				r.Get("/", app.getPostHandler)
				r.Patch("/", app.checkPostOwnership(store.MODERATOR, app.updatePostHandler))
				r.Delete("/", app.checkPostOwnership(store.ADMIN, app.deletePostHandler))
				r.Post("/comments", app.createCommentHandler)
			})
		})
		// user routers
		r.Route("/users", func(r chi.Router) {
			// activate user
			r.Put("/activate/{token}", app.activateUserHandler)

			r.Route("/{id}", func(r chi.Router) {
				// user middleware
				r.Use(app.authTokenMiddleware)

				r.Get("/", app.getUserHandler)
				r.Put("/follow", app.followUserHandler)
				r.Put("/unfollow", app.unfollowUserHandler)
			})

			// user feed
			r.Group(func(r chi.Router) {
				r.Use(app.authTokenMiddleware)
				r.Get("/feed", app.getUserFeedHandler)
			})

		})

		// auth routers
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", app.registerUserHandler)
			r.Post("/token", app.createTokenHandler)
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
