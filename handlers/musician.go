package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"github.com/Edafi/MusicVibe/middleware"
	"github.com/Edafi/MusicVibe/models"
)

type MusicianHandler struct {
	DB *sql.DB
}

// --------------------- GET /musicians --------------------- //
func (handler *MusicianHandler) GetMusicians(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.ContextUserIDKey).(string)
	if !ok || userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Получение ID музыкантов, которых должен видеть пользователь
	musicianQuery := `
	SELECT DISTINCT m.id, m.user_id, m.name, u.email, u.avatar_path, u.background_path, 
	       u.description, u.has_complete_setup
	FROM musician m
	JOIN musician_genre mg ON m.id = mg.musician_id
	JOIN user_genre ug ON mg.genre_id = ug.genre_id
	JOIN user u ON m.user_id = u.id
	WHERE ug.user_id = ?
	`

	rows, err := handler.DB.Query(musicianQuery, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var musicians []models.MusicianResponse

	for rows.Next() {
		var m models.MusicianResponse
		err := rows.Scan(&m.ID, &m.UserID, &m.Name, &m.Email, &m.AvatarPath, &m.BackgroundPath, &m.Description, &m.HasCompleteSetup)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Получаем жанры
		genreRows, _ := handler.DB.Query(`
			SELECT g.name
			FROM genre g
			JOIN musician_genre mg ON mg.genre_id = g.id
			WHERE mg.musician_id = ?`, m.ID)
		for genreRows.Next() {
			var genre string
			genreRows.Scan(&genre)
			m.Genres = append(m.Genres, genre)
		}
		genreRows.Close()

		// Получаем соцсети
		socialRows, _ := handler.DB.Query(`
			SELECT sn.name, usn.profile_url
			FROM social_network sn
			JOIN user_social_network usn ON usn.social_network_id = sn.id
			WHERE usn.user_id = ?`, m.UserID)
		for socialRows.Next() {
			var sl models.SocialLink
			socialRows.Scan(&sl.Name, &sl.URL)
			m.SocialLinks = append(m.SocialLinks, sl)
		}
		socialRows.Close()

		// Получаем альбомы
		albumRows, _ := handler.DB.Query(`
			SELECT a.id, a.title, a.release_date, a.cover_path, a.description
			FROM album a
			WHERE a.musician_id = ?`, m.ID)
		for albumRows.Next() {
			var album models.AlbumPreview
			var releaseDate string
			albumRows.Scan(&album.ID, &album.Title, &releaseDate, &album.CoverUrl, &album.Description)

			// Год релиза
			if t, err := time.Parse("2006-01-02", releaseDate); err == nil {
				album.Year = t.Year()
			}

			// Получаем треки
			trackRows, _ := handler.DB.Query(`
				SELECT id FROM track WHERE album_id = ?`, album.ID)
			for trackRows.Next() {
				var trackID string
				trackRows.Scan(&trackID)
				album.Tracks = append(album.Tracks, trackID)
			}
			trackRows.Close()

			m.Albums = append(m.Albums, album)
		}
		albumRows.Close()

		musicians = append(musicians, m)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(musicians)
}

// --------------------- POST /user/following --------------------- //
func (handler *MusicianHandler) PostUserFollowing(response http.ResponseWriter, request *http.Request) {
	userID, ok := request.Context().Value(middleware.ContextUserIDKey).(string)
	if !ok || userID == "" {
		http.Error(response, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var payload struct {
		MusicianIDs []string `json:"selectedIds"`
	}

	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		log.Println("error decoding json:", err)
		http.Error(response, "Invalid JSON", http.StatusBadRequest)
		return
	}

	tx, err := handler.DB.Begin()
	if err != nil {
		log.Println("Failed to begin transaction:", err)
		http.Error(response, "Failed to begin transaction", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("INSERT INTO user_following (user_id, musician_id) VALUES (?, ?)")
	if err != nil {
		log.Println("Failed to prepare insert:", err)
		http.Error(response, "Failed to prepare insert", http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	for _, musicianID := range payload.MusicianIDs {
		if _, err := stmt.Exec(userID, musicianID); err != nil {
			log.Println("Failed to insert follow:", err)
			http.Error(response, "Failed to insert follow", http.StatusInternalServerError)
			return
		}
	}

	updateStmt, err := tx.Prepare("UPDATE user SET has_complete_setup = 1 WHERE id = ?")

	if err != nil {
		log.Println("Failed to prepare update:", err)
		http.Error(response, "Failed to prepare update", http.StatusInternalServerError)
		return
	}
	defer updateStmt.Close()

	if _, err := updateStmt.Exec(userID); err != nil {
		log.Println("Failed to update user setup:", err)
		http.Error(response, "Failed to update user setup", http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		log.Println("Failed to commit transaction:", err)
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
