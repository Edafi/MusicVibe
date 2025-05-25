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
func (h *RecommendHandler) GetRecommendedTracks(response http.ResponseWriter, request *http.Request) {
	userID := request.Context().Value(middleware.ContextUserIDKey).(string)

	query := `
	SELECT 
		t.id, t.musician_id, t.album_id, t.title, t.duration,
		t.file_path, t.genre_id, t.stream_count, t.visibility,
		a.cover_path,
		m.name AS musician_name
	FROM track t
	JOIN album a ON t.album_id = a.id
	JOIN musician m ON t.musician_id = m.id
	JOIN user_genres ug ON t.genre_id = ug.genre_id
	WHERE ug.user_id = ?
	`

	rows, err := h.DB.Query(query, userID)
	if err != nil {
		http.Error(response, "Database query error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type RecommendedTrack struct {
		models.Track
		CoverPath    string `json:"cover_path"`
		MusicianName string `json:"musician_name"`
	}

	var tracks []RecommendedTrack
	for rows.Next() {
		var t RecommendedTrack
		err := rows.Scan(
			&t.ID, &t.MusicianID, &t.AlbumID, &t.Title, &t.Duration,
			&t.FilePath, &t.GenreID, &t.StreamCount, &t.Visibility,
			&t.CoverPath, &t.MusicianName,
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

func (h *RecommendHandler) GetRecommendedAlbums(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.ContextUserIDKey).(string)

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

	rows, err := h.DB.Query(query, userID)
	if err != nil {
		http.Error(w, "Database query error: "+err.Error(), http.StatusInternalServerError)
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
			http.Error(w, "Scan error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		albums = append(albums, a)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(albums)
}
