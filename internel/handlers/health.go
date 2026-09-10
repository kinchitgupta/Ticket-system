package handlers

import (
	"net/http"

	"ticket-system/internal/handlers/respond"
)

func Health(w http.ResponseWriter, r *http.Request) {
	respond.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
