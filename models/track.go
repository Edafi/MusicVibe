package models

// TODO: Добавить путь до альбома трека
type Track struct {
	ID          string `json:"id"`
	MusicianID  string `json:"artistId"`
	AlbumID     string `json:"album_id"`
	Title       string `json:"title"`
	Duration    int    `json:"duration"`
	FilePath    string `json:"file_path"`
	GenreID     int    `json:"genre_id"`
	StreamCount int    `json:"stream_count"`
	Visibility  string `json:"visibility"`
}

type TrackResponse struct {
	ID         int    `json:"id"`
	Title      string `json:"title"`
	ArtistID   int    `json:"artistId"`
	ArtistName string `json:"artistName"`
	ImageURL   string `json:"imageUrl"`
	AudioURL   string `json:"audioUrl"`
	Duration   int    `json:"duration"`
	Plays      int    `json:"plays"`
	Visibility string `json:"visibility"`
}
