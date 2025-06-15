package models

type Musician struct {
	MusicianPage
	UserID string `json:"userId"`
}

type MusicianPage struct {
	ID               string         `json:"id"`
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
	Tracks      []string `json:"tracks"`
	Description string   `json:"description"`
}
