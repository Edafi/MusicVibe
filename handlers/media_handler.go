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
	userID, ok := request.Context().Value("userID").(string)
	if !ok || userID == "" {
		http.Error(response, "Unauthorized", http.StatusUnauthorized)
		return
	}

	trackID := mux.Vars(request)["trackId"]
	if trackID == "" {
		http.Error(response, "Missing track ID", http.StatusBadRequest)
		return
	}

	objectName := fmt.Sprintf("track_%s", trackID)

	obj, err := h.MinioClient.GetObject(context.Background(), h.BucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		log.Println("ServeAudio: error getting object:", err)
		http.Error(response, "Failed to fetch audio", http.StatusInternalServerError)
		return
	}
	defer obj.Close()

	// Проверка, существует ли объект
	stat, err := obj.Stat()
	if err != nil {
		log.Println("ServeAudio: object stat failed:", err)
		http.Error(response, "Audio not found", http.StatusNotFound)
		return
	}

	response.Header().Set("Content-Type", "audio/mpeg")
	response.Header().Set("Content-Length", fmt.Sprintf("%d", stat.Size))
	io.Copy(response, obj)
}

// GET /media/image/{filename}
func (h *MediaHandler) ServeImage(response http.ResponseWriter, request *http.Request) {
	userID, ok := request.Context().Value("userID").(string)
	if !ok || userID == "" {
		http.Error(response, "Unauthorized", http.StatusUnauthorized)
		return
	}

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
