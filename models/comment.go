package models

import "time"

type CommentResponse struct {
	ID        string        `json:"id"`
	Text      string        `json:"text"`
	CreatedAt time.Time     `json:"createdAt"`
	User      CommentAuthor `json:"user"`
}

type CommentAuthor struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatarUrl"`
}

type CreateCommentRequest struct {
	Text string `json:"text"`
}

type TrackComment struct {
	TrackID   string    `bson:"track_id"`
	UserID    string    `bson:"user_id"`
	Comment   string    `bson:"comment"`
	CreatedAt time.Time `bson:"created_at"`
}
