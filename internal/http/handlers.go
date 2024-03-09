package http

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/mehix/gopher-burrows/internal/burrows"
)

func Handler(manager burrows.Manager) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("OK")) })
	mux.HandleFunc("GET /", showStatus(manager))
	mux.HandleFunc("POST /rent", rentBurrow(manager))
	return mux
}

func showStatus(manager burrows.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		burrows := manager.CurrentStatus()

		w.Header().Set("Content-type", "application/json")
		if err := json.NewEncoder(w).Encode(burrows); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func rentBurrow(manager burrows.Manager) http.HandlerFunc {
	type Response struct {
		Burrow burrows.Burrow
		Error  string
	}
	return func(w http.ResponseWriter, r *http.Request) {
		allowedTime, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()

		b, err := manager.Rentout(allowedTime)

		w.Header().Set("Content-type", "application/json")
		if err != nil {
			_ = json.NewEncoder(w).Encode(Response{Error: err.Error()})
			return
		}

		_ = json.NewEncoder(w).Encode(Response{Burrow: b})
	}
}
