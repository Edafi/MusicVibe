package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/Edafi/MusicVibe/middleware"
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
		SELECT t.id
		FROM liked_tracks lt
		JOIN track t ON lt.track_id = t.id
		WHERE lt.user_id = ?
	`

	rows, err := handler.DB.Query(query, userID)
	if err != nil {
		log.Println("GetFavoriteTracks - DB Query error:", err)
		http.Error(response, "Failed to load favorite track IDs", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var trackIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			log.Println("GetFavoriteTracks - Row Scan error:", err)
			http.Error(response, "Failed to parse track ID", http.StatusInternalServerError)
			return
		}
		trackIDs = append(trackIDs, id)
	}

	response.Header().Set("Content-Type", "application/json")
	json.NewEncoder(response).Encode(trackIDs)
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

	trackID := mux.Vars(request)["id"]

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

	trackID := mux.Vars(request)["id"]

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
		SELECT a.id
		FROM liked_albums la
		JOIN album a ON la.album_id = a.id
		WHERE la.user_id = ?
	`

	rows, err := handler.DB.Query(query, userID)
	if err != nil {
		log.Println("GetFavoriteAlbums - DB Query error:", err)
		http.Error(response, "Failed to load favorite album IDs", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var albumIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			log.Println("GetFavoriteAlbums - Row Scan error:", err)
			http.Error(response, "Failed to parse album ID", http.StatusInternalServerError)
			return
		}
		albumIDs = append(albumIDs, id)
	}

	response.Header().Set("Content-Type", "application/json")
	json.NewEncoder(response).Encode(albumIDs)
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

	albumID := mux.Vars(request)["id"]

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

	albumID := mux.Vars(request)["id"]

	_, err := handler.DB.Exec(`DELETE FROM liked_albums WHERE user_id = ? AND album_id = ?`, userID, albumID)
	if err != nil {
		log.Println("DeleteFavoriteAlbum - Delete error:", err)
		http.Error(response, "Failed to remove album from favorites", http.StatusInternalServerError)
		return
	}

	response.WriteHeader(http.StatusNoContent)
}
