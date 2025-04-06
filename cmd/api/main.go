package main

import (
	"github/hassanharga/go-social/internal/db"
	"github/hassanharga/go-social/internal/env"
	"github/hassanharga/go-social/internal/store"
	"log"

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

//	@securityDefinitions.apikey	ApiKeyAuth
//	@in							header
//	@name						Authorization
//	@description				API Key for authorization
func main() {
	config := config{
		addr:    env.GetString("ADDR", ":3000"),
		env:     env.GetString("ENV", "development"),
		version: env.GetString("VERSION", "1.0.0"),
		apiURL:  env.GetString("API_URL", "localhost:3000"),
		db: dbConfig{
			addr:         env.GetString("DB_ADDR", "postgres://admin:adminpassword@localhost/social?sslmode=disable"),
			maxOpenConns: env.GetInt("DB_MAX_OPEN_CONNS", 30),
			maxIdleConns: env.GetInt("DB_MAX_IDLE_CONNS", 30),
			maxIdleTime:  env.GetString("DB_MAX_IDLE_TIME", "15m"),
		},
	}

	db, err := db.New(
		config.db.addr,
		config.db.maxOpenConns,
		config.db.maxIdleConns,
		config.db.maxIdleTime,
	)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	log.Println("Connected to the database")

	store := store.NewStorage(db)

	app := &application{
		config: config,
		store:  store,
	}

	// initialize the server mux
	app.run()
}
