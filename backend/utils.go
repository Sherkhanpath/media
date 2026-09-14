package main

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

// generateID returns a short random hex ID for new media items.
func generateID() string {
	b := make([]byte, 6)
	_, _ = rand.Read(b)
	return "m" + hex.EncodeToString(b)
}

// withCORS allows the frontend (deployed on a different domain, e.g.
// Vercel) to call this API and open the WebSocket connection.
func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
