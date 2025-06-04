package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/Edafi/MusicVibe/middleware"
	"github.com/Edafi/MusicVibe/models"
	"github.com/gorilla/mux"
)

type FavoritesHandler struct {
	DB *sql.DB
}

// Получить избранные треки
func (handler *FavoritesHandler) GetFavoriteTracks(response http.ResponseWriter, request *http.Request) {
	val := request.Context().Value(middleware.ContextUserIDKey)
	userID, ok := val.(string)
	if !ok {
		log.Println("UserID not found in context")
		http.Error(response, "Unauthorized", http.StatusUnauthorized)
		return
	}

	query := `
		SELECT t.id, t.title, m.id, m.name, t.image_path, t.audio_path,
		t.duration, t.plays, t.visibility
		FROM liked_tracks lt
		JOIN track t ON lt.track_id = t.id
		JOIN musician m ON t.musician_id = m.id
		WHERE lt.user_id = ?
	`

	rows, err := handler.DB.Query(query, userID)
	if err != nil {
		log.Println("GetFavoriteTracks - DB Query error:", err)
		http.Error(response, "Failed to load favorite tracks", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var tracks []models.TrackResponse = make([]models.TrackResponse, 0)
	for rows.Next() {
		var track models.TrackResponse

		if err := rows.Scan(&track.ID, &track.Title, &track.ArtistID, &track.ArtistName,
			&track.ImageURL, &track.AudioURL, &track.Duration, &track.Plays, &track.Visibility); err != nil {
			log.Println("GetFavoriteTracks - Row Scan error:", err)
			http.Error(response, "Failed to parse track", http.StatusInternalServerError)
			return
		}
		tracks = append(tracks, track)
	}
	json.NewEncoder(response).Encode(tracks)
}

// Добавить трек в избранное
func (handler *FavoritesHandler) AddFavoriteTrack(response http.ResponseWriter, request *http.Request) {
	val := request.Context().Value(middleware.ContextUserIDKey)
	userID, ok := val.(string)
	if !ok {
		log.Println("UserID not found in context")
		http.Error(response, "Unauthorized", http.StatusUnauthorized)
		return
	}

	trackID := mux.Vars(request)["trackId"]

	_, err := handler.DB.Exec(`INSERT IGNORE INTO liked_tracks (user_id, track_id) VALUES (?, ?)`, userID, trackID)
	if err != nil {
		log.Println("AddFavoriteTrack - Insert error:", err)
		http.Error(response, "Failed to add to favorites", http.StatusInternalServerError)
		return
	}

	response.WriteHeader(http.StatusCreated)
}

// Удалить трек из избранного
func (handler *FavoritesHandler) DeleteFavoriteTrack(response http.ResponseWriter, request *http.Request) {
	val := request.Context().Value(middleware.ContextUserIDKey)
	userID, ok := val.(string)
	if !ok {
		log.Println("UserID not found in context")
		http.Error(response, "Unauthorized", http.StatusUnauthorized)
		return
	}

	trackID := mux.Vars(request)["trackId"]

	_, err := handler.DB.Exec(`DELETE FROM liked_tracks WHERE user_id = ? AND track_id = ?`, userID, trackID)
	if err != nil {
		log.Println("DeleteFavoriteTrack - Delete error:", err)
		http.Error(response, "Failed to remove from favorites", http.StatusInternalServerError)
		return
	}

	response.WriteHeader(http.StatusNoContent)
}

// Получить избранные альбомы
func (handler *FavoritesHandler) GetFavoriteAlbums(response http.ResponseWriter, request *http.Request) {
	val := request.Context().Value(middleware.ContextUserIDKey)
	userID, ok := val.(string)
	if !ok {
		log.Println("UserID not found in context")
		http.Error(response, "Unauthorized", http.StatusUnauthorized)
		return
	}

	query := `
		SELECT a.id, a.title, a.image_path, YEAR(a.release_date), m.id, m.name
		FROM liked_albums la
		JOIN album a ON la.album_id = a.id
		JOIN musician m ON a.musician_id = m.id
		WHERE la.user_id = ?
	`

	rows, err := handler.DB.Query(query, userID)
	if err != nil {
		log.Println("GetFavoriteAlbums - DB Query error:", err)
		http.Error(response, "Failed to load favorite albums", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var albums []models.AlbumPageResponse = make([]models.AlbumPageResponse, 0)
	for rows.Next() {
		var album models.AlbumPageResponse

		if err := rows.Scan(&album.ID, &album.Title, &album.CoverURL,
			&album.Year, &album.ArtistID, &album.ArtistName); err != nil {
			log.Println("GetFavoriteAlbums - Row Scan error:", err)
			http.Error(response, "Failed to parse album", http.StatusInternalServerError)
			return
		}
		albums = append(albums, album)
	}

	json.NewEncoder(response).Encode(albums)
}

// Добавить альбом в избранное
func (handler *FavoritesHandler) AddFavoriteAlbum(response http.ResponseWriter, request *http.Request) {
	val := request.Context().Value(middleware.ContextUserIDKey)
	userID, ok := val.(string)
	if !ok {
		log.Println("UserID not found in context")
		http.Error(response, "Unauthorized", http.StatusUnauthorized)
		return
	}

	albumID := mux.Vars(request)["albumId"]

	_, err := handler.DB.Exec(`INSERT IGNORE INTO liked_albums (user_id, album_id) VALUES (?, ?)`, userID, albumID)
	if err != nil {
		log.Println("AddFavoriteAlbum - Insert error:", err)
		http.Error(response, "Failed to add album to favorites", http.StatusInternalServerError)
		return
	}

	response.WriteHeader(http.StatusCreated)
}

// Удалить альбом из избранного
func (handler *FavoritesHandler) DeleteFavoriteAlbum(response http.ResponseWriter, request *http.Request) {
	val := request.Context().Value(middleware.ContextUserIDKey)
	userID, ok := val.(string)
	if !ok {
		log.Println("UserID not found in context")
		http.Error(response, "Unauthorized", http.StatusUnauthorized)
		return
	}

	albumID := mux.Vars(request)["albumId"]

	_, err := handler.DB.Exec(`DELETE FROM liked_albums WHERE user_id = ? AND album_id = ?`, userID, albumID)
	if err != nil {
		log.Println("DeleteFavoriteAlbum - Delete error:", err)
		http.Error(response, "Failed to remove album from favorites", http.StatusInternalServerError)
		return
	}

	response.WriteHeader(http.StatusNoContent)
}
