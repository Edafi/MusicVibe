package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/Edafi/MusicVibe/models"
)

type MusicianHandler struct {
	DB *sql.DB
}

// GET /musicians?user_id=123
func (handler *MusicianHandler) GetMusicians(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "Missing user_id", http.StatusBadRequest)
		return
	}

	query := `
	SELECT DISTINCT m.id, m.user_id, m.name, m.avatar_path
	FROM musician m
	JOIN musician_genre mg ON m.id = mg.musician_id
	JOIN user_genres ug ON mg.genre_id = ug.genre_id
	WHERE ug.user_id = ?
	`

	rows, err := handler.DB.Query(query, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var musicians []models.Musician
	for rows.Next() {
		var m models.Musician
		if err := rows.Scan(&m.ID, &m.UserID, &m.Name, &m.AvatarPath); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		musicians = append(musicians, m)
	}

	json.NewEncoder(w).Encode(musicians)
}

// POST /user/following
func (handler *MusicianHandler) PostUserFollowing(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		UserID      string   `json:"user_id"`
		MusicianIDs []string `json:"musician_ids"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tx, err := handler.DB.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("INSERT INTO user_following (user_id, musician_id) VALUES (?, ?)")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	for _, musicianID := range payload.MusicianIDs {
		if _, err := stmt.Exec(payload.UserID, musicianID); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}
