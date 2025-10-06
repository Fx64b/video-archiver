package testutil

import (
	"database/sql"
	"testing"
	"time"
	"video-archiver/internal/domain"

	_ "github.com/mattn/go-sqlite3"
)

// CreateTestDB creates an in-memory SQLite database for testing
func CreateTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}

	// Create tables schema
	schema := `
	CREATE TABLE IF NOT EXISTS jobs (
		job_id TEXT PRIMARY KEY,
		url TEXT NOT NULL,
		status TEXT NOT NULL,
		progress REAL NOT NULL,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL
	);

	CREATE TABLE IF NOT EXISTS videos (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		job_id TEXT UNIQUE NOT NULL,
		title TEXT NOT NULL,
		metadata_json TEXT NOT NULL,
		FOREIGN KEY(job_id) REFERENCES jobs(job_id)
	);

	CREATE TABLE IF NOT EXISTS playlists (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		job_id TEXT UNIQUE NOT NULL,
		title TEXT NOT NULL,
		metadata_json TEXT NOT NULL,
		FOREIGN KEY(job_id) REFERENCES jobs(job_id)
	);

	CREATE TABLE IF NOT EXISTS channels (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		job_id TEXT UNIQUE NOT NULL,
		name TEXT NOT NULL,
		metadata_json TEXT NOT NULL,
		FOREIGN KEY(job_id) REFERENCES jobs(job_id)
	);

	CREATE TABLE IF NOT EXISTS video_memberships (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		video_job_id TEXT NOT NULL,
		parent_job_id TEXT NOT NULL,
		membership_type TEXT NOT NULL,
		FOREIGN KEY(video_job_id) REFERENCES jobs(job_id),
		FOREIGN KEY(parent_job_id) REFERENCES jobs(job_id),
		UNIQUE(video_job_id, parent_job_id)
	);
	`

	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	return db
}

// CreateTestJob creates a test job with default values
func CreateTestJob(id, url string) *domain.Job {
	return &domain.Job{
		ID:        id,
		URL:       url,
		Status:    domain.JobStatusPending,
		Progress:  0,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// CreateTestVideoMetadata creates test video metadata
func CreateTestVideoMetadata() *domain.VideoMetadata {
	return &domain.VideoMetadata{
		ID:                "test-video-id",
		Title:             "Test Video",
		Description:       "Test Description",
		Thumbnail:         "https://example.com/thumb.jpg",
		Duration:          300,
		DurationString:    "5:00",
		ViewCount:         1000,
		LikeCount:         50,
		CommentCount:      10,
		Channel:           "Test Channel",
		ChannelID:         "test-channel-id",
		ChannelURL:        "https://youtube.com/c/testchannel",
		ChannelFollowers:  5000,
		ChannelIsVerified: true,
		Uploader:          "Test Uploader",
		UploaderID:        "test-uploader-id",
		UploaderURL:       "https://youtube.com/u/testuploader",
		Tags:              []string{"tag1", "tag2"},
		Categories:        []string{"Entertainment"},
		UploadDate:        "20240101",
		FileSize:          1024000,
		Format:            "1080p",
		Extension:         "mp4",
		Type:              "video",
	}
}

// CreateTestPlaylistMetadata creates test playlist metadata
func CreateTestPlaylistMetadata() *domain.PlaylistMetadata {
	return &domain.PlaylistMetadata{
		ID:               "test-playlist-id",
		Title:            "Test Playlist",
		Description:      "Test Playlist Description",
		UploaderID:       "test-uploader-id",
		UploaderURL:      "https://youtube.com/u/testuploader",
		ChannelID:        "test-channel-id",
		Channel:          "Test Channel",
		ChannelURL:       "https://youtube.com/c/testchannel",
		ChannelFollowers: 5000,
		ItemCount:        3,
		ViewCount:        10000,
		Items: []domain.PlaylistItem{
			{
				ID:             "video1",
				Title:          "Video 1",
				Duration:       300,
				DurationString: "5:00",
			},
			{
				ID:             "video2",
				Title:          "Video 2",
				Duration:       400,
				DurationString: "6:40",
			},
			{
				ID:             "video3",
				Title:          "Video 3",
				Duration:       500,
				DurationString: "8:20",
			},
		},
		Type: "playlist",
	}
}

// CreateTestChannelMetadata creates test channel metadata
func CreateTestChannelMetadata() *domain.ChannelMetadata {
	return &domain.ChannelMetadata{
		ID:               "test-channel-id",
		Channel:          "Test Channel",
		URL:              "https://youtube.com/c/testchannel",
		Description:      "Test Channel Description",
		ChannelFollowers: 10000,
		PlaylistCount:    5,
		VideoCount:       100,
		TotalStorage:     1024000000,
		TotalViews:       500000,
		RecentVideos: []domain.PlaylistItem{
			{
				ID:             "recent1",
				Title:          "Recent Video 1",
				Duration:       300,
				DurationString: "5:00",
			},
		},
		Type: "channel",
	}
}

// MockJobRepository is a mock implementation of domain.JobRepository for testing
type MockJobRepository struct {
	jobs     map[string]*domain.Job
	metadata map[string]domain.Metadata
	parents  map[string][]*domain.JobWithMetadata
	videos   map[string][]*domain.JobWithMetadata
}

// NewMockJobRepository creates a new mock repository
func NewMockJobRepository() *MockJobRepository {
	return &MockJobRepository{
		jobs:     make(map[string]*domain.Job),
		metadata: make(map[string]domain.Metadata),
		parents:  make(map[string][]*domain.JobWithMetadata),
		videos:   make(map[string][]*domain.JobWithMetadata),
	}
}

func (m *MockJobRepository) Create(job *domain.Job) error {
	m.jobs[job.ID] = job
	return nil
}

func (m *MockJobRepository) Update(job *domain.Job) error {
	if _, exists := m.jobs[job.ID]; !exists {
		return sql.ErrNoRows
	}
	m.jobs[job.ID] = job
	return nil
}

func (m *MockJobRepository) GetByID(id string) (*domain.Job, error) {
	job, exists := m.jobs[id]
	if !exists {
		return nil, sql.ErrNoRows
	}
	return job, nil
}

func (m *MockJobRepository) GetRecent(limit int) ([]*domain.Job, error) {
	jobs := make([]*domain.Job, 0, len(m.jobs))
	for _, job := range m.jobs {
		jobs = append(jobs, job)
	}
	if len(jobs) > limit {
		jobs = jobs[:limit]
	}
	return jobs, nil
}

func (m *MockJobRepository) StoreMetadata(jobID string, metadata domain.Metadata) error {
	m.metadata[jobID] = metadata
	return nil
}

func (m *MockJobRepository) GetJobWithMetadata(jobID string) (*domain.JobWithMetadata, error) {
	job, exists := m.jobs[jobID]
	if !exists {
		return nil, sql.ErrNoRows
	}
	return &domain.JobWithMetadata{
		Job:      job,
		Metadata: m.metadata[jobID],
	}, nil
}

func (m *MockJobRepository) GetRecentWithMetadata(limit int) ([]*domain.JobWithMetadata, error) {
	result := make([]*domain.JobWithMetadata, 0, len(m.jobs))
	for id, job := range m.jobs {
		result = append(result, &domain.JobWithMetadata{
			Job:      job,
			Metadata: m.metadata[id],
		})
	}
	if len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}

func (m *MockJobRepository) GetAllJobsWithMetadata() ([]*domain.JobWithMetadata, error) {
	result := make([]*domain.JobWithMetadata, 0, len(m.jobs))
	for id, job := range m.jobs {
		result = append(result, &domain.JobWithMetadata{
			Job:      job,
			Metadata: m.metadata[id],
		})
	}
	return result, nil
}

func (m *MockJobRepository) GetJobs() ([]*domain.Job, error) {
	jobs := make([]*domain.Job, 0, len(m.jobs))
	for _, job := range m.jobs {
		jobs = append(jobs, job)
	}
	return jobs, nil
}

func (m *MockJobRepository) CountVideos() (int, error) {
	count := 0
	for _, meta := range m.metadata {
		if _, ok := meta.(*domain.VideoMetadata); ok {
			count++
		}
	}
	return count, nil
}

func (m *MockJobRepository) CountPlaylists() (int, error) {
	count := 0
	for _, meta := range m.metadata {
		if _, ok := meta.(*domain.PlaylistMetadata); ok {
			count++
		}
	}
	return count, nil
}

func (m *MockJobRepository) CountChannels() (int, error) {
	count := 0
	for _, meta := range m.metadata {
		if _, ok := meta.(*domain.ChannelMetadata); ok {
			count++
		}
	}
	return count, nil
}

func (m *MockJobRepository) GetMetadataByType(contentType string, page int, limit int, sortBy string, order string) ([]*domain.JobWithMetadata, int, error) {
	result := make([]*domain.JobWithMetadata, 0)
	for id, meta := range m.metadata {
		var matches bool
		switch contentType {
		case "videos":
			_, matches = meta.(*domain.VideoMetadata)
		case "playlists":
			_, matches = meta.(*domain.PlaylistMetadata)
		case "channels":
			_, matches = meta.(*domain.ChannelMetadata)
		}
		if matches {
			result = append(result, &domain.JobWithMetadata{
				Job:      m.jobs[id],
				Metadata: meta,
			})
		}
	}
	return result, len(result), nil
}

func (m *MockJobRepository) AddVideoToParent(videoJobID, parentJobID, membershipType string) error {
	if m.videos[parentJobID] == nil {
		m.videos[parentJobID] = make([]*domain.JobWithMetadata, 0)
	}
	m.videos[parentJobID] = append(m.videos[parentJobID], &domain.JobWithMetadata{
		Job:      m.jobs[videoJobID],
		Metadata: m.metadata[videoJobID],
	})

	if m.parents[videoJobID] == nil {
		m.parents[videoJobID] = make([]*domain.JobWithMetadata, 0)
	}
	m.parents[videoJobID] = append(m.parents[videoJobID], &domain.JobWithMetadata{
		Job:      m.jobs[parentJobID],
		Metadata: m.metadata[parentJobID],
	})
	return nil
}

func (m *MockJobRepository) GetVideosForParent(parentJobID string) ([]*domain.JobWithMetadata, error) {
	return m.videos[parentJobID], nil
}

func (m *MockJobRepository) GetParentsForVideo(videoJobID string) ([]*domain.JobWithMetadata, error) {
	return m.parents[videoJobID], nil
}
