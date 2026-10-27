package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"fullstack-app/backend/internal/api"
)

func main() {
	config, err := api.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	server := &http.Server{
		Addr:              ":8080",
		Handler:           api.NewServer(config),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Printf("API listening on %s", server.Addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Printf("API server failed: %v", err)
		os.Exit(1)
	}
}
