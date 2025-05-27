package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/Edafi/MusicVibe/middleware"
	"github.com/Edafi/MusicVibe/models"
)

type ChartHandler struct {
	DB *sql.DB
}

func (h *ChartHandler) GetChartTracks(w http.ResponseWriter, r *http.Request) {

	_ = r.Context().Value(middleware.ContextUserIDKey).(string)

	query := `
	SELECT 
		t.id, t.musician_id, t.album_id, t.title, t.duration,
		t.file_path, t.genre_id, t.stream_count, t.visibility,
		a.cover_path,
		m.name AS musician_name
	FROM track t
	JOIN album a ON t.album_id = a.id
	JOIN musician m ON t.musician_id = m.id
	ORDER BY t.stream_count DESC
	LIMIT 50;
	`

	rows, err := h.DB.Query(query)
	if err != nil {
		http.Error(w, "Database query error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type ChartTrack struct {
		models.Track
		CoverPath    string `json:"cover_path"`
		MusicianName string `json:"musician_name"`
	}

	var chart []ChartTrack
	for rows.Next() {
		var ct ChartTrack
		err := rows.Scan(
			&ct.ID, &ct.MusicianID, &ct.AlbumID, &ct.Title, &ct.Duration,
			&ct.FilePath, &ct.GenreID, &ct.StreamCount, &ct.Visibility,
			&ct.CoverPath,
			&ct.MusicianName,
		)
		if err != nil {
			http.Error(w, "Scan error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		chart = append(chart, ct)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(chart)
}
