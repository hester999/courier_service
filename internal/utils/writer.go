package utils

import (
	"encoding/json"
	"net/http"
)

func WriteError(w http.ResponseWriter, status int, msg string) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(struct {
		Message string `json:"message"`
		Code    int    `json:"code"`
	}{
		Message: msg,
		Code:    status,
	})
}

func WriteJSON(w http.ResponseWriter, status int, data any) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
