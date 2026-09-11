package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"ticket-system/internal/handlers"
	"ticket-system/internal/middleware"
	"ticket-system/internal/store"
)

func main() {
	s := store.New()
	authHandler := handlers.NewAuthHandler(s)
	ticketHandler := handlers.NewTicketHandler(s)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	mux.HandleFunc("POST /auth/register", authHandler.Register)
	mux.HandleFunc("POST /auth/login", authHandler.Login)

	mux.HandleFunc("POST /tickets", middleware.RequireAuth(ticketHandler.CreateTicket))
	mux.HandleFunc("GET /tickets", middleware.RequireAuth(ticketHandler.ListTickets))
	mux.HandleFunc("GET /tickets/{id}", middleware.RequireAuth(ticketHandler.GetTicket))
	mux.HandleFunc("PATCH /tickets/{id}/status", middleware.RequireAuth(ticketHandler.UpdateTicketStatus))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("ticket-system listening on port %s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}
