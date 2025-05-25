package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/Edafi/MusicVibe/middleware"
	"github.com/Edafi/MusicVibe/models"
)

type HomeHandler struct {
	DB *sql.DB
}

type TrackResponse struct {
	models.Track
	Artist    string `json:"artist"`
	CoverPath string `json:"cover_path"`
}

type AlbumResponse struct {
	models.Album
	Artist string `json:"artist"`
}

func (handler *HomeHandler) GetRecommendedTracks(response http.ResponseWriter, request *http.Request) {
	query := `
	SELECT t.id, t.musician_id, t.album_id, t.title, t.duration, t.file_path, 
	       t.genre_id, t.stream_count, t.visibility,
	       m.name AS artist, m.avatar_path AS cover_path
	FROM track t
	JOIN musician m ON t.musician_id = m.id
	LIMIT 8;
	`

	rows, err := handler.DB.Query(query)
	if err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var tracks []TrackResponse
	for rows.Next() {
		var tr TrackResponse
		if err := rows.Scan(
			&tr.ID, &tr.MusicianID, &tr.AlbumID, &tr.Title, &tr.Duration,
			&tr.FilePath, &tr.GenreID, &tr.StreamCount, &tr.Visibility,
			&tr.Artist, &tr.CoverPath,
		); err != nil {
			http.Error(response, err.Error(), http.StatusInternalServerError)
			return
		}
		tracks = append(tracks, tr)
	}
	json.NewEncoder(response).Encode(tracks)
}

func (handler *HomeHandler) GetRecommendedAlbums(response http.ResponseWriter, request *http.Request) {
	query := `
	SELECT a.id, a.musician_id, a.title, a.release_date, a.cover_path, a.genre_id, a.description,
	       m.name AS artist
	FROM album a
	JOIN musician m ON a.musician_id = m.id
	LIMIT 8;
	`

	rows, err := handler.DB.Query(query)
	if err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var albums []AlbumResponse
	for rows.Next() {
		var ar AlbumResponse
		if err := rows.Scan(
			&ar.ID, &ar.MusicianID, &ar.Title, &ar.ReleaseDate,
			&ar.CoverPath, &ar.GenreID, &ar.Description,
			&ar.Artist,
		); err != nil {
			http.Error(response, err.Error(), http.StatusInternalServerError)
			return
		}
		albums = append(albums, ar)
	}
	json.NewEncoder(response).Encode(albums)
}

func (handler *HomeHandler) GetTrackedTracks(response http.ResponseWriter, request *http.Request) {
	userID := request.Context().Value(middleware.ContextUserIDKey).(string)

	query := `
	SELECT t.id, t.musician_id, t.album_id, t.title, t.duration, t.file_path, 
	       t.genre_id, t.stream_count, t.visibility,
	       m.name AS artist, m.avatar_path AS cover_path
	FROM track t
	JOIN musician m ON t.musician_id = m.id
	JOIN user_following uf ON m.id = uf.musician_id
	WHERE uf.user_id = ?
	LIMIT 8;
	`

	rows, err := handler.DB.Query(query, userID)
	if err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var tracks []TrackResponse
	for rows.Next() {
		var tr TrackResponse
		if err := rows.Scan(
			&tr.ID, &tr.MusicianID, &tr.AlbumID, &tr.Title, &tr.Duration,
			&tr.FilePath, &tr.GenreID, &tr.StreamCount, &tr.Visibility,
			&tr.Artist, &tr.CoverPath,
		); err != nil {
			http.Error(response, err.Error(), http.StatusInternalServerError)
			return
		}
		tracks = append(tracks, tr)
	}
	json.NewEncoder(response).Encode(tracks)
}
