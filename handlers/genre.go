package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/Edafi/MusicVibe/middleware"
	"github.com/Edafi/MusicVibe/models"
)

type GenreHandler struct {
	DB *sql.DB
}

// -------------------------------GET---------------------------------------------------//
func (handler *GenreHandler) GetGenres(response http.ResponseWriter, request *http.Request) {
	rows, err := handler.DB.Query("SELECT id, name FROM genre")
	if err != nil {
		log.Fatal(err)
		http.Error(response, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var genres []models.Genre
	for rows.Next() {
		var genre models.Genre
		if err := rows.Scan(&genre.ID, &genre.Name); err != nil {
			http.Error(response, err.Error(), http.StatusInternalServerError)
			return
		}
		genres = append(genres, genre)
	}

	response.Header().Set("Content-Type", "application/json")
	json.NewEncoder(response).Encode(genres)
}

// -------------------------------POST---------------------------------------------------//
func (handler *GenreHandler) PostUserGenres(response http.ResponseWriter, request *http.Request) {
	userID, ok := request.Context().Value(middleware.ContextUserIDKey).(string)
	if !ok || userID == "" {
		log.Fatal(ok)
		http.Error(response, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		GenreIDs []int `json:"genre_ids"`
	}
	if err := json.NewDecoder(request.Body).Decode(&req); err != nil {
		http.Error(response, "Invalid JSON", http.StatusBadRequest)
		return
	}

	transaction, err := handler.DB.Begin()
	if err != nil {
		log.Fatal(err)
		http.Error(response, "Failed to begin transaction", http.StatusInternalServerError)
		return
	}

	_, err = transaction.Exec("DELETE FROM user_genre WHERE user_id = ?", userID)
	if err != nil {
		log.Fatal(err)
		transaction.Rollback()
		http.Error(response, "Failed to clear existing genres", http.StatusInternalServerError)
		return
	}

	preparedSQL, err := transaction.Prepare("INSERT INTO user_genre (user_id, genre_id) VALUES (?, ?)")
	if err != nil {
		log.Fatal(err)
		transaction.Rollback()
		http.Error(response, "Failed to prepare insert", http.StatusInternalServerError)
		return
	}
	defer preparedSQL.Close()

	for _, genreID := range req.GenreIDs {
		_, err := preparedSQL.Exec(userID, genreID)
		if err != nil {
			log.Fatal(err)
			transaction.Rollback()
			http.Error(response, "Failed to insert genre", http.StatusInternalServerError)
			return
		}
	}

	if err := transaction.Commit(); err != nil {
		log.Fatal(err)
		http.Error(response, "Failed to commit", http.StatusInternalServerError)
		return
	}

	response.WriteHeader(http.StatusOK)
}
