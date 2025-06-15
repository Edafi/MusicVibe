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
func (handler *MusicianHandler) GetMusicians(response http.ResponseWriter, request *http.Request) {
	userID, ok := request.Context().Value(middleware.ContextUserIDKey).(string)
	if !ok || userID == "" {
		http.Error(response, "Unauthorized", http.StatusUnauthorized)
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
		http.Error(response, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var musicians []models.Musician = make([]models.Musician, 0)

	for rows.Next() {
		var m models.Musician
		err := rows.Scan(&m.ID, &m.UserID, &m.Name, &m.Email, &m.AvatarPath, &m.BackgroundPath, &m.Description, &m.HasCompleteSetup)
		if err != nil {
			http.Error(response, err.Error(), http.StatusInternalServerError)
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
		MusicianIDs []string `json:"musicianIds"`
	}

	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		log.Println("error decoding json:", err)
		http.Error(response, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if len(payload.MusicianIDs) == 0 {
		http.Error(response, "No musicians selected", http.StatusBadRequest)
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
		log.Println("Following musicianID:", musicianID)
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

// GET /musician/&{id}
func (handler *MusicianHandler) GetMusician(response http.ResponseWriter, request *http.Request) {
	musicianID := mux.Vars(request)["id"]

	var musician models.Musician

	// Получаем данные из musician и user
	err := handler.DB.QueryRow(`
		SELECT m.id, m.user_id, m.name, u.email, u.avatar_path, u.background_path, u.description, u.has_complete_setup
		FROM musician m
		JOIN user u ON m.user_id = u.id
		WHERE m.id = ?
	`, musicianID).Scan(
		&musician.ID, &musician.UserID, &musician.Name, &musician.Email,
		&musician.AvatarPath, &musician.BackgroundPath, &musician.Description, &musician.HasCompleteSetup,
	)
	if err != nil {
		log.Println("GetMusician - Musician not found: ", err)
		http.Error(response, "Musician not found", http.StatusNotFound)
		return
	}

	// Получаем жанры
	var genres []string = make([]string, 0)
	rows, err := handler.DB.Query(`
		SELECT g.name
		FROM genre g
		JOIN musician_genre mg ON g.id = mg.genre_id
		WHERE mg.musician_id = ?
	`, musicianID)
	if err != nil {
		log.Println("GetMusician - Error getting genres: ", err)
		http.Error(response, "Error getting genres", http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var genre string
		rows.Scan(&genre)
		genres = append(genres, genre)
	}
	musician.Genres = genres

	// Получаем социальные сети пользователя
	var socialLinks []models.SocialLink = make([]models.SocialLink, 0)
	rows, err = handler.DB.Query(`
		SELECT sn.name, usn.profile_url
		FROM user_social_network usn
		JOIN social_network sn ON usn.social_network_id = sn.id
		WHERE usn.user_id = ?
	`, musician.UserID)
	if err != nil {
		log.Println("GetMusician - Error getting social links: ", err)
		http.Error(response, "Error getting social links", http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var link models.SocialLink
		rows.Scan(&link.Name, &link.URL)
		socialLinks = append(socialLinks, link)
	}
	musician.SocialLinks = socialLinks

	// Получаем количество прослушиваний
	err = handler.DB.QueryRow(`
		SELECT COALESCE(SUM(stream_count), 0)
		FROM track
		WHERE musician_id = ? AND track.visibility = 'public'
	`, musicianID).Scan(&musician.Auditions)
	if err != nil {
		log.Println("GetMusician - Error getting auditions: ", err)
		http.Error(response, "Error getting auditions", http.StatusInternalServerError)
		return
	}

	// Получаем альбомы
	albumRows, err := handler.DB.Query(`
		SELECT id, title, YEAR(release_date), cover_path, description
		FROM album
		WHERE musician_id = ?
	`, musicianID)
	if err != nil {
		log.Println("GetMusician - Error getting albums: ", err)
		http.Error(response, "Error getting albums", http.StatusInternalServerError)
		return
	}
	defer albumRows.Close()

	var albums []models.AlbumPreview = make([]models.AlbumPreview, 0)
	for albumRows.Next() {
		var album models.AlbumPreview
		var albumID string
		err := albumRows.Scan(&albumID, &album.Title, &album.Year, &album.CoverUrl, &album.Description)
		if err != nil {
			log.Println("GetMusician - Error scanning album: ", err)
			http.Error(response, "Error scanning album", http.StatusInternalServerError)
			return
		}
		album.ID = albumID

		// Получаем ID треков альбома
		trackIDs := []string{}
		trackRows, err := handler.DB.Query(`
			SELECT id FROM track WHERE album_id = ? AND track.visibility = 'public'
		`, albumID)
		if err != nil {
			log.Println("GetMusician - Error getting tracks: ", err)
			http.Error(response, "Error getting tracks", http.StatusInternalServerError)
			return
		}
		for trackRows.Next() {
			var trackID string
			trackRows.Scan(&trackID)
			trackIDs = append(trackIDs, trackID)
		}
		trackRows.Close()

		album.Tracks = trackIDs
		albums = append(albums, album)
	}
	musician.Albums = albums

	// Возвращаем JSON
	response.Header().Set("Content-Type", "application/json")
	json.NewEncoder(response).Encode(musician)
}

// GET /musician/{id}/popular-tracks
func (handler *MusicianHandler) GetPopularTracks(response http.ResponseWriter, request *http.Request) {
	_ = request.Context().Value(middleware.ContextUserIDKey).(string)

	musicianID := mux.Vars(request)["id"]

	query := `
	SELECT track.id, track.musician_id, track.title, track.duration, track.file_path, track.stream_count, m.name, a.cover_path, track.visibility
	FROM track
	JOIN musician m ON track.musician_id = m.id
	JOIN album a ON track.album_id = a.id
	WHERE track.musician_id = ? AND track.visibility = 'public'
	ORDER BY stream_count DESC
	LIMIT 10;
	`
	rows, err := handler.DB.Query(query, musicianID)
	if err != nil {
		log.Println("GetPopularTracks - Error fetching popular tracks: ", err)
		http.Error(response, "Error fetching popular tracks", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var tracks []models.TrackResponse = make([]models.TrackResponse, 0)
	for rows.Next() {
		var t models.TrackResponse
		err := rows.Scan(&t.ID, &t.ArtistID, &t.Title, &t.Duration, &t.AudioURL,
			&t.Plays, &t.ArtistName, &t.ImageURL, &t.Visibility)
		if err != nil {
			log.Println("GetPopularTracks - Error scanning track: ", err)
			http.Error(response, "Error scanning track", http.StatusInternalServerError)
			return
		}
		tracks = append(tracks, t)
	}

	response.Header().Set("Content-Type", "application/json")
	json.NewEncoder(response).Encode(tracks)
}
