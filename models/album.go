package models

type Album struct {
	ID          string `json:"id"`
	MusicianID  string `json:"musician_id"`
	Title       string `json:"title"`
	ReleaseDate string `json:"release_date"`
	CoverPath   string `json:"cover_path"`
	GenreID     int    `json:"genre_id"`
	Description string `json:"description"`
}

type RecommendedAlbum struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	ArtistID    string `json:"artistId"`
	ArtistName  string `json:"artistName"`
	CoverUrl    string `json:"coverUrl"`
	Year        int    `json:"year"`
	Description string `json:"description"`
}

type AlbumPageResponse struct {
	ID              string `json:"id"`
	Title           string `json:"title"`
	Year            int    `json:"year"`
	CoverURL        string `json:"coverUrl"`
	Tracks          []int  `json:"tracks"`
	Description     string `json:"description"`
	ArtistID        string `json:"artistId"`
	ArtistName      string `json:"artistName"`
	ArtistAvatarURL string `json:"artistAvatarUrl"`
}
