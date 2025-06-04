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

	log.Println("User ID:", userID)

	rows, err := handler.DB.Query(`
		SELECT m.id, u.email, m.name, u.avatar_path, u.background_path, 
		u.description, u.has_complete_setup
		FROM user_following uf
		JOIN musician m ON uf.musician_id = m.id
		JOIN user u ON u.id = m.user_id
		WHERE uf.user_id = ?`, userID)
	if err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var musicians []models.MusicianPage = make([]models.MusicianPage, 0)
	for rows.Next() {
		var m models.MusicianPage
		err := rows.Scan(&m.ID, &m.Email, &m.Name, &m.AvatarPath, &m.BackgroundPath, &m.Description, &m.HasCompleteSetup)
		if err != nil {
			http.Error(response, err.Error(), http.StatusInternalServerError)
			return
		}

		// Жанры
		genreRows, err := handler.DB.Query(`
			SELECT g.name 
			FROM musician_genre mg 
			JOIN genre g ON g.id = mg.genre_id 
			WHERE mg.musician_id = ?`, m.ID)
		if err == nil {
			for genreRows.Next() {
				var genre string
				genreRows.Scan(&genre)
				m.Genres = append(m.Genres, genre)
			}
			genreRows.Close()
		}

		// Социальные ссылки
		socialRows, err := handler.DB.Query(`
			SELECT name, url 
			FROM musician_social_link 
			WHERE musician_id = ?`, m.ID)
		if err == nil {
			for socialRows.Next() {
				var link models.SocialLink
				socialRows.Scan(&link.Name, &link.URL)
				m.SocialLinks = append(m.SocialLinks, link)
			}
			socialRows.Close()
		}

		// Альбомы (предпросмотр)
		albumRows, err := handler.DB.Query(`
			SELECT id, title, cover_path 
			FROM album 
			WHERE musician_id = ?`, m.ID)
		if err == nil {
			for albumRows.Next() {
				var album models.AlbumPreview
				albumRows.Scan(&album.ID, &album.Title, &album.CoverUrl)
				m.Albums = append(m.Albums, album)
			}
			albumRows.Close()
		}

		musicians = append(musicians, m)
	}

	response.Header().Set("Content-Type", "application/json")
	json.NewEncoder(response).Encode(musicians)
}

// Подписаться
func (handler *FollowingHandler) FollowMusician(response http.ResponseWriter, request *http.Request) {
	val := request.Context().Value(middleware.ContextUserIDKey)
	userID, ok := val.(string)
	if !ok {
		log.Println("UserID not found in context")
		http.Error(response, "Unauthorized", http.StatusUnauthorized)
		return
	}

	log.Println("User ID:", userID)

	musicianID := mux.Vars(request)["id"]

	_, err := handler.DB.Exec("INSERT IGNORE INTO user_following (user_id, musician_id) VALUES (?, ?)", userID, musicianID)
	if err != nil {
		http.Error(response, "Insert error", http.StatusInternalServerError)
		return
	}

	response.WriteHeader(http.StatusCreated)
}

// Отписаться
func (handler *FollowingHandler) UnfollowMusician(response http.ResponseWriter, request *http.Request) {
	val := request.Context().Value(middleware.ContextUserIDKey)
	userID, ok := val.(string)
	if !ok {
		log.Println("UserID not found in context")
		http.Error(response, "Unauthorized", http.StatusUnauthorized)
		return
	}

	log.Println("User ID:", userID)

	musicianID := mux.Vars(request)["id"]

	_, err := handler.DB.Exec("DELETE FROM user_following WHERE user_id = ? AND musician_id = ?", userID, musicianID)
	if err != nil {
		http.Error(response, "Delete error", http.StatusInternalServerError)
		return
	}

	response.WriteHeader(http.StatusNoContent)
}
