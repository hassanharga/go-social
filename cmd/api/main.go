package main

import (
	"github/hassanharga/go-api/internal/env"
	"github/hassanharga/go-api/internal/store"
)

func main() {
	store := store.NewStorage(nil)
	app := &application{
		config: config{addr: env.GetString("ADDR", ":8080")},
		store:  store,
	}

	// initialize the server mux
	app.run()
}
