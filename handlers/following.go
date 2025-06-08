package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/Edafi/MusicVibe/middleware"
	"github.com/gorilla/mux"
)

type FollowingHandler struct {
	DB *sql.DB
}

// Получить подписанных музыкантов
func (handler *FollowingHandler) GetFollowingMusicians(response http.ResponseWriter, request *http.Request) {
	val := request.Context().Value(middleware.ContextUserIDKey)
	userID, ok := val.(string)
	if !ok {
		log.Println("UserID not found in context")
		http.Error(response, "Unauthorized", http.StatusUnauthorized)
		return
	}

	rows, err := handler.DB.Query(`
		SELECT m.id
		FROM user_following uf
		JOIN musician m ON uf.musician_id = m.id
		WHERE uf.user_id = ?`, userID)
	if err != nil {
		log.Println("GetFollowingMusicians - DB Query error:", err)
		http.Error(response, "Failed to load followed musicians", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var musicianIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			log.Println("GetFollowingMusicians - Row Scan error:", err)
			http.Error(response, "Failed to parse musician ID", http.StatusInternalServerError)
			return
		}
		musicianIDs = append(musicianIDs, id)
	}

	response.Header().Set("Content-Type", "application/json")
	json.NewEncoder(response).Encode(musicianIDs)
}

// Подписаться
func (handler *FollowingHandler) FollowMusician(response http.ResponseWriter, request *http.Request) {

	log.Println("FollowMusician handler called")
	log.Println("Path:", request.URL.Path)
	log.Println("Vars:", mux.Vars(request))

	val := request.Context().Value(middleware.ContextUserIDKey)
	userID, ok := val.(string)
	if !ok {
		log.Println("UserID not found in context")
		http.Error(response, "Unauthorized", http.StatusUnauthorized)
		return
	}

	log.Println("User ID:", userID)

	musicianID := mux.Vars(request)["id"]

	log.Println("Musician ID:", musicianID)

	_, err := handler.DB.Exec("INSERT IGNORE INTO user_following (user_id, musician_id) VALUES (?, ?)", userID, musicianID)
	if err != nil {
		http.Error(response, "Insert error", http.StatusInternalServerError)
		return
	}

	response.WriteHeader(http.StatusCreated)
}

// Отписаться
func (handler *FollowingHandler) UnfollowMusician(response http.ResponseWriter, request *http.Request) {

	log.Println("Path:", request.URL.Path)
	log.Println("Vars:", mux.Vars(request))

	val := request.Context().Value(middleware.ContextUserIDKey)
	userID, ok := val.(string)
	if !ok {
		log.Println("UserID not found in context")
		http.Error(response, "Unauthorized", http.StatusUnauthorized)
		return
	}

	log.Println("User ID:", userID)

	musicianID := mux.Vars(request)["id"]

	log.Println("Musician ID:", musicianID)

	_, err := handler.DB.Exec("DELETE FROM user_following WHERE user_id = ? AND musician_id = ?", userID, musicianID)
	if err != nil {
		http.Error(response, "Delete error", http.StatusInternalServerError)
		return
	}

	response.WriteHeader(http.StatusNoContent)
}
