package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	// Load variables from a local .env file if one exists (for local
	// development). In production (Render, etc.) real environment
	// variables are set directly, and this is a harmless no-op.
	_ = godotenv.Load()

	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		log.Fatal("MONGODB_URI environment variable is required (see .env.example)")
	}
	dbName := os.Getenv("MONGODB_DB")
	if dbName == "" {
		dbName = "mediasequencer"
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	store, err := NewStore(mongoURI, dbName)
	if err != nil {
		log.Fatalf("failed to connect to MongoDB: %v", err)
	}

	seedCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := store.SeedIfEmpty(seedCtx); err != nil {
		log.Fatalf("failed to seed database: %v", err)
	}

	hub := NewHub()
	api := &API{store: store, hub: hub}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", handleHealth)
	mux.HandleFunc("GET /windows", api.handleListWindows)
	mux.HandleFunc("POST /windows/{id}/media", api.handleAddMedia)
	mux.HandleFunc("POST /sync", api.handleSync)
	mux.HandleFunc("GET /ws", hub.HandleWS)

	addr := ":" + port
	log.Printf("media-sequencer backend listening on %s", addr)
	if err := http.ListenAndServe(addr, withCORS(mux)); err != nil {
		log.Fatal(err)
	}
}
