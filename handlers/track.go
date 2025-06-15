package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path/filepath"

	"github.com/Edafi/MusicVibe/models"
	"github.com/gorilla/mux"
)

type TrackHandler struct {
	DB *sql.DB
}

func (handler *TrackHandler) GetTrack(response http.ResponseWriter, request *http.Request) {
	trackID := mux.Vars(request)["id"]

	var track models.TrackResponse

	query := `
		SELECT 
		t.id, t.title, t.musician_id, m.name, a.cover_path, t.file_path, t.duration,
		t.stream_count, t.visibility
		FROM track t
		JOIN musician m ON t.musician_id = m.id
		JOIN album a ON t.album_id = a.id
		WHERE t.id = ?
	`

	err := handler.DB.QueryRow(query, trackID).Scan(&track.ID, &track.Title, &track.ArtistID,
		&track.ArtistName, &track.ImageURL, &track.AudioURL, &track.Duration, &track.Plays,
		&track.Visibility,
	)
	if err != nil {
		log.Println("GetTrack - Error fetching track: ", err)
		http.Error(response, "Track not found", http.StatusNotFound)
		return
	}
	baseURL := "http://37.46.130.29:8080"
	track.AudioURL = fmt.Sprintf("%s/media/audio/%s", baseURL, track.ID)
	track.ImageURL = fmt.Sprintf("%s/media/image/%s", baseURL, filepath.Base(track.ImageURL))

	response.Header().Set("Content-Type", "application/json")
	json.NewEncoder(response).Encode(track)
}
