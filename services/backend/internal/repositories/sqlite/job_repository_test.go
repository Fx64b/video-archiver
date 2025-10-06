package sqlite

import (
	"testing"
	"time"
	"video-archiver/internal/domain"
	"video-archiver/internal/testutil"
)

func TestJobRepository_Create(t *testing.T) {
	db := testutil.CreateTestDB(t)
	defer db.Close()

	repo := NewJobRepository(db)
	job := testutil.CreateTestJob("test-id", "https://youtube.com/watch?v=test")

	err := repo.Create(job)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// Verify job was created
	retrieved, err := repo.GetByID("test-id")
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if retrieved.ID != job.ID {
		t.Errorf("ID = %v, want %v", retrieved.ID, job.ID)
	}
	if retrieved.URL != job.URL {
		t.Errorf("URL = %v, want %v", retrieved.URL, job.URL)
	}
}

func TestJobRepository_Update(t *testing.T) {
	db := testutil.CreateTestDB(t)
	defer db.Close()

	repo := NewJobRepository(db)
	job := testutil.CreateTestJob("test-id", "https://youtube.com/watch?v=test")

	// Create initial job
	err := repo.Create(job)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// Update job
	job.Status = domain.JobStatusInProgress
	job.Progress = 50.0
	time.Sleep(10 * time.Millisecond) // Ensure UpdatedAt changes

	err = repo.Update(job)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	// Verify update
	retrieved, err := repo.GetByID("test-id")
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if retrieved.Status != domain.JobStatusInProgress {
		t.Errorf("Status = %v, want %v", retrieved.Status, domain.JobStatusInProgress)
	}
	if retrieved.Progress != 50.0 {
		t.Errorf("Progress = %v, want %v", retrieved.Progress, 50.0)
	}
}

func TestJobRepository_GetByID(t *testing.T) {
	db := testutil.CreateTestDB(t)
	defer db.Close()

	repo := NewJobRepository(db)

	tests := []struct {
		name    string
		jobID   string
		wantErr bool
	}{
		{
			name:    "existing job",
			jobID:   "test-id",
			wantErr: false,
		},
		{
			name:    "non-existing job",
			jobID:   "non-existent",
			wantErr: true,
		},
	}

	// Create test job
	job := testutil.CreateTestJob("test-id", "https://youtube.com/watch?v=test")
	repo.Create(job)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			retrieved, err := repo.GetByID(tt.jobID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && retrieved.ID != tt.jobID {
				t.Errorf("GetByID() ID = %v, want %v", retrieved.ID, tt.jobID)
			}
		})
	}
}

func TestJobRepository_GetRecent(t *testing.T) {
	db := testutil.CreateTestDB(t)
	defer db.Close()

	repo := NewJobRepository(db)

	// Create multiple jobs
	for i := 1; i <= 10; i++ {
		job := testutil.CreateTestJob("test-id-"+string(rune('0'+i)), "https://youtube.com/watch?v=test"+string(rune('0'+i)))
		time.Sleep(1 * time.Millisecond) // Ensure different timestamps
		repo.Create(job)
	}

	tests := []struct {
		name      string
		limit     int
		wantCount int
	}{
		{
			name:      "get 5 recent",
			limit:     5,
			wantCount: 5,
		},
		{
			name:      "get all",
			limit:     20,
			wantCount: 10,
		},
		{
			name:      "get 1",
			limit:     1,
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jobs, err := repo.GetRecent(tt.limit)
			if err != nil {
				t.Fatalf("GetRecent() error = %v", err)
			}
			if len(jobs) != tt.wantCount {
				t.Errorf("GetRecent() count = %v, want %v", len(jobs), tt.wantCount)
			}
		})
	}
}

func TestJobRepository_StoreVideoMetadata(t *testing.T) {
	db := testutil.CreateTestDB(t)
	defer db.Close()

	repo := NewJobRepository(db)
	job := testutil.CreateTestJob("test-id", "https://youtube.com/watch?v=test")
	repo.Create(job)

	metadata := testutil.CreateTestVideoMetadata()
	err := repo.StoreMetadata("test-id", metadata)
	if err != nil {
		t.Fatalf("StoreMetadata() error = %v", err)
	}

	// Verify metadata was stored
	retrieved, err := repo.GetJobWithMetadata("test-id")
	if err != nil {
		t.Fatalf("GetJobWithMetadata() error = %v", err)
	}

	videoMeta, ok := retrieved.Metadata.(*domain.VideoMetadata)
	if !ok {
		t.Fatalf("Metadata is not VideoMetadata")
	}
	if videoMeta.Title != metadata.Title {
		t.Errorf("Title = %v, want %v", videoMeta.Title, metadata.Title)
	}
	if videoMeta.ID != metadata.ID {
		t.Errorf("ID = %v, want %v", videoMeta.ID, metadata.ID)
	}
}

func TestJobRepository_StorePlaylistMetadata(t *testing.T) {
	db := testutil.CreateTestDB(t)
	defer db.Close()

	repo := NewJobRepository(db)
	job := testutil.CreateTestJob("test-id", "https://youtube.com/playlist?list=test")
	repo.Create(job)

	metadata := testutil.CreateTestPlaylistMetadata()
	err := repo.StoreMetadata("test-id", metadata)
	if err != nil {
		t.Fatalf("StoreMetadata() error = %v", err)
	}

	// Verify metadata was stored
	retrieved, err := repo.GetJobWithMetadata("test-id")
	if err != nil {
		t.Fatalf("GetJobWithMetadata() error = %v", err)
	}

	playlistMeta, ok := retrieved.Metadata.(*domain.PlaylistMetadata)
	if !ok {
		t.Fatalf("Metadata is not PlaylistMetadata")
	}
	if playlistMeta.Title != metadata.Title {
		t.Errorf("Title = %v, want %v", playlistMeta.Title, metadata.Title)
	}
	if playlistMeta.ItemCount != metadata.ItemCount {
		t.Errorf("ItemCount = %v, want %v", playlistMeta.ItemCount, metadata.ItemCount)
	}
}

func TestJobRepository_StoreChannelMetadata(t *testing.T) {
	db := testutil.CreateTestDB(t)
	defer db.Close()

	repo := NewJobRepository(db)
	job := testutil.CreateTestJob("test-id", "https://youtube.com/@testchannel")
	repo.Create(job)

	metadata := testutil.CreateTestChannelMetadata()
	err := repo.StoreMetadata("test-id", metadata)
	if err != nil {
		t.Fatalf("StoreMetadata() error = %v", err)
	}

	// Verify metadata was stored
	retrieved, err := repo.GetJobWithMetadata("test-id")
	if err != nil {
		t.Fatalf("GetJobWithMetadata() error = %v", err)
	}

	channelMeta, ok := retrieved.Metadata.(*domain.ChannelMetadata)
	if !ok {
		t.Fatalf("Metadata is not ChannelMetadata")
	}
	if channelMeta.Channel != metadata.Channel {
		t.Errorf("Channel = %v, want %v", channelMeta.Channel, metadata.Channel)
	}
	if channelMeta.VideoCount != metadata.VideoCount {
		t.Errorf("VideoCount = %v, want %v", channelMeta.VideoCount, metadata.VideoCount)
	}
}

func TestJobRepository_CountVideos(t *testing.T) {
	db := testutil.CreateTestDB(t)
	defer db.Close()

	repo := NewJobRepository(db)

	// Create jobs with video metadata
	for i := 1; i <= 3; i++ {
		job := testutil.CreateTestJob("video-"+string(rune('0'+i)), "https://youtube.com/watch?v=test"+string(rune('0'+i)))
		repo.Create(job)
		repo.StoreMetadata(job.ID, testutil.CreateTestVideoMetadata())
	}

	count, err := repo.CountVideos()
	if err != nil {
		t.Fatalf("CountVideos() error = %v", err)
	}
	if count != 3 {
		t.Errorf("CountVideos() = %v, want %v", count, 3)
	}
}

func TestJobRepository_CountPlaylists(t *testing.T) {
	db := testutil.CreateTestDB(t)
	defer db.Close()

	repo := NewJobRepository(db)

	// Create jobs with playlist metadata
	for i := 1; i <= 2; i++ {
		job := testutil.CreateTestJob("playlist-"+string(rune('0'+i)), "https://youtube.com/playlist?list=test"+string(rune('0'+i)))
		repo.Create(job)
		repo.StoreMetadata(job.ID, testutil.CreateTestPlaylistMetadata())
	}

	count, err := repo.CountPlaylists()
	if err != nil {
		t.Fatalf("CountPlaylists() error = %v", err)
	}
	if count != 2 {
		t.Errorf("CountPlaylists() = %v, want %v", count, 2)
	}
}

func TestJobRepository_CountChannels(t *testing.T) {
	db := testutil.CreateTestDB(t)
	defer db.Close()

	repo := NewJobRepository(db)

	// Create jobs with channel metadata
	for i := 1; i <= 1; i++ {
		job := testutil.CreateTestJob("channel-"+string(rune('0'+i)), "https://youtube.com/@channel"+string(rune('0'+i)))
		repo.Create(job)
		repo.StoreMetadata(job.ID, testutil.CreateTestChannelMetadata())
	}

	count, err := repo.CountChannels()
	if err != nil {
		t.Fatalf("CountChannels() error = %v", err)
	}
	if count != 1 {
		t.Errorf("CountChannels() = %v, want %v", count, 1)
	}
}

func TestJobRepository_GetMetadataByType(t *testing.T) {
	db := testutil.CreateTestDB(t)
	defer db.Close()

	repo := NewJobRepository(db)

	// Create test data
	for i := 1; i <= 5; i++ {
		job := testutil.CreateTestJob("video-"+string(rune('0'+i)), "https://youtube.com/watch?v=test"+string(rune('0'+i)))
		repo.Create(job)
		repo.StoreMetadata(job.ID, testutil.CreateTestVideoMetadata())
	}

	tests := []struct {
		name        string
		contentType string
		page        int
		limit       int
		wantCount   int
		wantTotal   int
		wantErr     bool
	}{
		{
			name:        "get first page of videos",
			contentType: "videos",
			page:        1,
			limit:       3,
			wantCount:   3,
			wantTotal:   5,
			wantErr:     false,
		},
		{
			name:        "get second page of videos",
			contentType: "videos",
			page:        2,
			limit:       3,
			wantCount:   2,
			wantTotal:   5,
			wantErr:     false,
		},
		{
			name:        "invalid content type",
			contentType: "invalid",
			page:        1,
			limit:       10,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			items, total, err := repo.GetMetadataByType(tt.contentType, tt.page, tt.limit, "created_at", "desc")
			if (err != nil) != tt.wantErr {
				t.Errorf("GetMetadataByType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if len(items) != tt.wantCount {
				t.Errorf("GetMetadataByType() count = %v, want %v", len(items), tt.wantCount)
			}
			if total != tt.wantTotal {
				t.Errorf("GetMetadataByType() total = %v, want %v", total, tt.wantTotal)
			}
		})
	}
}

func TestJobRepository_AddVideoToParent(t *testing.T) {
	db := testutil.CreateTestDB(t)
	defer db.Close()

	repo := NewJobRepository(db)

	// Create parent playlist job
	parentJob := testutil.CreateTestJob("parent-id", "https://youtube.com/playlist?list=test")
	repo.Create(parentJob)
	repo.StoreMetadata(parentJob.ID, testutil.CreateTestPlaylistMetadata())

	// Create video job
	videoJob := testutil.CreateTestJob("video-id", "https://youtube.com/watch?v=test")
	repo.Create(videoJob)
	repo.StoreMetadata(videoJob.ID, testutil.CreateTestVideoMetadata())

	// Link video to parent
	err := repo.AddVideoToParent("video-id", "parent-id", "playlist")
	if err != nil {
		t.Fatalf("AddVideoToParent() error = %v", err)
	}

	// Verify relationship
	videos, err := repo.GetVideosForParent("parent-id")
	if err != nil {
		t.Fatalf("GetVideosForParent() error = %v", err)
	}
	if len(videos) != 1 {
		t.Errorf("GetVideosForParent() count = %v, want %v", len(videos), 1)
	}

	parents, err := repo.GetParentsForVideo("video-id")
	if err != nil {
		t.Fatalf("GetParentsForVideo() error = %v", err)
	}
	if len(parents) != 1 {
		t.Errorf("GetParentsForVideo() count = %v, want %v", len(parents), 1)
	}
}

func TestJobRepository_GetVideosForParent(t *testing.T) {
	db := testutil.CreateTestDB(t)
	defer db.Close()

	repo := NewJobRepository(db)

	// Create parent
	parentJob := testutil.CreateTestJob("parent-id", "https://youtube.com/playlist?list=test")
	repo.Create(parentJob)
	repo.StoreMetadata(parentJob.ID, testutil.CreateTestPlaylistMetadata())

	// Create multiple videos
	for i := 1; i <= 3; i++ {
		videoJob := testutil.CreateTestJob("video-"+string(rune('0'+i)), "https://youtube.com/watch?v=test"+string(rune('0'+i)))
		repo.Create(videoJob)
		repo.StoreMetadata(videoJob.ID, testutil.CreateTestVideoMetadata())
		repo.AddVideoToParent(videoJob.ID, "parent-id", "playlist")
	}

	videos, err := repo.GetVideosForParent("parent-id")
	if err != nil {
		t.Fatalf("GetVideosForParent() error = %v", err)
	}
	if len(videos) != 3 {
		t.Errorf("GetVideosForParent() count = %v, want %v", len(videos), 3)
	}

	// Verify all are videos
	for _, v := range videos {
		if _, ok := v.Metadata.(*domain.VideoMetadata); !ok {
			t.Errorf("Expected VideoMetadata, got %T", v.Metadata)
		}
	}
}

func TestJobRepository_GetParentsForVideo(t *testing.T) {
	db := testutil.CreateTestDB(t)
	defer db.Close()

	repo := NewJobRepository(db)

	// Create video
	videoJob := testutil.CreateTestJob("video-id", "https://youtube.com/watch?v=test")
	repo.Create(videoJob)
	repo.StoreMetadata(videoJob.ID, testutil.CreateTestVideoMetadata())

	// Create multiple parents (playlist and channel)
	playlistJob := testutil.CreateTestJob("playlist-id", "https://youtube.com/playlist?list=test")
	repo.Create(playlistJob)
	repo.StoreMetadata(playlistJob.ID, testutil.CreateTestPlaylistMetadata())
	repo.AddVideoToParent("video-id", "playlist-id", "playlist")

	channelJob := testutil.CreateTestJob("channel-id", "https://youtube.com/@testchannel")
	repo.Create(channelJob)
	repo.StoreMetadata(channelJob.ID, testutil.CreateTestChannelMetadata())
	repo.AddVideoToParent("video-id", "channel-id", "channel")

	parents, err := repo.GetParentsForVideo("video-id")
	if err != nil {
		t.Fatalf("GetParentsForVideo() error = %v", err)
	}
	if len(parents) != 2 {
		t.Errorf("GetParentsForVideo() count = %v, want %v", len(parents), 2)
	}
}
