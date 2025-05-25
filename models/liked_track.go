package models

type LikedTrack struct {
	UserID  string `json:"user_id"`
	TrackID string `json:"track_id"`
}
