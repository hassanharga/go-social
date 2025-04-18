package main

import (
	"expvar"
	"github/hassanharga/go-social/internal/auth"
	"github/hassanharga/go-social/internal/db"
	"github/hassanharga/go-social/internal/env"
	"github/hassanharga/go-social/internal/mailer"
	"github/hassanharga/go-social/internal/ratelimiter"
	"github/hassanharga/go-social/internal/store"
	"github/hassanharga/go-social/internal/store/cache"
	"log"
	"log/slog"
	"os"
	"runtime"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
)

func init() {
	// Load environment variables from .env file
	log.Println("loading environment variables")
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}
}

//	@title			Special API
//	@description	This is a sample server social network.
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	API Support
//	@contact.url	http://www.swagger.io/support
//	@contact.email	support@swagger.io

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

//	@BasePath	/v1

// @securityDefinitions.apikey	ApiKeyAuth
// @in							header
// @name						Authorization
// @description				API Key for authorization
func main() {
	config := config{
		addr:        env.GetString("ADDR", ":3000"),
		env:         env.GetString("ENV", "development"),
		version:     env.GetString("VERSION", "1.0.0"),
		apiURL:      env.GetString("API_URL", "localhost:3000"),
		frontendURL: env.GetString("FRONT_URL", "localhost:3001"),
		db: dbConfig{
			addr:         env.GetString("DB_ADDR", "postgres://admin:adminpassword@localhost/social?sslmode=disable"),
			maxOpenConns: env.GetInt("DB_MAX_OPEN_CONNS", 30),
			maxIdleConns: env.GetInt("DB_MAX_IDLE_CONNS", 30),
			maxIdleTime:  env.GetString("DB_MAX_IDLE_TIME", "15m"),
		},
		mail: mailConfig{
			expiry:    time.Hour * 24 * 3, // 3 days,
			fromEmail: env.GetString("FROM_EMAIL", "noreply@localhost"),
			sendGrid: sendGridConfig{
				apiKey: env.GetString("SENDGRID_API_KEY", ""),
			},
			mailTrap: mailTrapConfig{
				apiKey: env.GetString("MAILTRAP_API_KEY", "12121"),
			},
		},
		auth: authConfig{
			basic: basicConfig{
				user:     env.GetString("BASIC_AUTH_USER", "admin"),
				password: env.GetString("BASIC_AUTH_PASSWORD", "adminpassword"),
			},
			jwt: jwtConfig{
				secret: env.GetString("JWT_SECRET", "secret"),
				aud:    env.GetString("JWT_AUD", "goSocial"),
				iss:    env.GetString("JWT_ISS", "goSocial"),
				exp:    time.Hour * 24 * 3, // 3 days
			},
		},
		cache: cacheConfig{
			addr:     env.GetString("REDIS_ADDR", "localhost:6379"),
			password: env.GetString("REDIS_PASSWORD", ""),
			db:       env.GetInt("REDIS_DB", 0),
			enabled:  env.GetBool("REDIS_ENABLED", true),
		},
		rateLimiter: ratelimiter.Config{
			RequestsPerTimeFrame: env.GetInt("REQUESTS_PER_TIME_FRAME", 100),
			TimeFrame:            time.Second * 5,
			Enabled:              env.GetBool("RATE_LIMITER_ENABLED", true),
		},
	}

	// initialize the logger
	// logger := zap.Must(zap.NewProduction()).Sugar()
	// defer logger.Sync() // flushes buffer, if any
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{}))
	slog.SetDefault(logger)

	// initialize the database connection
	logger.Info("Connecting to the database")

	db, err := db.New(
		config.db.addr,
		config.db.maxOpenConns,
		config.db.maxIdleConns,
		config.db.maxIdleTime,
	)
	if err != nil {
		logger.Error("failed to connect to the database", "error", err)
		os.Exit(1)
		// log.Fatal(err)
	}
	defer db.Close()

	logger.Info("Connected to the database")

	// initialize the mailer
	logger.Info("initialize to the mailer")
	// mailer := mailer.NewSendgrid(config.mail.sendGrid.apiKey, config.mail.fromEmail)
	mailer, err := mailer.NewMailTrapClient(config.mail.mailTrap.apiKey, config.mail.fromEmail)
	if err != nil {
		logger.Error("failed to connect to the mailer", "error", err)
		os.Exit(1)
		// log.Fatal(err)
	}

	// initialize the JWT authenticator
	jwtConfig := auth.NewJwtConfig(config.auth.jwt.secret, config.auth.jwt.aud, config.auth.jwt.iss)

	// initialize the store
	store := store.NewStorage(db)

	// init cache
	logger.Info("init cache")
	var rdb *redis.Client
	if config.cache.enabled {
		rdb = cache.NewRedisClient(config.cache.addr, config.cache.password, config.cache.db)
		logger.Info("Connected to the cache database")
	}

	// init the rate limiter
	rateLimiter := ratelimiter.NewFixedWindowLimiter(
		config.rateLimiter.RequestsPerTimeFrame,
		config.rateLimiter.TimeFrame,
	)

	cacheStorage := cache.NewRedisStorage(rdb)

	// metric collection
	expvar.NewString("version").Set(config.version)
	expvar.Publish("database", expvar.Func(func() any {
		return db.Stats()
	}))
	expvar.Publish("goroutines", expvar.Func(func() any {
		return runtime.NumGoroutine()
	}))

	app := &application{
		config:        config,
		store:         store,
		logger:        logger,
		mailer:        mailer,
		authenticator: jwtConfig,
		cacheStorage:  cacheStorage,
		rateLimiter:   rateLimiter,
	}

	// initialize the server mux
	if err := app.run(); err != nil {
		logger.Error("server error", "error", err)
		log.Fatal(err)
	}
}
