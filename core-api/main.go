package main

import (
	"context"
	"os"
)

func main() {
	cfg, err := LoadAppConfig()
	if err != nil {
		panic(err)
	}

	app, err := InitServerFromConfig(context.Background(), cfg)
	if err != nil {
		panic(err)
	}
	address := os.Getenv("HTTP_ADDR")
	if address == "" {
		address = ":8080"
	}
	if err := app.Start(address); err != nil {
		panic(err)
	}
}
