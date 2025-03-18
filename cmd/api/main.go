package main

import "github/hassanharga/go-api/internal/env"

func main() {
	app := &application{
		config: config{addr: env.GetString("ADDR", ":8080")},
	}

	// initialize the server mux
	app.run()
}
