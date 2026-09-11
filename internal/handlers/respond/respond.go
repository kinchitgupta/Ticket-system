package respond

import (
	"encoding/json"
	"net/http"
)

func JSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if payload != nil {
		_ = json.NewEncoder(w).Encode(payload)
	}
}

type errorBody struct {
	Error string `json:"error"`
}

func Error(w http.ResponseWriter, status int, message string) {
	JSON(w, status, errorBody{Error: message})
}