package handlers

import (
	"context"
	"fmt"
	"github.com/cloudinary/cloudinary-go/v2/api"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/google/uuid"
	"io"
	"log"
	"net/http"
	"os"
)

const (
	MAX_IMAGE_UPLOAD_RETRIES = 3
)

func (h *Handler) UploadImageFileHandler(w http.ResponseWriter, r *http.Request) {

	imageKey := "image_file"
	file, fileHeader, err := r.FormFile("image_file")
	if err != nil {
		writeJSONError(w, fmt.Sprintf("failed to extract file from multipart/form-data with key %s", imageKey), http.StatusBadRequest)
		return
	}
	defer file.Close()

	log.Println("Uploading image file...")
	log.Println("Uploaded file name: ", fileHeader.Filename)
	log.Println("Uploaded file size: ", fileHeader.Size)
	log.Println("MIME header: ", fileHeader.Header)

	// ./uploads/filename.ext
	uniqueFileName := fmt.Sprintf("%s-%s", fileHeader.Filename, uuid.New().String())
	uploadedFileDestination := fmt.Sprintf("./uploads/%s", uniqueFileName)
	uploadedFile, err := os.Create(uploadedFileDestination)
	if err != nil {
		log.Printf("failed to create uploaded file destination %v\n", err)
		writeJSONError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	_, err = io.Copy(uploadedFile, file)
	if err != nil {
		log.Printf("failed to copy file contents to uploaded file destination: %v\n", err)
		uploadedFile.Close()
		writeJSONError(w, "internal server error", http.StatusInternalServerError)
		return
	}
	_ = uploadedFile.Close()

	newUploadedFile, err := os.Open(uploadedFileDestination)
	if err != nil {
		log.Printf("failed to open uploaded file destination: %v\n", err)
		writeJSONError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	isUploadedFileToCloudinary := false
	var uploadFileToCloudinaryErr error
	var uploadResult *uploader.UploadResult

	for i := 0; i < MAX_IMAGE_UPLOAD_RETRIES; i++ {

		newUploadedFile.Seek(0, 0)

		uploadResult, err = h.cloudinaryClient.Upload.Upload(context.Background(), newUploadedFile, uploader.UploadParams{
			PublicID:       uniqueFileName,
			Overwrite:      api.Bool(true),
			UniqueFilename: api.Bool(false),
		})

		if err != nil {
			uploadFileToCloudinaryErr = err
			log.Printf("failed to upload file to cloudinary, attempt:%d", i+1)
			continue
		}

		isUploadedFileToCloudinary = true
		break
	}

	if !isUploadedFileToCloudinary {
		log.Printf("failed to upload file to cloudinary: %v\n", uploadFileToCloudinaryErr)
		_ = newUploadedFile.Close()
		writeJSONError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	_ = newUploadedFile.Close()
	if err = os.Remove(newUploadedFile.Name()); err != nil {
		log.Printf("failed to remove uploaded file from uploads: %v\n", err)
	}

	type Response struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
		Url     string `json:"url"`
	}

	if err := writeJSON(w, Response{Success: true, Message: "uploaded file successfully", Url: uploadResult.SecureURL}, http.StatusOK); err != nil {
		writeJSONError(w, "internal server error", http.StatusInternalServerError)
		return
	}
}

//30/08/25
//1/09/25
//2/09/25
