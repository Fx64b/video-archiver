package domain

import "time"

// Collection is a user-defined set of downloaded videos — a custom playlist.
// Unlike playlists and channels, which mirror content on the source platform,
// collections are curated locally and can mix videos from any source. The
// tools section treats a collection like a playlist: one collection ID
// expands into its member videos.
type Collection struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	// VideoCount and Thumbnail are derived from the members when listing;
	// Thumbnail is the first member's thumbnail.
	VideoCount int    `json:"video_count"`
	Thumbnail  string `json:"thumbnail,omitempty"`
}

//tygo:ignore
type CollectionRepository interface {
	Create(collection *Collection) error
	Update(collection *Collection) error
	Delete(id string) error
	GetByID(id string) (*Collection, error)
	List() ([]*Collection, error)
	// AddVideos appends the given video jobs to the collection, skipping IDs
	// that are already members or do not reference a downloaded video.
	AddVideos(collectionID string, videoJobIDs []string) error
	RemoveVideo(collectionID string, videoJobID string) error
	// GetVideos returns the member videos in collection order.
	GetVideos(collectionID string) ([]*JobWithMetadata, error)
	// ListForVideo returns the IDs of the collections containing the video.
	ListForVideo(videoJobID string) ([]string, error)
}
