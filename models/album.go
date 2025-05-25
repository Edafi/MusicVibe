package models

type Album struct {
	ID          string `json:"id"`
	MusicianID  string `json:"musician_id"`
	Title       string `json:"title"`
	ReleaseDate string `json:"release_date"`
	CoverPath   string `json:"cover_path"`
	GenreID     int    `json:"genre_id"`
	Description string `json:"description"`
}
