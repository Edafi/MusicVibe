package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/Edafi/MusicVibe/middleware"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
)

type UploadHandler struct {
	DB          *sql.DB
	MinioClient *minio.Client
}

func saveTempFile(file multipart.File, filename string) (string, error) {
	tmpfile, err := os.CreateTemp("", filename)
	if err != nil {
		return "", err
	}
	defer tmpfile.Close()

	_, err = io.Copy(tmpfile, file)
	if err != nil {
		return "", err
	}

	return tmpfile.Name(), nil
}

func getAudioDuration(filePath string) (int, error) {
	cmd := exec.Command("ffprobe", "-v", "quiet", "-print_format", "json", "-show_format", filePath)
	out, err := cmd.Output()
	if err != nil {
		return 0, err
	}
	var data struct {
		Format struct {
			Duration string `json:"duration"`
		} `json:"format"`
	}
	if err := json.Unmarshal(out, &data); err != nil {
		return 0, err
	}
	seconds, _ := strconv.ParseFloat(data.Format.Duration, 64)
	return int(seconds), nil
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

	_, err = handler.DB.Exec(`INSERT IGNORE INTO musician_genre (musician_id, genre_id)
    	VALUES (?, ?)`, musicianID, genreID)
	if err != nil {
		log.Println("UploadAlbum: failed to insert into musician_genre:", err)
	}

	albumID := uuid.New().String()
	bucketName := "music"

	// Путь к обложке: musician_{id}/cover/album_{id}.jpg
	coverObject := fmt.Sprintf("musician_%s/cover/album_%s", musicianID, albumID)
	coverPath, err := uploadToMinIO(handler.MinioClient, bucketName, coverObject, coverFile, coverHeader.Size, coverHeader.Header.Get("Content-Type"))
	if err != nil {
		log.Println("UploadAlbum: ", err)
		http.Error(response, "Failed to upload cover", http.StatusInternalServerError)
		return
	}

	releaseDate := time.Now()

	// Сохраняем альбом
	_, err = handler.DB.Exec(`
		INSERT INTO album (id, musician_id, title, release_date, cover_path, genre_id, description, title_lower)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		albumID, musicianID, albumTitle, releaseDate, coverPath, genreID, albumDescription, strings.ToLower(albumTitle))
	if err != nil {
		log.Println("UploadAlbum: ", err)
		http.Error(response, "Failed to insert album", http.StatusInternalServerError)
		return
	}

	// Обрабатываем треки
	form := request.MultipartForm

	trackTitles := form.Value["trackTitles[]"]
	trackFiles := form.File["trackFiles[]"]

	log.Println("trackTitles:", trackTitles)
	log.Println("trackFiles:", trackFiles)

	if len(trackTitles) != len(trackFiles) {
		log.Println("Количество названий треков не совпадает с количеством файлов")
		http.Error(response, "Invalid track data", http.StatusBadRequest)
		return
	}

	for i := range trackTitles {
		title := trackTitles[i]
		audioFileHeader := trackFiles[i]

		audioFile, err := audioFileHeader.Open()
		if err != nil {
			log.Println("Failed to open audio:", err)
			continue
		}
		defer audioFile.Close()

		trackID := uuid.New().String()

		objectName := fmt.Sprintf("musician_%s/tracks/track_%s", musicianID, trackID)
		audioPath, err := uploadToMinIO(handler.MinioClient, bucketName, objectName, audioFile, audioFileHeader.Size, audioFileHeader.Header.Get("Content-Type"))
		if err != nil {
			log.Println("Failed to upload audio:", err)
			continue
		}

		audioFile.Seek(0, 0)
		tmpPath, err := saveTempFile(audioFile, "audio_*.mp3")
		if err != nil {
			log.Println("Failed to save temp audio file:", err)
			continue
		}
		defer os.Remove(tmpPath) // удаляем временный файл позже

		duration, err := getAudioDuration(tmpPath)
		if err != nil {
			log.Println("Failed to get duration of the track:", err)
			continue
		}

		titleLower := strings.ToLower(title)
		_, err = handler.DB.Exec(`
			INSERT INTO track (id, title, album_id, musician_id, file_path, genre_id, duration, stream_count, visibility, title_lower)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			trackID, title, albumID, musicianID, audioPath, genreID, duration, 0, "public", titleLower,
		)
		if err != nil {
			log.Println("Failed to insert track:", err)
			continue
		}
	}

	response.WriteHeader(http.StatusCreated)
	json.NewEncoder(response).Encode(map[string]string{"message": "Album uploaded successfully", "albumId": albumID})
}
