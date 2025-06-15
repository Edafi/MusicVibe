package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/Edafi/MusicVibe/models"
)

type SearchHandler struct {
	DB *sql.DB
}

func (handler *SearchHandler) GetNewTracks(response http.ResponseWriter, request *http.Request) {
	query := `
	SELECT 
	t.id, t.musician_id, t.title, t.duration, t.file_path,
	t.stream_count, m.name AS artist, a.cover_path, t.visibility
	FROM track t
	JOIN musician m ON t.musician_id = m.id
	LEFT JOIN album a ON t.album_id = a.id
	WHERE t.visibility = 'public'
	ORDER BY a.release_date DESC
	LIMIT 8;
	`

	rows, err := handler.DB.Query(query)
	if err != nil {
		log.Println("GetNewTracks: ", err)
		http.Error(response, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var tracks []models.TrackResponse = make([]models.TrackResponse, 0)
	for rows.Next() {
		var tr models.TrackResponse
		if err := rows.Scan(
			&tr.ID, &tr.ArtistID, &tr.Title, &tr.Duration, &tr.AudioURL, &tr.Plays,
			&tr.ArtistName, &tr.ImageURL, &tr.Visibility,
		); err != nil {
			log.Println("GetNewTracks: ", err)
			http.Error(response, err.Error(), http.StatusInternalServerError)
			return
		}
		baseURL := "http://37.46.130.29:8080"
		tr.AudioURL = fmt.Sprintf("%s/media/audio/%s", baseURL, tr.ID)
		tr.ImageURL = fmt.Sprintf("%s/media/image/%s", baseURL, filepath.Base(tr.ImageURL))
		tracks = append(tracks, tr)
	}
	json.NewEncoder(response).Encode(tracks)
}

func (handler *SearchHandler) GetChartTracks(response http.ResponseWriter, request *http.Request) {
	query := `
	SELECT 
	t.id, t.musician_id, t.title, t.duration, t.file_path,
	t.stream_count, m.name AS artist, a.cover_path, t.visibility
	FROM track t
	JOIN musician m ON t.musician_id = m.id
	LEFT JOIN album a ON t.album_id = a.id
	WHERE t.visibility = 'public'
	ORDER BY t.stream_count DESC
	LIMIT 8;
	`

	rows, err := handler.DB.Query(query)
	if err != nil {
		log.Println("GetChartTracks: ", err)
		http.Error(response, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var tracks []models.TrackResponse = make([]models.TrackResponse, 0)
	for rows.Next() {
		var tr models.TrackResponse
		if err := rows.Scan(
			&tr.ID, &tr.ArtistID, &tr.Title, &tr.Duration, &tr.AudioURL, &tr.Plays,
			&tr.ArtistName, &tr.ImageURL, &tr.Visibility,
		); err != nil {
			log.Println("GetChartTracks: ", err)
			http.Error(response, err.Error(), http.StatusInternalServerError)
			return
		}
		baseURL := "http://37.46.130.29:8080"
		tr.AudioURL = fmt.Sprintf("%s/media/audio/%s", baseURL, tr.ID)
		tr.ImageURL = fmt.Sprintf("%s/media/image/%s", baseURL, filepath.Base(tr.ImageURL))
		tracks = append(tracks, tr)
	}
	json.NewEncoder(response).Encode(tracks)
}

func (handler *SearchHandler) SearchTracks(response http.ResponseWriter, request *http.Request) {
	q := request.URL.Query().Get("q")
	if q == "" {
		log.Println("SearchTracks: q is not valid")
		http.Error(response, "Missing query parameter", http.StatusBadRequest)
		return
	}
	query := `
	SELECT 
    t.id, t.title, t.musician_id, m.name, a.cover_path, 
    t.file_path, t.duration, t.stream_count, t.visibility
	FROM track t
	JOIN musician m ON t.musician_id = m.id
	LEFT JOIN album a ON t.album_id = a.id
	WHERE 
    (t.title_lower LIKE ? OR m.name_lower LIKE ?)
    AND t.visibility = 'public'
	LIMIT 50;
	`

	likePattern := "%" + q + "%"
	rows, err := handler.DB.Query(query, strings.ToLower(likePattern), strings.ToLower(likePattern))
	if err != nil {
		log.Println("SearchTracks: ", err)
		http.Error(response, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var tracks []models.TrackResponse = make([]models.TrackResponse, 0)
	for rows.Next() {
		var tr models.TrackResponse
		if err := rows.Scan(
			&tr.ID, &tr.Title, &tr.ArtistID, &tr.ArtistName, &tr.ImageURL,
			&tr.AudioURL, &tr.Duration, &tr.Plays, &tr.Visibility,
		); err != nil {
			log.Println("SearchTracks: ", err)
			http.Error(response, err.Error(), http.StatusInternalServerError)
			return
		}
		baseURL := "http://37.46.130.29:8080"
		tr.AudioURL = fmt.Sprintf("%s/media/audio/%s", baseURL, tr.ID)
		tr.ImageURL = fmt.Sprintf("%s/media/image/%s", baseURL, filepath.Base(tr.ImageURL))
		tracks = append(tracks, tr)
	}
	json.NewEncoder(response).Encode(tracks)
}
