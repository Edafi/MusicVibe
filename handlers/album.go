package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path/filepath"

	"github.com/Edafi/MusicVibe/models"
	"github.com/gorilla/mux"
)

type AlbumHandler struct {
	DB *sql.DB
}

func (handler *AlbumHandler) GetAlbum(response http.ResponseWriter, request *http.Request) {
	albumID := mux.Vars(request)["id"]

	var album models.AlbumPageResponse
	var releaseDate string

	query := `
		SELECT album.id, album.title, YEAR(album.release_date), album.cover_path,
		album.description, musician.id, musician.name, musician.avatar_path
		FROM album
		JOIN musician ON album.musician_id = musician.id
		WHERE album.id = ?
	`

	err := handler.DB.QueryRow(query, albumID).Scan(&album.ID, &album.Title, &releaseDate,
		&album.CoverURL, &album.Description, &album.ArtistID, &album.ArtistName,
		&album.ArtistAvatarURL,
	)
	baseURL := "http://37.46.130.29:8080"
	album.CoverURL = fmt.Sprintf("%s/media/image/%s", baseURL, filepath.Base(album.CoverURL))
	if err != nil {
		log.Println("GetAlbum - Error fetching album:", err)
		http.Error(response, "Album not found", http.StatusNotFound)
		return
	}

	// Получаем ID треков альбома
	rows, err := handler.DB.Query(`SELECT id FROM track WHERE album_id = ? ORDER BY id`, albumID)
	if err != nil {
		log.Println("GetAlbum - Error fetching tracks:", err)
		http.Error(response, "Failed to load tracks", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var trackID int
		if err := rows.Scan(&trackID); err == nil {
			album.Tracks = append(album.Tracks, trackID)
		}
	}

	response.Header().Set("Content-Type", "application/json")
	json.NewEncoder(response).Encode(album)
}

func (handler *AlbumHandler) GetAlbumTracks(response http.ResponseWriter, request *http.Request) {
	albumID := mux.Vars(request)["id"]

	query := `
		SELECT 
			t.id, t.title, t.musician_id, m.name, a.cover_path,
			t.file_path, t.duration, t.stream_count, t.visibility
		FROM track t
		JOIN musician m ON t.musician_id = m.id
		JOIN album a ON t.album_id = a.id
		WHERE t.album_id = ?
		ORDER BY t.id
	`

	rows, err := handler.DB.Query(query, albumID)
	if err != nil {
		log.Println("GetAlbumTracks - Error querying tracks:", err)
		http.Error(response, "Failed to fetch tracks", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var tracks []models.TrackResponse

	for rows.Next() {
		var track models.TrackResponse

		err := rows.Scan(&track.ID, &track.Title, &track.ArtistID, &track.ArtistName,
			&track.ImageURL, &track.AudioURL, &track.Duration, &track.Plays, &track.Visibility,
		)
		if err != nil {
			log.Println("GetAlbumTracks - Error scanning row:", err)
			continue
		}
		baseURL := "http://37.46.130.29:8080"
		track.AudioURL = fmt.Sprintf("%s/media/audio/%s", baseURL, track.ID)
		track.ImageURL = fmt.Sprintf("%s/media/image/%s", baseURL, filepath.Base(track.ImageURL))
		tracks = append(tracks, track)
	}

	if err = rows.Err(); err != nil {
		log.Println("GetAlbumTracks - Row error:", err)
		http.Error(response, "Failed to fetch tracks", http.StatusInternalServerError)
		return
	}

	response.Header().Set("Content-Type", "application/json")
	json.NewEncoder(response).Encode(tracks)
}
