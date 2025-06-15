package handlers

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/minio/minio-go/v7"
)

type MediaHandler struct {
	MinioClient *minio.Client
	BucketName  string
}

// GET /media/audio/{trackId}
func (h *MediaHandler) ServeAudio(response http.ResponseWriter, request *http.Request) {
	trackID := mux.Vars(request)["trackId"]
	if trackID == "" {
		http.Error(response, "Missing track ID", http.StatusBadRequest)
		return
	}

	// Извлекаем путь из БД (можно также передавать весь путь в параметре)
	var filePath string
	err := h.MinioClient.FGetObject(context.Background(), h.BucketName, fmt.Sprintf("track_%s", trackID), filePath, minio.GetObjectOptions{})
	if err != nil {
		http.Error(response, "Audio not found", http.StatusNotFound)
		return
	}

	obj, err := h.MinioClient.GetObject(context.Background(), h.BucketName, filePath, minio.GetObjectOptions{})
	if err != nil {
		log.Println("ServeAudio: error getting object:", err)
		http.Error(response, "Failed to fetch audio", http.StatusInternalServerError)
		return
	}
	defer obj.Close()

	response.Header().Set("Content-Type", "audio/mpeg")
	io.Copy(response, obj)
}

// GET /media/image/{filename}
func (h *MediaHandler) ServeImage(response http.ResponseWriter, request *http.Request) {
	filename := mux.Vars(request)["filename"]
	if filename == "" {
		http.Error(response, "Missing filename", http.StatusBadRequest)
		return
	}

	objectPath := "music/" + filename // или другой путь в бакете
	obj, err := h.MinioClient.GetObject(context.Background(), h.BucketName, objectPath, minio.GetObjectOptions{})
	if err != nil {
		log.Println("ServeImage: error getting object:", err)
		http.Error(response, "Image not found", http.StatusNotFound)
		return
	}
	defer obj.Close()

	if strings.HasSuffix(filename, ".jpg") {
		response.Header().Set("Content-Type", "image/jpeg")
	} else {
		response.Header().Set("Content-Type", "application/octet-stream")
	}
	io.Copy(response, obj)
}
