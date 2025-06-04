package models

type TrackResponse struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	ArtistID   string `json:"artistId"`
	ArtistName string `json:"artistName"`
	ImageURL   string `json:"imageUrl"`
	AudioURL   string `json:"audioUrl"`
	Duration   int    `json:"duration"`
	Plays      int    `json:"plays"`
	Visibility string `json:"visibility"`
}
