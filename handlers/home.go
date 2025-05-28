package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/Edafi/MusicVibe/middleware"
	"github.com/Edafi/MusicVibe/models"
	"github.com/golang-jwt/jwt/v5"
)

type HomeHandler struct {
	DB *sql.DB
}

type AlbumResponse struct {
	models.Album
	Artist string `json:"artist"`
}

func (handler *HomeHandler) GetRecommendedTracks(response http.ResponseWriter, request *http.Request) {
	claims, ok := request.Context().Value("claims").(*Claims)
	if !ok {
		http.Error(response, "Unauthorized", http.StatusUnauthorized)
		return
	}

	query := `
		SELECT t.id, t.musician_id, t.title, t.duration, t.file_path, t.stream_count,
		m.name AS artist, a.cover_path
		FROM track t
		JOIN album a ON t.album_id = a.id
		JOIN musician m ON t.musician_id = m.id
		JOIN musician_genre mg ON m.id = mg.musician_id
		JOIN user_genre ug ON mg.genre_id = ug.genre_id
		WHERE ug.user_id = ?
		GROUP BY t.id
		ORDER BY RAND()
		LIMIT 50;
	`

	rows, err := handler.DB.Query(query, claims.UserID)
	if err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var tracks []models.RecommendedTrack
	for rows.Next() {
		var tr models.RecommendedTrack
		if err := rows.Scan(
			&tr.ID, &tr.ArtistID, &tr.Title, &tr.Duration, &tr.AudioURL, &tr.Plays,
			&tr.ArtistName, &tr.ImageURL,
		); err != nil {
			http.Error(response, err.Error(), http.StatusInternalServerError)
			return
		}
		tracks = append(tracks, tr)
	}

	response.Header().Set("Content-Type", "application/json")
	json.NewEncoder(response).Encode(tracks)
}

func (handler *HomeHandler) GetRecommendedAlbums(response http.ResponseWriter, request *http.Request) {
	claims, ok := request.Context().Value("claims").(*Claims)
	if !ok {
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

	rows, err := handler.DB.Query(query, claims.UserID)
	if err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var albums []models.RecommendedAlbum
	for rows.Next() {
		var al models.RecommendedAlbum
		if err := rows.Scan(
			&al.ID, &al.Title, &al.ArtistID, &al.ArtistName,
			&al.CoverUrl, &al.Year, &al.Description,
		); err != nil {
			http.Error(response, err.Error(), http.StatusInternalServerError)
			return
		}
		albums = append(albums, al)
	}

	response.Header().Set("Content-Type", "application/json")
	json.NewEncoder(response).Encode(albums)
}

func (handler *HomeHandler) GetTrackedTracks(response http.ResponseWriter, request *http.Request) {
	userID := request.Context().Value(middleware.ContextUserIDKey).(string)

	query := `
		SELECT t.id, t.title, t.musician_id, m.name AS artist_name, a.cover_path,
		t.file_path, t.duration, t.stream_count
		FROM liked_tracks lt
		JOIN track t ON lt.track_id = t.id
		JOIN album a ON t.album_id = a.id
		JOIN musician m ON t.musician_id = m.id
		WHERE lt.user_id = ?
	`

	rows, err := handler.DB.Query(query, userID)
	if err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var tracks []models.RecommendedTrack
	for rows.Next() {
		var tr models.RecommendedTrack
		if err := rows.Scan(&tr.ID, &tr.Title, &tr.ArtistID, &tr.ArtistName,
			&tr.ImageURL, &tr.AudioURL, &tr.Duration, &tr.Plays,
		); err != nil {
			http.Error(response, err.Error(), http.StatusInternalServerError)
			return
		}
		tracks = append(tracks, tr)
	}

	response.Header().Set("Content-Type", "application/json")
	json.NewEncoder(response).Encode(tracks)
}

func (handler *HomeHandler) GetHomeRecommendedTracks(response http.ResponseWriter, request *http.Request) {
	authHeader := request.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(response, "Missing token", http.StatusUnauthorized)
		return
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil || !token.Valid {
		http.Error(response, "Invalid token", http.StatusUnauthorized)
		return
	}

	query := `
		SELECT 
		t.id, t.title, t.musician_id, m.name AS artist_name, a.cover_path, t.stream_count
		FROM track t
		JOIN album a ON t.album_id = a.id
		JOIN musician m ON t.musician_id = m.id
		JOIN user_genre ug ON ug.genre_id = t.genre_id
		WHERE ug.user_id = ?
		ORDER BY RAND()
		LIMIT 8;
	`
	rows, err := handler.DB.Query(query, claims.UserID)
	if err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var tracks []models.HomeRecomendedTrack
	for rows.Next() {
		var tr models.HomeRecomendedTrack
		if err := rows.Scan(&tr.ID, &tr.Title, &tr.MusicianID, &tr.MusicianName, &tr.ImageURL, &tr.Plays); err != nil {
			http.Error(response, err.Error(), http.StatusInternalServerError)
			return
		}
		tracks = append(tracks, tr)
	}

	response.Header().Set("Content-Type", "application/json")
	json.NewEncoder(response).Encode(tracks)
}

func (handler *HomeHandler) GetHomeRecommendedAlbums(response http.ResponseWriter, request *http.Request) {
	claims, ok := request.Context().Value("claims").(*Claims)
	if !ok {
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

	rows, err := handler.DB.Query(query, claims.UserID)
	if err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var albums []models.RecommendedAlbum
	for rows.Next() {
		var al models.RecommendedAlbum
		if err := rows.Scan(
			&al.ID, &al.Title, &al.ArtistID, &al.ArtistName,
			&al.CoverUrl, &al.Year, &al.Description,
		); err != nil {
			http.Error(response, err.Error(), http.StatusInternalServerError)
			return
		}
		albums = append(albums, al)
	}

	response.Header().Set("Content-Type", "application/json")
	json.NewEncoder(response).Encode(albums)
}

func (handler *HomeHandler) GetHomeTrackedTracks(response http.ResponseWriter, request *http.Request) {
	userID := request.Context().Value(middleware.ContextUserIDKey).(string)

	query := `
		SELECT t.id, t.title, t.musician_id, m.name AS artist_name, a.cover_path,
		t.file_path, t.duration, t.stream_count
		FROM liked_tracks lt
		JOIN track t ON lt.track_id = t.id
		JOIN album a ON t.album_id = a.id
		JOIN musician m ON t.musician_id = m.id
		WHERE lt.user_id = ?
		LIMIT 8
	`

	rows, err := handler.DB.Query(query, userID)
	if err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var tracks []models.RecommendedTrack
	for rows.Next() {
		var tr models.RecommendedTrack
		if err := rows.Scan(&tr.ID, &tr.Title, &tr.ArtistID, &tr.ArtistName,
			&tr.ImageURL, &tr.AudioURL, &tr.Duration, &tr.Plays,
		); err != nil {
			http.Error(response, err.Error(), http.StatusInternalServerError)
			return
		}
		tracks = append(tracks, tr)
	}

	response.Header().Set("Content-Type", "application/json")
	json.NewEncoder(response).Encode(tracks)
}
