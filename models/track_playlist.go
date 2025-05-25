package models

type TrackPlaylist struct {
	PlaylistID string `json:"playlist_id"`
	TrackID    string `json:"track_id"`
	Position   int    `json:"position"`
}
