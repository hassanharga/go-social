package main

import (
	"github/hassanharga/go-social/internal/auth"
	"github/hassanharga/go-social/internal/db"
	"github/hassanharga/go-social/internal/env"
	"github/hassanharga/go-social/internal/mailer"
	"github/hassanharga/go-social/internal/store"
	"log"
	"time"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
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
	}

	// initialize the logger
	logger := zap.Must(zap.NewProduction()).Sugar()
	defer logger.Sync() // flushes buffer, if any

	// initialize the database connection
	logger.Info("Connecting to the database")
	db, err := db.New(
		config.db.addr,
		config.db.maxOpenConns,
		config.db.maxIdleConns,
		config.db.maxIdleTime,
	)
	if err != nil {
		logger.Fatal(err)
	}
	defer db.Close()

	logger.Info("Connected to the database")

	// initialize the mailer
	logger.Info("initialize to the mailer")
	// mailer := mailer.NewSendgrid(config.mail.sendGrid.apiKey, config.mail.fromEmail)
	mailer, err := mailer.NewMailTrapClient(config.mail.mailTrap.apiKey, config.mail.fromEmail)
	if err != nil {
		logger.Fatal(err)
	}

	// initialize the JWT authenticator
	jwtConfig := auth.NewJwtConfig(config.auth.jwt.secret, config.auth.jwt.aud, config.auth.jwt.iss)

	// initialize the store
	store := store.NewStorage(db)

	app := &application{
		config:        config,
		store:         store,
		logger:        logger,
		mailer:        mailer,
		authenticator: jwtConfig,
	}

	// initialize the server mux
	app.run()
}
