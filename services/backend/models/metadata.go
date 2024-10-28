package models

type VideoMetadata struct {
	ID        string  `json:"id"`
	Title     string  `json:"title"`
	Uploader  string  `json:"uploader"`
	Filesize  int64   `json:"filesize"`
	Duration  float64 `json:"duration"`
	Format    string  `json:"format"`
	Thumbnail string  `json:"thumbnail"`
}

type PlaylistMetadata struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
}
