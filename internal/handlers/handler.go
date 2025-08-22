package handlers

import (
	"github.com/dhruv15803/go-blog-app/internal/mailer"
	"github.com/dhruv15803/go-blog-app/internal/storage"
	"net/http"
)

type Handler struct {
	storage   *storage.Storage
	mailer    *mailer.Mailer
	clientUrl string
}

func NewHandler(storage *storage.Storage, mailer *mailer.Mailer, clientUrl string) *Handler {
	return &Handler{
		storage:   storage,
		mailer:    mailer,
		clientUrl: clientUrl,
	}
}

func (h *Handler) HealthCheckHandler(w http.ResponseWriter, r *http.Request) {

	type Response struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	}

	if err := writeJSON(w, Response{Success: true, Message: "health OK"}, http.StatusOK); err != nil {
		writeJSONError(w, "internal server error", http.StatusInternalServerError)
	}
}
