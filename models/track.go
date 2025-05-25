package models

type Track struct {
	ID          string `json:"id"`
	MusicianID  string `json:"musician_id"`
	AlbumID     string `json:"album_id"`
	Title       string `json:"title"`
	Duration    int    `json:"duration"`
	FilePath    string `json:"file_path"`
	GenreID     int    `json:"genre_id"`
	StreamCount int    `json:"stream_count"`
	Visibility  string `json:"visibility"`
}
