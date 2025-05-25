package models

type Playlist struct {
	ID           string `json:"id"`
	UserID       string `json:"user_id"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	IsPublic     bool   `json:"is_public"`
	CoverPath    string `json:"cover_path"`
	CreationDate string `json:"creation_date"`
}
