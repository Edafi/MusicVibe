package models

type Musician struct {
	ID         string `json:"id"`
	UserID     string `json:"user_id"`
	Name       string `json:"name"`
	AvatarPath string `json:"avatar_path"`
}
