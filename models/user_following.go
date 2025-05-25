package models

type UserFollowing struct {
	UserID     string `json:"user_id"`
	MusicianID string `json:"musician_id"`
}
