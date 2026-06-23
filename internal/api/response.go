package api

import (
	"encoding/json"
	"net/http"

	"taskbridge/internal/model"
)

func writeError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	json.NewEncoder(w).Encode(model.ErrorResponse{
		Error: msg,
	})
}