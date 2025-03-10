package main

import (
	"fmt"
	"httpserver/internal/api"
	"httpserver/internal/storage/db"
	"httpserver/internal/storage/file"
	"net/http"
	"os"
)

func main() {
	// Initialize database connection
	dbPool, err := db.NewPool(db.Config{
		Host:     os.Getenv("POSTGRES_HOST"),
		Port:     os.Getenv("POSTGRES_PORT"),
		User:     os.Getenv("POSTGRES_USER"),
		Password: os.Getenv("POSTGRES_PASSWORD"),
		DBName:   os.Getenv("POSTGRES_DB"),
	})
	if err != nil {
		fmt.Printf("Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer dbPool.Close()

	// Initialize stores
	dbStore := db.NewStore(dbPool)
	fileStore := file.NewStore("uploads/staging", "uploads/final")

	// Initialize and start server
	server := api.NewServer(dbStore, fileStore)

	httpServer := &http.Server{
		Addr:    ":3333",
		Handler: server.Routes(),
	}

	fmt.Printf("Server starting on http://localhost%s\n", httpServer.Addr)
	if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
		fmt.Printf("Error starting server: %s\n", err)
		os.Exit(1)
	}
}
