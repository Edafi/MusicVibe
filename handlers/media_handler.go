package handlers

import (
	"context"
	"database/sql"
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
	DB          *sql.DB
}

func (h *MediaHandler) ServeAudio(w http.ResponseWriter, r *http.Request) {
	trackID := mux.Vars(r)["trackId"]
	if trackID == "" {
		http.Error(w, "Missing track ID", http.StatusBadRequest)
		return
	}

	// 1. Получаем musician_id из БД
	var musicianID string
	log.Println("Trying to fetch musicianID for track:", trackID)
	err := h.DB.QueryRow("SELECT musician_id FROM track WHERE id = ?", trackID).Scan(&musicianID)
	if err != nil {
		log.Println("ServeAudio: failed to get musician_id:", err)
		http.Error(w, "Track not found", http.StatusNotFound)
		return
	}
	log.Println("Found musicianID:", musicianID)

	// 2. Увеличиваем счетчик прослушиваний
	_, err = h.DB.Exec("UPDATE track SET stream_count = stream_count + 1 WHERE id = ?", trackID)
	if err != nil {
		log.Println("ServeAudio: failed to increment stream count:", err)
		// Не прерываем выполнение, просто логируем ошибку
	}

	// 3. Достаём аудио из MinIO
	objectName := fmt.Sprintf("musician_%s/tracks/track_%s.mp3", musicianID, trackID)
	obj, err := h.MinioClient.GetObject(context.Background(), h.BucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		log.Println("ServeAudio: error getting object:", err)
		http.Error(w, "Failed to fetch audio", http.StatusInternalServerError)
		return
	}
	defer obj.Close()

	stat, err := obj.Stat()
	if err != nil {
		log.Println("ServeAudio: object stat failed:", err)
		http.Error(w, "Audio not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "audio/mpeg")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", stat.Size))
	io.Copy(w, obj)
}

func (h *MediaHandler) ServeImage(w http.ResponseWriter, r *http.Request) {
	filename := mux.Vars(r)["filename"]
	if filename == "" {
		http.Error(w, "Missing filename", http.StatusBadRequest)
		return
	}

	if !strings.HasPrefix(filename, "album_") {
		http.Error(w, "Invalid filename format", http.StatusBadRequest)
		return
	}

	if !strings.HasSuffix(filename, ".jpg") {
		filename += ".jpg"
	}

	albumID := strings.TrimPrefix(strings.TrimSuffix(filename, ".jpg"), "album_")

	var musicianID string
	log.Println("Trying to fetch albumID:", albumID)
	err := h.DB.QueryRow("SELECT musician_id FROM album WHERE id = ?", albumID).Scan(&musicianID)
	if err != nil {
		log.Println("ServeImage: failed to get musician_id for album", albumID, "error:", err)
		http.Error(w, "Album not found", http.StatusNotFound)
		return
	}
	log.Println("Found musicianID:", musicianID)

	objectPath := fmt.Sprintf("musician_%s/cover/%s", musicianID, filename)

	obj, err := h.MinioClient.GetObject(r.Context(), h.BucketName, objectPath, minio.GetObjectOptions{})
	if err != nil {
		log.Println("ServeImage: error getting object:", err)
		http.Error(w, "Image not found", http.StatusNotFound)
		return
	}
	defer obj.Close()

	// Проверяем, что объект существует
	if _, err := obj.Stat(); err != nil {
		log.Println("ServeImage: object stat failed:", err)
		http.Error(w, "Image not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "image/jpeg")
	io.Copy(w, obj)
}
