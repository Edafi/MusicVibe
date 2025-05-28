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

type TrackResponse struct {
	Track
	Artist    string `json:"artist"`
	CoverPath string `json:"cover_path"`
}

type RecommendedTrack struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	ArtistID   string `json:"artistId"`
	ArtistName string `json:"artistName"`
	ImageURL   string `json:"imageUrl"`
	AudioURL   string `json:"audioUrl"`
	Duration   int    `json:"duration"`
	Plays      int    `json:"plays"`
}

type HomeRecomendedTrack struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	MusicianID   string `json:"artistId"`
	MusicianName string `json:"artistName"`
	ImageURL     string `json:"imageUrl"`
	Plays        int    `json:"plays"`
}
