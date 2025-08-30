package handlers

import (
	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/dhruv15803/go-blog-app/internal/mailer"
	"github.com/dhruv15803/go-blog-app/internal/storage"
	"github.com/redis/go-redis/v9"
	"net/http"
)

type Handler struct {
	storage          *storage.Storage
	mailer           *mailer.Mailer
	redisClient      *redis.Client
	cloudinaryClient *cloudinary.Cloudinary
	clientUrl        string
}

func NewHandler(storage *storage.Storage, mailer *mailer.Mailer, redisClient *redis.Client, cloudinaryClient *cloudinary.Cloudinary, clientUrl string) *Handler {
	return &Handler{
		storage:          storage,
		mailer:           mailer,
		redisClient:      redisClient,
		cloudinaryClient: cloudinaryClient,
		clientUrl:        clientUrl,
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
