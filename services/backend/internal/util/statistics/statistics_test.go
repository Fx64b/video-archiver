package statistics

import (
	"os"
	"path/filepath"
	"testing"
	"video-archiver/internal/testutil"
)

func TestGetStatistics(t *testing.T) {
	mockRepo := testutil.NewMockJobRepository()
	tmpDir := t.TempDir()

	// Create test data
	for i := 1; i <= 3; i++ {
		job := testutil.CreateTestJob("video-"+string(rune('0'+i)), "https://youtube.com/watch?v=test"+string(rune('0'+i)))
		mockRepo.Create(job)
		mockRepo.StoreMetadata(job.ID, testutil.CreateTestVideoMetadata())
	}

	playlistJob := testutil.CreateTestJob("playlist-1", "https://youtube.com/playlist?list=test")
	mockRepo.Create(playlistJob)
	mockRepo.StoreMetadata(playlistJob.ID, testutil.CreateTestPlaylistMetadata())

	channelJob := testutil.CreateTestJob("channel-1", "https://youtube.com/@testchannel")
	mockRepo.Create(channelJob)
	mockRepo.StoreMetadata(channelJob.ID, testutil.CreateTestChannelMetadata())

	stats, err := GetStatistics(mockRepo, tmpDir)
	if err != nil {
		t.Fatalf("GetStatistics() error = %v", err)
	}

	// Verify counts
	if stats.TotalJobs != 5 {
		t.Errorf("TotalJobs = %v, want %v", stats.TotalJobs, 5)
	}
	if stats.TotalVideos != 3 {
		t.Errorf("TotalVideos = %v, want %v", stats.TotalVideos, 3)
	}
	if stats.TotalPlaylists != 1 {
		t.Errorf("TotalPlaylists = %v, want %v", stats.TotalPlaylists, 1)
	}
	if stats.TotalChannels != 1 {
		t.Errorf("TotalChannels = %v, want %v", stats.TotalChannels, 1)
	}

	// Verify LastUpdate is set
	if stats.LastUpdate.IsZero() {
		t.Error("LastUpdate should not be zero")
	}

	// TopVideos should be initialized (even if empty)
	// Note: TopVideos can be empty if no video files exist in the download directory
}

func TestGetStatistics_EmptyRepo(t *testing.T) {
	mockRepo := testutil.NewMockJobRepository()
	tmpDir := t.TempDir()

	stats, err := GetStatistics(mockRepo, tmpDir)
	if err != nil {
		t.Fatalf("GetStatistics() error = %v", err)
	}

	if stats.TotalJobs != 0 {
		t.Errorf("TotalJobs = %v, want %v", stats.TotalJobs, 0)
	}
	if stats.TotalVideos != 0 {
		t.Errorf("TotalVideos = %v, want %v", stats.TotalVideos, 0)
	}
}

func TestCalculateDirSize(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	files := []struct {
		name string
		size int
	}{
		{"file1.txt", 1024},
		{"file2.txt", 2048},
		{"file3.txt", 512},
	}

	expectedTotal := 0
	for _, f := range files {
		path := filepath.Join(tmpDir, f.name)
		data := make([]byte, f.size)
		if err := os.WriteFile(path, data, 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		expectedTotal += f.size
	}

	// Calculate directory size
	size, err := calculateDirSize(tmpDir)
	if err != nil {
		t.Fatalf("calculateDirSize() error = %v", err)
	}

	if size != expectedTotal {
		t.Errorf("Size = %v, want %v", size, expectedTotal)
	}
}

func TestCalculateDirSize_Subdirectories(t *testing.T) {
	tmpDir := t.TempDir()

	// Create subdirectories with files
	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	// Create files in both directories
	file1 := filepath.Join(tmpDir, "file1.txt")
	file2 := filepath.Join(subDir, "file2.txt")

	data1 := make([]byte, 1024)
	data2 := make([]byte, 2048)

	os.WriteFile(file1, data1, 0644)
	os.WriteFile(file2, data2, 0644)

	size, err := calculateDirSize(tmpDir)
	if err != nil {
		t.Fatalf("calculateDirSize() error = %v", err)
	}

	expectedSize := 1024 + 2048
	if size != expectedSize {
		t.Errorf("Size = %v, want %v", size, expectedSize)
	}
}

func TestCalculateDirSize_NonExistent(t *testing.T) {
	_, err := calculateDirSize("/nonexistent/path")
	if err == nil {
		t.Error("Expected error for non-existent directory, got nil")
	}
}

func TestFindTopVideosBySize(t *testing.T) {
	tmpDir := t.TempDir()

	// Create channel directory
	channelDir := filepath.Join(tmpDir, "TestChannel")
	if err := os.MkdirAll(channelDir, 0755); err != nil {
		t.Fatalf("Failed to create channel directory: %v", err)
	}

	// Create test video files with different sizes
	videos := []struct {
		name string
		size int
	}{
		{"Video1.mp4", 10485760}, // 10 MB
		{"Video2.mp4", 5242880},  // 5 MB
		{"Video3.mp4", 2097152},  // 2 MB
		{"Video4.mp4", 1048576},  // 1 MB
	}

	for _, v := range videos {
		path := filepath.Join(channelDir, v.name)
		data := make([]byte, v.size)
		if err := os.WriteFile(path, data, 0644); err != nil {
			t.Fatalf("Failed to create test video: %v", err)
		}
	}

	// Also create some non-video files that should be excluded
	os.WriteFile(filepath.Join(channelDir, "metadata.info.json"), []byte("{}"), 0644)
	os.WriteFile(filepath.Join(channelDir, "temp.part"), []byte("temp"), 0644)

	// Find top 3 videos
	topVideos, otherStorage, totalStorage, err := findTopVideosBySize(tmpDir, 3)
	if err != nil {
		t.Fatalf("findTopVideosBySize() error = %v", err)
	}

	// Should return top 3 videos
	if len(topVideos) != 3 {
		t.Errorf("Number of top videos = %v, want %v", len(topVideos), 3)
	}

	// First video should be the largest
	if topVideos[0].Size != 10485760 {
		t.Errorf("Largest video size = %v, want %v", topVideos[0].Size, 10485760)
	}

	// Verify sorting (largest to smallest)
	for i := 1; i < len(topVideos); i++ {
		if topVideos[i].Size > topVideos[i-1].Size {
			t.Error("Videos are not sorted by size (largest first)")
		}
	}

	// Verify channel name is extracted
	if topVideos[0].Channel != "TestChannel" {
		t.Errorf("Channel = %v, want %v", topVideos[0].Channel, "TestChannel")
	}

	// Verify title is extracted (without extension)
	if topVideos[0].Title != "Video1" {
		t.Errorf("Title = %v, want %v", topVideos[0].Title, "Video1")
	}

	// Verify total storage calculation
	expectedTotal := 10485760 + 5242880 + 2097152 + 1048576 + len("{}") + len("temp")
	if totalStorage != expectedTotal {
		t.Errorf("Total storage = %v, want %v", totalStorage, expectedTotal)
	}

	// Verify other storage calculation (total - top 3)
	topSize := 10485760 + 5242880 + 2097152
	expectedOther := expectedTotal - topSize
	if otherStorage != expectedOther {
		t.Errorf("Other storage = %v, want %v", otherStorage, expectedOther)
	}
}

func TestFindTopVideosBySize_FewerThanLimit(t *testing.T) {
	tmpDir := t.TempDir()

	channelDir := filepath.Join(tmpDir, "TestChannel")
	os.MkdirAll(channelDir, 0755)

	// Create only 2 videos
	videos := []struct {
		name string
		size int
	}{
		{"Video1.mp4", 1048576},
		{"Video2.mp4", 524288},
	}

	for _, v := range videos {
		path := filepath.Join(channelDir, v.name)
		data := make([]byte, v.size)
		os.WriteFile(path, data, 0644)
	}

	// Request top 10, but only 2 exist
	topVideos, _, _, err := findTopVideosBySize(tmpDir, 10)
	if err != nil {
		t.Fatalf("findTopVideosBySize() error = %v", err)
	}

	// Should return all 2 videos
	if len(topVideos) != 2 {
		t.Errorf("Number of videos = %v, want %v", len(topVideos), 2)
	}
}

func TestFindTopVideosBySize_EmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	topVideos, otherStorage, totalStorage, err := findTopVideosBySize(tmpDir, 10)
	if err != nil {
		t.Fatalf("findTopVideosBySize() error = %v", err)
	}

	if len(topVideos) != 0 {
		t.Errorf("Number of videos = %v, want %v", len(topVideos), 0)
	}
	if otherStorage != 0 {
		t.Errorf("Other storage = %v, want %v", otherStorage, 0)
	}
	if totalStorage != 0 {
		t.Errorf("Total storage = %v, want %v", totalStorage, 0)
	}
}

func TestFindTopVideosBySize_VideoExtensions(t *testing.T) {
	tmpDir := t.TempDir()
	channelDir := filepath.Join(tmpDir, "TestChannel")
	os.MkdirAll(channelDir, 0755)

	// Create videos with different extensions (all should be detected)
	extensions := []string{".mp4", ".webm", ".mkv", ".avi", ".mov"}

	for i, ext := range extensions {
		filename := "Video" + string(rune('1'+i)) + ext
		path := filepath.Join(channelDir, filename)
		data := make([]byte, 1024)
		os.WriteFile(path, data, 0644)
	}

	topVideos, _, _, err := findTopVideosBySize(tmpDir, 10)
	if err != nil {
		t.Fatalf("findTopVideosBySize() error = %v", err)
	}

	if len(topVideos) != len(extensions) {
		t.Errorf("Number of videos = %v, want %v (all extensions should be detected)", len(topVideos), len(extensions))
	}
}

func TestFindTopVideosBySize_ExcludesNonVideos(t *testing.T) {
	tmpDir := t.TempDir()
	channelDir := filepath.Join(tmpDir, "TestChannel")
	os.MkdirAll(channelDir, 0755)

	// Create video files
	os.WriteFile(filepath.Join(channelDir, "video.mp4"), make([]byte, 1024), 0644)

	// Create files that should be excluded
	os.WriteFile(filepath.Join(channelDir, "video.txt"), make([]byte, 1024), 0644)
	os.WriteFile(filepath.Join(channelDir, "video.part"), make([]byte, 1024), 0644)
	os.WriteFile(filepath.Join(channelDir, "video.info.json"), make([]byte, 1024), 0644)
	os.WriteFile(filepath.Join(channelDir, "video.ytdl"), make([]byte, 1024), 0644)

	topVideos, _, _, err := findTopVideosBySize(tmpDir, 10)
	if err != nil {
		t.Fatalf("findTopVideosBySize() error = %v", err)
	}

	// Should only find 1 video (the .mp4 file)
	if len(topVideos) != 1 {
		t.Errorf("Number of videos = %v, want %v (non-video files should be excluded)", len(topVideos), 1)
	}
}
