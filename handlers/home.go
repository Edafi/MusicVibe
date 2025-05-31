package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/Edafi/MusicVibe/middleware"
	"github.com/Edafi/MusicVibe/models"
)

type HomeHandler struct {
	DB *sql.DB
}

type AlbumResponse struct {
	models.Album
	Artist string `json:"artist"`
}

func (handler *HomeHandler) GetRecommendedTracks(response http.ResponseWriter, request *http.Request) {
	val := request.Context().Value(middleware.ContextUserIDKey)
	userID, ok := val.(string)
	if !ok {
		log.Println("UserID not found in context")
		http.Error(response, "Unauthorized", http.StatusUnauthorized)
		return
	}

	query := `
		SELECT t.id, t.musician_id, t.title, t.duration, t.file_path, t.stream_count, t.visibility,
		m.name AS artist, a.cover_path
		FROM track t
		JOIN album a ON t.album_id = a.id
		JOIN musician m ON t.musician_id = m.id
		JOIN musician_genre mg ON m.id = mg.musician_id
		JOIN user_genre ug ON mg.genre_id = ug.genre_id
		WHERE ug.user_id = ? and t.visibility = 'public'
		GROUP BY t.id
		ORDER BY RAND()
		LIMIT 50;
	`

	rows, err := handler.DB.Query(query, userID)
	if err != nil {
		log.Println("GetRecommendedTracks:", err)
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
			log.Println("GetRecommendedTracks:", err)
			http.Error(response, err.Error(), http.StatusInternalServerError)
			return
		}
		tracks = append(tracks, tr)
	}

	response.Header().Set("Content-Type", "application/json")
	json.NewEncoder(response).Encode(tracks)
}

func (handler *HomeHandler) GetRecommendedAlbums(response http.ResponseWriter, request *http.Request) {
	val := request.Context().Value(middleware.ContextUserIDKey)
	userID, ok := val.(string)
	if !ok {
		log.Println("UserID not found in context")
		http.Error(response, "Unauthorized", http.StatusUnauthorized)
		return
	}

	query := `
		SELECT a.id, a.title, a.musician_id, m.name AS artist_name,
		a.cover_path, YEAR(a.release_date), a.description
		FROM album a
		JOIN musician m ON a.musician_id = m.id
		JOIN user_genre ug ON a.genre_id = ug.genre_id
		WHERE ug.user_id = ?
		ORDER BY RAND()
		LIMIT 50;
	`

	rows, err := handler.DB.Query(query, userID)
	if err != nil {
		log.Println("GetRecommendedAlbums:", err)
		http.Error(response, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	var albums []models.RecommendedAlbum = make([]models.RecommendedAlbum, 0)
	for rows.Next() {
		var al models.RecommendedAlbum
		if err := rows.Scan(
			&al.ID, &al.Title, &al.ArtistID, &al.ArtistName,
			&al.CoverUrl, &al.Year, &al.Description,
		); err != nil {
			log.Println("GetRecommendedAlbums:", err)
			http.Error(response, err.Error(), http.StatusInternalServerError)
			return
		}
		albums = append(albums, al)
	}

	response.Header().Set("Content-Type", "application/json")
	json.NewEncoder(response).Encode(albums)
}

func (handler *HomeHandler) GetTrackedTracks(response http.ResponseWriter, request *http.Request) {
	val := request.Context().Value(middleware.ContextUserIDKey)
	userID, ok := val.(string)
	if !ok {
		log.Println("UserID not found in context")
		http.Error(response, "Unauthorized", http.StatusUnauthorized)
		return
	}
	query := `
		SELECT t.id, t.title, t.musician_id, m.name AS artist_name, a.cover_path, t.visibility
		t.file_path, t.duration, t.stream_count
		FROM liked_tracks lt
		JOIN track t ON lt.track_id = t.id
		JOIN album a ON t.album_id = a.id
		JOIN musician m ON t.musician_id = m.id
		WHERE lt.user_id = ? and t.visibility = 'public'
	`

	rows, err := handler.DB.Query(query, userID)
	if err != nil {
		log.Println("GetTrackedTracks:", err)
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
			log.Println("GetTrackedTracks:", err)
			http.Error(response, err.Error(), http.StatusInternalServerError)
			return
		}
		tracks = append(tracks, tr)
	}

	response.Header().Set("Content-Type", "application/json")
	json.NewEncoder(response).Encode(tracks)
}

func (handler *HomeHandler) GetHomeRecommendedTracks(response http.ResponseWriter, request *http.Request) {
	val := request.Context().Value(middleware.ContextUserIDKey)
	userID, ok := val.(string)
	if !ok {
		log.Println("UserID not found in context")
		http.Error(response, "Unauthorized", http.StatusUnauthorized)
		return
	}
	query := `
		SELECT t.id, t.musician_id, t.title, t.duration, t.file_path, t.stream_count, t.visibility,
		m.name AS artist, a.cover_path
		FROM track t
		JOIN album a ON t.album_id = a.id
		JOIN musician m ON t.musician_id = m.id
		JOIN musician_genre mg ON m.id = mg.musician_id
		JOIN user_genre ug ON mg.genre_id = ug.genre_id
		WHERE ug.user_id = ? and t.visibility = 'public'
		GROUP BY t.id
		ORDER BY RAND()
		LIMIT 8;
	`
	rows, err := handler.DB.Query(query, userID)
	if err != nil {
		log.Println("GetHomeRecommendedTracks:", err)
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
			log.Println("GetHomeRecommendedTracks:", err)
			http.Error(response, err.Error(), http.StatusInternalServerError)
			return
		}
		tracks = append(tracks, tr)
	}

	response.Header().Set("Content-Type", "application/json")
	json.NewEncoder(response).Encode(tracks)
}

func (handler *HomeHandler) GetHomeRecommendedAlbums(response http.ResponseWriter, request *http.Request) {
	val := request.Context().Value(middleware.ContextUserIDKey)
	userID, ok := val.(string)
	if !ok {
		log.Println("UserID not found in context")
		http.Error(response, "Unauthorized", http.StatusUnauthorized)
		return
	}
	query := `
		SELECT a.id, a.title, a.musician_id, m.name AS artist_name,
		a.cover_path, YEAR(a.release_date), a.description
		FROM album a
		JOIN musician m ON a.musician_id = m.id
		JOIN user_genre ug ON a.genre_id = ug.genre_id
		WHERE ug.user_id = ?
		ORDER BY RAND()
		LIMIT 8;
	`

	rows, err := handler.DB.Query(query, userID)
	if err != nil {
		log.Println("GetHomeRecommendedAlbums:", err)
		http.Error(response, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var albums []models.RecommendedAlbum = make([]models.RecommendedAlbum, 0)
	for rows.Next() {
		var al models.RecommendedAlbum
		if err := rows.Scan(
			&al.ID, &al.Title, &al.ArtistID, &al.ArtistName,
			&al.CoverUrl, &al.Year, &al.Description,
		); err != nil {
			log.Println("GetHomeRecommendedAlbums:", err)
			http.Error(response, err.Error(), http.StatusInternalServerError)
			return
		}
		albums = append(albums, al)
	}

	response.Header().Set("Content-Type", "application/json")
	json.NewEncoder(response).Encode(albums)
}

func (handler *HomeHandler) GetHomeTrackedTracks(response http.ResponseWriter, request *http.Request) {
	val := request.Context().Value(middleware.ContextUserIDKey)
	userID, ok := val.(string)
	if !ok {
		log.Println("UserID not found in context")
		http.Error(response, "Unauthorized", http.StatusUnauthorized)
		return
	}

	query := `
		SELECT t.id, t.title, t.musician_id, m.name AS artist_name, a.cover_path,
		t.file_path, t.duration, t.stream_count, t.visibility
		FROM liked_tracks lt
		JOIN track t ON lt.track_id = t.id
		JOIN album a ON t.album_id = a.id
		JOIN musician m ON t.musician_id = m.id
		WHERE lt.user_id = ? and t.visibility = 'public'
		LIMIT 8
	`

	rows, err := handler.DB.Query(query, userID)
	if err != nil {
		log.Println("GetHomeTrackedTracks:", err)
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
			log.Println("GetHomeTrackedTracks:", err)
			http.Error(response, err.Error(), http.StatusInternalServerError)
			return
		}
		tracks = append(tracks, tr)
	}

	response.Header().Set("Content-Type", "application/json")
	json.NewEncoder(response).Encode(tracks)
}
