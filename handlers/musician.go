package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/Edafi/MusicVibe/middleware"
	"github.com/Edafi/MusicVibe/models"
)

type MusicianHandler struct {
	DB *sql.DB
}

// --------------------- GET /musicians --------------------- //
func (handler *MusicianHandler) GetMusicians(response http.ResponseWriter, request *http.Request) {
	userID, ok := request.Context().Value(middleware.ContextUserIDKey).(string)
	if !ok || userID == "" {
		http.Error(response, "Unauthorized", http.StatusUnauthorized)
		return
	}

	query := `
	SELECT DISTINCT m.id, m.user_id, m.name, m.avatar_path
	FROM musician m
	JOIN musician_genre mg ON m.id = mg.musician_id
	JOIN user_genre ug ON mg.genre_id = ug.genre_id
	WHERE ug.user_id = ?
	`

	rows, err := handler.DB.Query(query, userID)
	if err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var musicians []models.Musician
	for rows.Next() {
		var m models.Musician
		if err := rows.Scan(&m.ID, &m.UserID, &m.Name, &m.AvatarPath); err != nil {
			http.Error(response, err.Error(), http.StatusInternalServerError)
			return
		}
		musicians = append(musicians, m)
	}

	response.Header().Set("Content-Type", "application/json")
	json.NewEncoder(response).Encode(musicians)
}

// --------------------- POST /user/following --------------------- //
func (handler *MusicianHandler) PostUserFollowing(response http.ResponseWriter, request *http.Request) {
	userID, ok := request.Context().Value(middleware.ContextUserIDKey).(string)
	if !ok || userID == "" {
		http.Error(response, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var payload struct {
		MusicianIDs []string `json:"musician_ids"`
	}

	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		http.Error(response, "Invalid JSON", http.StatusBadRequest)
		return
	}

	tx, err := handler.DB.Begin()
	if err != nil {
		http.Error(response, "Failed to begin transaction", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("INSERT INTO user_following (user_id, musician_id) VALUES (?, ?)")
	if err != nil {
		http.Error(response, "Failed to prepare insert", http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	for _, musicianID := range payload.MusicianIDs {
		if _, err := stmt.Exec(userID, musicianID); err != nil {
			http.Error(response, "Failed to insert follow", http.StatusInternalServerError)
			return
		}
	}

	if err := tx.Commit(); err != nil {
		http.Error(response, "Failed to commit transaction", http.StatusInternalServerError)
		return
	}

	response.WriteHeader(http.StatusCreated)
}

// GET /musician/{id}
func (h *MusicianHandler) GetMusician(w http.ResponseWriter, r *http.Request) {
	_ = r.Context().Value(middleware.ContextUserIDKey).(string)

	musicianID := mux.Vars(r)["id"]

	var musician models.Musician
	err := h.DB.QueryRow("SELECT id, user_id, name, avatar_path FROM musician WHERE id = ?", musicianID).
		Scan(&musician.ID, &musician.UserID, &musician.Name, &musician.AvatarPath)
	if err != nil {
		http.Error(w, "Musician not found", http.StatusNotFound)
		return
	}

	// Получаем альбомы музыканта
	albumsQuery := `
	SELECT id, musician_id, title, release_date, cover_path, genre_id, description
	FROM album
	WHERE musician_id = ?
	`
	albumRows, err := h.DB.Query(albumsQuery, musicianID)
	if err != nil {
		http.Error(w, "Error getting albums", http.StatusInternalServerError)
		return
	}
	defer albumRows.Close()

	var albums []models.Album
	for albumRows.Next() {
		var album models.Album
		err := albumRows.Scan(&album.ID, &album.MusicianID, &album.Title, &album.ReleaseDate,
			&album.CoverPath, &album.GenreID, &album.Description)
		if err != nil {
			http.Error(w, "Error scanning album", http.StatusInternalServerError)
			return
		}
		albums = append(albums, album)
	}

	// Получаем треки музыканта
	tracksQuery := `
	SELECT id, musician_id, album_id, title, duration, file_path, genre_id, stream_count, visibility
	FROM track
	WHERE musician_id = ?
	`
	trackRows, err := h.DB.Query(tracksQuery, musicianID)
	if err != nil {
		http.Error(w, "Error getting tracks", http.StatusInternalServerError)
		return
	}
	defer trackRows.Close()

	var tracks []models.Track
	for trackRows.Next() {
		var track models.Track
		err := trackRows.Scan(&track.ID, &track.MusicianID, &track.AlbumID, &track.Title, &track.Duration,
			&track.FilePath, &track.GenreID, &track.StreamCount, &track.Visibility)
		if err != nil {
			http.Error(w, "Error scanning track", http.StatusInternalServerError)
			return
		}
		tracks = append(tracks, track)
	}

	type MusicianResponse struct {
		Musician models.Musician `json:"musician"`
		Albums   []models.Album  `json:"albums"`
		Tracks   []models.Track  `json:"tracks"`
	}

	response := MusicianResponse{
		Musician: musician,
		Albums:   albums,
		Tracks:   tracks,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GET /musician/{id}/popular-tracks
func (h *MusicianHandler) GetPopularTracks(w http.ResponseWriter, r *http.Request) {
	_ = r.Context().Value(middleware.ContextUserIDKey).(string)

	musicianID := mux.Vars(r)["id"]

	query := `
	SELECT id, musician_id, album_id, title, duration, file_path, genre_id, stream_count, visibility
	FROM track
	WHERE musician_id = ?
	ORDER BY stream_count DESC
	LIMIT 10;
	`
	rows, err := h.DB.Query(query, musicianID)
	if err != nil {
		http.Error(w, "Error fetching popular tracks", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var tracks []models.Track
	for rows.Next() {
		var t models.Track
		err := rows.Scan(&t.ID, &t.MusicianID, &t.AlbumID, &t.Title, &t.Duration,
			&t.FilePath, &t.GenreID, &t.StreamCount, &t.Visibility)
		if err != nil {
			http.Error(w, "Error scanning track", http.StatusInternalServerError)
			return
		}
		tracks = append(tracks, t)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tracks)
}
