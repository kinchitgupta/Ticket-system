package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"ticket-system/internal/auth"
	"ticket-system/internal/handlers"
	"ticket-system/internal/middleware"
	"ticket-system/internal/store"
)

func main() {
	s := store.New()

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET environment variable is required")
	}
	tokens := auth.NewTokenIssuer(jwtSecret)

	authHandler := handlers.NewAuthHandler(s, tokens)
	ticketHandler := handlers.NewTicketHandler(s)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Ticket system API is running. See /health for status."))
	})

	mux.HandleFunc("POST /auth/register", authHandler.Register)
	mux.HandleFunc("POST /auth/login", authHandler.Login)

	mux.Handle("POST /tickets", middleware.RequireAuth(tokens)(http.HandlerFunc(ticketHandler.Create)))
	mux.Handle("GET /tickets", middleware.RequireAuth(tokens)(http.HandlerFunc(ticketHandler.List)))
	mux.Handle("GET /tickets/{id}", middleware.RequireAuth(tokens)(http.HandlerFunc(ticketHandler.Get)))
	mux.Handle("PATCH /tickets/{id}/status", middleware.RequireAuth(tokens)(http.HandlerFunc(ticketHandler.UpdateStatus)))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("ticket-system listening on port %s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}