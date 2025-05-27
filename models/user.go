package models

type User struct {
	ID           string `json:"id"`
	Email        string `json:"email"`
	PasswdHash   string `json:"passwd_hash"`
	Role         string `json:"role"`
	Username     string `json:"username"`
	AvatarPath   string `json:"avatar_path"`
	CreationDate string `json:"creation_date"`
}
