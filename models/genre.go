package models

type Genre struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type UserGenresRequest struct {
	UserID   string `json:"user_id"`
	GenreIDs []int  `json:"genres_ids"`
}
