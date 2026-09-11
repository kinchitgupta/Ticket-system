package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"ticket-system/internal/handlers/respond"
	"ticket-system/internal/middleware"
	"ticket-system/internal/models"
	"ticket-system/internal/store"
)

type TicketHandler struct {
	Store *store.Store
}

func NewTicketHandler(s *store.Store) *TicketHandler {
	return &TicketHandler{Store: s}
}

type createTicketRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

func (h *TicketHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		respond.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req createTicketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Title = strings.TrimSpace(req.Title)
	if req.Title == "" {
		respond.Error(w, http.StatusBadRequest, "title is required")
		return
	}

	now := time.Now().UTC()
	ticket := &models.Ticket{
		UserID:      userID,
		Title:       req.Title,
		Description: req.Description,
		Status:      models.StatusOpen,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	h.Store.CreateTicket(ticket)

	respond.JSON(w, http.StatusCreated, ticket)
}

func (h *TicketHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		respond.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	tickets := h.Store.ListTicketsByUser(userID)
	respond.JSON(w, http.StatusOK, tickets)
}

func (h *TicketHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		respond.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id := r.PathValue("id")
	ticket, err := h.Store.GetTicket(id)
	if err != nil {
		respond.Error(w, http.StatusNotFound, "ticket not found")
		return
	}

	if ticket.UserID != userID {
		// Do not reveal that a ticket exists but belongs to someone else.
		respond.Error(w, http.StatusNotFound, "ticket not found")
		return
	}

	respond.JSON(w, http.StatusOK, ticket)
}

type updateStatusRequest struct {
	Status string `json:"status"`
}

func (h *TicketHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		respond.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id := r.PathValue("id")
	ticket, err := h.Store.GetTicket(id)
	if err != nil {
		respond.Error(w, http.StatusNotFound, "ticket not found")
		return
	}

	if ticket.UserID != userID {
		respond.Error(w, http.StatusNotFound, "ticket not found")
		return
	}

	var req updateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	newStatus := models.TicketStatus(strings.TrimSpace(strings.ToLower(req.Status)))
	if !newStatus.IsValid() {
		respond.Error(w, http.StatusBadRequest, "status must be one of: open, in_progress, closed")
		return
	}

	if ticket.Status == newStatus {
		respond.Error(w, http.StatusBadRequest, "ticket is already in this status")
		return
	}

	if !models.CanTransition(ticket.Status, newStatus) {
		respond.Error(w, http.StatusConflict, "invalid status transition from "+string(ticket.Status)+" to "+string(newStatus))
		return
	}

	ticket.Status = newStatus
	ticket.UpdatedAt = time.Now().UTC()
	h.Store.UpdateTicket(ticket)

	respond.JSON(w, http.StatusOK, ticket)
}
