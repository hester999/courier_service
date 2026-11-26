package service

import (
	"encoding/json"
	"net/http"

	"avito/pkg"
)

type Handler struct {
	logger pkg.Logger
}

func NewHandler(logger pkg.Logger) *Handler {
	return &Handler{
		logger: logger,
	}
}

func (h *Handler) Ping(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(struct {
		Message string `json:"message"`
	}{Message: "pong"})
	if err != nil {
		h.logger.Error("Error encoding response: %v", err)
		panic(err)
	}
}

func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
}
