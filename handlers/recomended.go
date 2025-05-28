package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/Edafi/MusicVibe/middleware"
	"github.com/Edafi/MusicVibe/models"
)

type RecommendHandler struct {
	DB *sql.DB
}

// GET /tracks/recommended
func (handler *RecommendHandler) GetRecommendedTracks(response http.ResponseWriter, request *http.Request) {
	userID := request.Context().Value(middleware.ContextUserIDKey).(string)

	query := `
	SELECT 
		t.id, t.title, t.musician_id, m.name, 
		a.cover_path, t.file_path, t.duration, t.stream_count
	FROM track t
	JOIN album a ON t.album_id = a.id
	JOIN musician m ON t.musician_id = m.id
	JOIN user_genre ug ON t.genre_id = ug.genre_id
	WHERE ug.user_id = ?
	`

	rows, err := handler.DB.Query(query, userID)
	if err != nil {
		http.Error(response, "Database query error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var tracks []models.RecommendedTrack

	for rows.Next() {
		var t models.RecommendedTrack
		err := rows.Scan(
			&t.ID, &t.Title, &t.ArtistID, &t.ArtistName,
			&t.ImageURL, &t.AudioURL, &t.Duration, &t.Plays,
		)
		if err != nil {
			http.Error(response, "Scan error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		tracks = append(tracks, t)
	}

	response.Header().Set("Content-Type", "application/json")
	json.NewEncoder(response).Encode(tracks)
}

func (handler *RecommendHandler) GetRecommendedAlbums(response http.ResponseWriter, request *http.Request) {
	userID := request.Context().Value(middleware.ContextUserIDKey).(string)

	query := `
	SELECT 
		a.id, a.musician_id, a.title, a.release_date,
		a.cover_path, a.genre_id, a.description,
		m.name AS musician_name
	FROM album a
	JOIN musician m ON a.musician_id = m.id
	JOIN user_genres ug ON a.genre_id = ug.genre_id
	WHERE ug.user_id = ?
	`

	rows, err := handler.DB.Query(query, userID)
	if err != nil {
		http.Error(response, "Database query error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type RecommendedAlbum struct {
		models.Album
		MusicianName string `json:"musician_name"`
	}

	var albums []RecommendedAlbum
	for rows.Next() {
		var a RecommendedAlbum
		err := rows.Scan(
			&a.ID, &a.MusicianID, &a.Title, &a.ReleaseDate,
			&a.CoverPath, &a.GenreID, &a.Description,
			&a.MusicianName,
		)
		if err != nil {
			http.Error(response, "Scan error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		albums = append(albums, a)
	}

	response.Header().Set("Content-Type", "application/json")
	json.NewEncoder(response).Encode(albums)
}
