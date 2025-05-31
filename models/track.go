package models

// TODO: Добавить путь до альбома трека
type Track struct {
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
