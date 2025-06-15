package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"

	"github.com/Edafi/MusicVibe/middleware"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
)

type UploadHandler struct {
	DB          *sql.DB
	MinioClient *minio.Client
}

func uploadToMinIO(client *minio.Client, bucketName, objectName string, file multipart.File, fileSize int64, contentType string) (string, error) {
	_, err := client.PutObject(context.Background(), bucketName, objectName, file, fileSize, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		log.Fatal("uploadToMinIO: ", err)
		return "", err
	}
	return fmt.Sprintf("/%s/%s", bucketName, objectName), nil
}

func (handler *UploadHandler) UploadAlbum(response http.ResponseWriter, request *http.Request) {
	userID, ok := request.Context().Value(middleware.ContextUserIDKey).(string)
	if !ok || userID == "" {
		http.Error(response, "Unauthorized", http.StatusUnauthorized)
		return
	}

	err := request.ParseMultipartForm(50 << 20) // 50MB
	if err != nil {
		log.Println("UploadAlbum: ", err)
		http.Error(response, "Cannot parse multipart form", http.StatusBadRequest)
		return
	}

	fmt.Println("Content-Type:", request.Header.Get("Content-Type"))

	// Парсим поля
	albumTitle := request.FormValue("albumTitle")
	albumDescription := request.FormValue("albumDescription")
	genreName := request.FormValue("genre")

	// Обложка
	coverFile, coverHeader, err := request.FormFile("cover")
	if err != nil {
		log.Println("UploadAlbum: ", err)
		log.Println("Content-Type:", request.Header.Get("Content-Type"))
		http.Error(response, "Missing cover file", http.StatusBadRequest)
		return
	}
	defer coverFile.Close()

	log.Println("Album:", albumTitle)
	log.Println("Genre:", genreName)
	log.Println("Cover filename:", coverHeader.Filename)
	log.Println("user_id:", userID)

	// Genre
	var genreID int
	err = handler.DB.QueryRow(`SELECT id FROM genre WHERE name = ?`, genreName).Scan(&genreID)
	if err != nil {
		log.Print("UploadAlbum: ", err)
		http.Error(response, "Invalid genre", http.StatusBadRequest)
		return
	}

	// Musician
	var musicianID string
	err = handler.DB.QueryRow(`SELECT id FROM musician WHERE user_id = ?`, userID).Scan(&musicianID)
	if err != nil {
		log.Println("UploadAlbum: ", err)
		http.Error(response, "Musician not found", http.StatusBadRequest)
		return
	}

	albumID := uuid.New().String()
	bucketName := "music"

	// Путь к обложке: musician_{id}/cover/album_{id}.jpg
	coverObject := fmt.Sprintf("musician_%s/cover/album_%s_%s", musicianID, albumID, coverHeader.Filename)
	coverPath, err := uploadToMinIO(handler.MinioClient, bucketName, coverObject, coverFile, coverHeader.Size, coverHeader.Header.Get("Content-Type"))
	if err != nil {
		log.Println("UploadAlbum: ", err)
		http.Error(response, "Failed to upload cover", http.StatusInternalServerError)
		return
	}

	// Сохраняем альбом
	_, err = handler.DB.Exec(`
		INSERT INTO album (id, musician_id, title, cover_path, genre_id, description)
		VALUES (?, ?, ?, ?, ?, ?)`,
		albumID, musicianID, albumTitle, coverPath, genreID, albumDescription)
	if err != nil {
		log.Println("UploadAlbum: ", err)
		http.Error(response, "Failed to insert album", http.StatusInternalServerError)
		return
	}

	// Обрабатываем треки
	form := request.MultipartForm
	trackTitles := form.Value["trackTitle"]
	trackFiles := form.File["trackAudio"]

	for i := range trackTitles {
		title := trackTitles[i]
		audioFileHeader := trackFiles[i]

		audioFile, err := audioFileHeader.Open()
		if err != nil {
			log.Println("UploadAlbum: ", err)
			log.Println("Failed to open audio:", err)
			continue
		}
		defer audioFile.Close()

		trackID := uuid.New().String()

		// Путь к треку: musician_{id}/tracks/track_{id}_название.mp3
		objectName := fmt.Sprintf("musician_%s/tracks/track_%s_%s", musicianID, trackID, audioFileHeader.Filename)
		audioPath, err := uploadToMinIO(handler.MinioClient, bucketName, objectName, audioFile, audioFileHeader.Size, audioFileHeader.Header.Get("Content-Type"))
		if err != nil {
			log.Println("Failed to upload audio:", err)
			continue
		}

		_, err = handler.DB.Exec(`
			INSERT INTO track (id, title, album_id, musician_id, file_path, genre_id)
			VALUES (?, ?, ?, ?, ?, ?)`,
			trackID, title, albumID, musicianID, audioPath, genreID)
		if err != nil {
			log.Println("Failed to insert track:", err)
			continue
		}
	}

	response.WriteHeader(http.StatusCreated)
	json.NewEncoder(response).Encode(map[string]string{"message": "Album uploaded successfully", "albumId": albumID})
}
