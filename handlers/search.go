package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/Edafi/MusicVibe/models"
)

type SearchHandler struct {
	DB *sql.DB
}

func (handler *SearchHandler) GetNewTracks(response http.ResponseWriter, request *http.Request) {
	query := `
	SELECT t.id, t.musician_id, t.album_id, t.title, t.duration, t.file_path,
	       t.genre_id, t.stream_count, t.visibility,
	       m.name AS artist, a.cover_path
	FROM track t
	JOIN musician m ON t.musician_id = m.id
	LEFT JOIN album a ON t.album_id = a.id
	ORDER BY t.id DESC
	WHERE t.visibility = 'public'
	LIMIT 8;
	`

	rows, err := handler.DB.Query(query)
	if err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var tracks []models.Track
	for rows.Next() {
		var tr models.Track
		if err := rows.Scan(
			&tr.ID, &tr.ArtistID, &tr.Title, &tr.Duration, &tr.AudioURL, &tr.Plays,
			&tr.ArtistName, &tr.ImageURL, &tr.Visibility,
		); err != nil {
			http.Error(response, err.Error(), http.StatusInternalServerError)
			return
		}
		tracks = append(tracks, tr)
	}
	json.NewEncoder(response).Encode(tracks)
}

func (handler *SearchHandler) GetChartTracks(response http.ResponseWriter, request *http.Request) {
	query := `
	SELECT t.id, t.musician_id, t.album_id, t.title, t.duration, t.file_path,
	       t.genre_id, t.stream_count, t.visibility,
	       m.name AS artist, a.cover_path
	FROM track t
	JOIN musician m ON t.musician_id = m.id
	LEFT JOIN album a ON t.album_id = a.id
	ORDER BY t.stream_count DESC
	WHERE t.visibility = 'public'
	LIMIT 8;
	`

	rows, err := handler.DB.Query(query)
	if err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var tracks []models.Track
	for rows.Next() {
		var tr models.Track
		if err := rows.Scan(
			&tr.ID, &tr.ArtistID, &tr.Title, &tr.Duration, &tr.AudioURL, &tr.Plays,
			&tr.ArtistName, &tr.ImageURL, &tr.Visibility,
		); err != nil {
			http.Error(response, err.Error(), http.StatusInternalServerError)
			return
		}
		tracks = append(tracks, tr)
	}
	json.NewEncoder(response).Encode(tracks)
}

func (handler *SearchHandler) SearchTracks(response http.ResponseWriter, request *http.Request) {
	q := request.URL.Query().Get("query")
	if q == "" {
		http.Error(response, "Missing query parameter", http.StatusBadRequest)
		return
	}

	query := `
	SELECT t.id, t.musician_id, t.album_id, t.title, t.duration, t.file_path,
	       t.genre_id, t.stream_count, t.visibility,
	       m.name AS artist, a.cover_path
	FROM track t
	JOIN musician m ON t.musician_id = m.id
	LEFT JOIN album a ON t.album_id = a.id
	WHERE t.title LIKE ? OR m.name LIKE ? AND t.visibility = 'public'
	LIMIT 20;
	`

	likePattern := "%" + q + "%"
	rows, err := handler.DB.Query(query, likePattern, likePattern)
	if err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var tracks []models.Track
	for rows.Next() {
		var tr models.Track
		if err := rows.Scan(
			&tr.ID, &tr.ArtistID, &tr.Title, &tr.Duration, &tr.AudioURL, &tr.Plays,
			&tr.ArtistName, &tr.ImageURL, &tr.Visibility,
		); err != nil {
			http.Error(response, err.Error(), http.StatusInternalServerError)
			return
		}
		tracks = append(tracks, tr)
	}
	json.NewEncoder(response).Encode(tracks)
}
