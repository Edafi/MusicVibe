package models

type Musician struct {
	ID         string `json:"id"`
	UserID     string `json:"user_id"`
	Name       string `json:"name"`
	AvatarPath string `json:"avatar_path"`
}

type MusicianResponse struct {
	ID               string         `json:"id"`
	UserID           string         `json:"userId"`
	Name             string         `json:"name"`
	Email            string         `json:"email"`
	AvatarPath       string         `json:"avatarUrl"`
	BackgroundPath   string         `json:"backgroundUrl"`
	Description      string         `json:"description"`
	Genres           []string       `json:"genres"`
	Auditions        int            `json:"auditions"`
	HasCompleteSetup bool           `json:"hasCompletedSetup"`
	SocialLinks      []SocialLink   `json:"socialLinks"`
	Albums           []AlbumPreview `json:"albums"`
}

type SocialLink struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type AlbumPreview struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Year        int      `json:"year"`
	CoverUrl    string   `json:"coverUrl"`
	Tracks      []string `json:"tracks"` // или []int
	Description string   `json:"description"`
}
