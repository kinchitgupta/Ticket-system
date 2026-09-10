package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"ticket-system/internal/auth"
	"ticket-system/internal/handlers/respond"
	"ticket-system/internal/models"
	"ticket-system/internal/store"
)

type AuthHandler struct {
	Store  *store.Store
	Tokens *auth.TokenIssuer
}

func NewAuthHandler(s *store.Store, t *auth.TokenIssuer) *AuthHandler {
	return &AuthHandler{Store: s, Tokens: t}
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type authResponse struct {
	Token string     `json:"token"`
	User  publicUser `json:"user"`
}

type publicUser struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if req.Email == "" || !strings.Contains(req.Email, "@") {
		respond.Error(w, http.StatusBadRequest, "a valid email is required")
		return
	}
	if len(req.Password) < 6 {
		respond.Error(w, http.StatusBadRequest, "password must be at least 6 characters")
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "failed to process password")
		return
	}

	user := &models.User{
		Email:        req.Email,
		PasswordHash: hash,
		CreatedAt:    time.Now().UTC(),
	}

	if err := h.Store.CreateUser(user); err != nil {
		if err == store.ErrUserExists {
			respond.Error(w, http.StatusConflict, "a user with this email already exists")
			return
		}
		respond.Error(w, http.StatusInternalServerError, "failed to create user")
		return
	}

	token, err := h.Tokens.GenerateToken(user.ID, user.Email)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	respond.JSON(w, http.StatusCreated, authResponse{
		Token: token,
		User:  publicUser{ID: user.ID, Email: user.Email},
	})
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	user, err := h.Store.GetUserByEmail(req.Email)
	if err != nil {
		respond.Error(w, http.StatusUnauthorized, "invalid email or password")
		return
	}

	if !auth.CheckPassword(user.PasswordHash, req.Password) {
		respond.Error(w, http.StatusUnauthorized, "invalid email or password")
		return
	}

	token, err := h.Tokens.GenerateToken(user.ID, user.Email)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	respond.JSON(w, http.StatusOK, authResponse{
		Token: token,
		User:  publicUser{ID: user.ID, Email: user.Email},
	})
}
