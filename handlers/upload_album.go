package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"

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
		return "", err
	}
	return fmt.Sprintf("/%s/%s", bucketName, objectName), nil
}

func (handler *UploadHandler) UploadAlbum(response http.ResponseWriter, request *http.Request) {
	err := request.ParseMultipartForm(50 << 20) // 50MB
	if err != nil {
		http.Error(response, "Cannot parse multipart form", http.StatusBadRequest)
		return
	}

	// Парсим описание альбома
	albumTitle := request.FormValue("albumTitle")
	albumDescription := request.FormValue("albumDescription")
	genreName := request.FormValue("genre")
	userID := request.FormValue("userId") // кто загружает

	// Обложка
	coverFile, coverHeader, err := request.FormFile("cover")
	if err != nil {
		http.Error(response, "Missing cover file", http.StatusBadRequest)
		return
	}
	defer coverFile.Close()

	// Найдём genre_id
	var genreID int
	err = handler.DB.QueryRow(`SELECT id FROM genre WHERE name = ?`, genreName).Scan(&genreID)
	if err != nil {
		http.Error(response, "Invalid genre", http.StatusBadRequest)
		return
	}

	// Получим musician_id
	var musicianID string
	err = handler.DB.QueryRow(`SELECT id FROM musician WHERE user_id = ?`, userID).Scan(&musicianID)
	if err != nil {
		http.Error(response, "Musician not found", http.StatusBadRequest)
		return
	}

	albumID := uuid.New().String()
	coverObject := fmt.Sprintf("cover/%s_%s", albumID, coverHeader.Filename)

	coverPath, err := uploadToMinIO(handler.MinioClient, "music", coverObject, coverFile, coverHeader.Size, coverHeader.Header.Get("Content-Type"))
	if err != nil {
		http.Error(response, "Failed to upload cover", http.StatusInternalServerError)
		return
	}

	// Сохраняем альбом
	_, err = handler.DB.Exec(`
		INSERT INTO album (id, musician_id, title, cover_path, genre_id, description)
		VALUES (?, ?, ?, ?, ?, ?)`,
		albumID, musicianID, albumTitle, coverPath, genreID, albumDescription)
	if err != nil {
		http.Error(response, "Failed to insert album", http.StatusInternalServerError)
		return
	}

	// Обрабатываем треки
	form := request.MultipartForm
	trackTitles := form.Value["trackTitle"] // массив названий треков
	trackFiles := form.File["trackAudio"]   // массив файлов

	for i := range trackTitles {
		title := trackTitles[i]
		audioFileHeader := trackFiles[i]

		audioFile, err := audioFileHeader.Open()
		if err != nil {
			log.Println("Failed to open audio:", err)
			continue
		}
		defer audioFile.Close()

		objectName := fmt.Sprintf("track/%s_%s", uuid.New().String(), audioFileHeader.Filename)

		audioPath, err := uploadToMinIO(handler.MinioClient, "music", objectName, audioFile, audioFileHeader.Size, audioFileHeader.Header.Get("Content-Type"))
		if err != nil {
			log.Println("Failed to upload audio:", err)
			continue
		}

		trackID := uuid.New().String()

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
