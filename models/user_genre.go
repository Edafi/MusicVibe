package models

type UserGenre struct {
	UserID  string `json:"user_id"`
	GenreID int    `json:"genre_id"`
}

// вторая страница
type UserGenresRequest struct {
	GenreIDs []int `json:"genre_ids"`
}
