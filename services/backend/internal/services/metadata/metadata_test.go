package metadata

import (
	"os"
	"path/filepath"
	"testing"
	"video-archiver/internal/domain"
	"video-archiver/internal/testutil"
)

func TestExtractMetadata_Video(t *testing.T) {
	// Create temporary file with video metadata
	tmpDir := t.TempDir()
	metadataFile := filepath.Join(tmpDir, "video.info.json")

	err := os.WriteFile(metadataFile, []byte(testutil.VideoMetadataJSON), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Extract metadata
	metadata, err := ExtractMetadata(metadataFile)
	if err != nil {
		t.Fatalf("ExtractMetadata() error = %v", err)
	}

	// Verify it's a VideoMetadata
	videoMeta, ok := metadata.(*domain.VideoMetadata)
	if !ok {
		t.Fatalf("Expected VideoMetadata, got %T", metadata)
	}

	// Verify fields
	if videoMeta.ID != "dQw4w9WgXcQ" {
		t.Errorf("ID = %v, want %v", videoMeta.ID, "dQw4w9WgXcQ")
	}
	if videoMeta.Title != "Rick Astley - Never Gonna Give You Up (Official Video)" {
		t.Errorf("Title = %v, want %v", videoMeta.Title, "Rick Astley - Never Gonna Give You Up (Official Video)")
	}
	if videoMeta.Channel != "Rick Astley" {
		t.Errorf("Channel = %v, want %v", videoMeta.Channel, "Rick Astley")
	}
	if videoMeta.Duration != 212 {
		t.Errorf("Duration = %v, want %v", videoMeta.Duration, 212)
	}
	if videoMeta.GetType() != "video" {
		t.Errorf("GetType() = %v, want %v", videoMeta.GetType(), "video")
	}
}

func TestExtractMetadata_Playlist(t *testing.T) {
	// Create temporary file with playlist metadata
	tmpDir := t.TempDir()
	metadataFile := filepath.Join(tmpDir, "playlist.info.json")

	err := os.WriteFile(metadataFile, []byte(testutil.PlaylistMetadataJSON), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Extract metadata
	metadata, err := ExtractMetadata(metadataFile)
	if err != nil {
		t.Fatalf("ExtractMetadata() error = %v", err)
	}

	// Verify it's a PlaylistMetadata
	playlistMeta, ok := metadata.(*domain.PlaylistMetadata)
	if !ok {
		t.Fatalf("Expected PlaylistMetadata, got %T", metadata)
	}

	// Verify fields
	if playlistMeta.ID != "PLrAXtmErZgOeiKm4sgNOknGvNjby9efdf" {
		t.Errorf("ID = %v, want %v", playlistMeta.ID, "PLrAXtmErZgOeiKm4sgNOknGvNjby9efdf")
	}
	if playlistMeta.Title != "Best Music Videos" {
		t.Errorf("Title = %v, want %v", playlistMeta.Title, "Best Music Videos")
	}
	if playlistMeta.ItemCount != 25 {
		t.Errorf("ItemCount = %v, want %v", playlistMeta.ItemCount, 25)
	}
	if playlistMeta.GetType() != "playlist" {
		t.Errorf("GetType() = %v, want %v", playlistMeta.GetType(), "playlist")
	}
}

func TestExtractMetadata_Channel(t *testing.T) {
	// Create temporary file with channel metadata
	tmpDir := t.TempDir()
	metadataFile := filepath.Join(tmpDir, "channel.info.json")

	err := os.WriteFile(metadataFile, []byte(testutil.ChannelMetadataJSON), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Extract metadata
	metadata, err := ExtractMetadata(metadataFile)
	if err != nil {
		t.Fatalf("ExtractMetadata() error = %v", err)
	}

	// Verify it's a ChannelMetadata
	channelMeta, ok := metadata.(*domain.ChannelMetadata)
	if !ok {
		t.Fatalf("Expected ChannelMetadata, got %T", metadata)
	}

	// Verify fields
	if channelMeta.ID != "UCuAXFkgsw1L7xaCfnd5JJOw" {
		t.Errorf("ID = %v, want %v", channelMeta.ID, "UCuAXFkgsw1L7xaCfnd5JJOw")
	}
	// Channel name should have " - Videos" suffix removed
	if channelMeta.Channel != "Rick Astley" {
		t.Errorf("Channel = %v, want %v", channelMeta.Channel, "Rick Astley")
	}
	if channelMeta.ChannelFollowers != 3500000 {
		t.Errorf("ChannelFollowers = %v, want %v", channelMeta.ChannelFollowers, 3500000)
	}
	if channelMeta.GetType() != "channel" {
		t.Errorf("GetType() = %v, want %v", channelMeta.GetType(), "channel")
	}
	if channelMeta.Type != "channel" {
		t.Errorf("Type field = %v, want %v", channelMeta.Type, "channel")
	}
}

func TestExtractMetadata_InvalidFile(t *testing.T) {
	// Try to extract from non-existent file
	_, err := ExtractMetadata("/nonexistent/file.json")
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}
}

func TestExtractMetadata_InvalidJSON(t *testing.T) {
	// Create temporary file with invalid JSON
	tmpDir := t.TempDir()
	metadataFile := filepath.Join(tmpDir, "invalid.info.json")

	err := os.WriteFile(metadataFile, []byte("{ invalid json }"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Try to extract metadata
	_, err = ExtractMetadata(metadataFile)
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

func TestExtractMetadata_EmptyFile(t *testing.T) {
	// Create temporary empty file
	tmpDir := t.TempDir()
	metadataFile := filepath.Join(tmpDir, "empty.info.json")

	err := os.WriteFile(metadataFile, []byte(""), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Try to extract metadata
	_, err = ExtractMetadata(metadataFile)
	if err == nil {
		t.Error("Expected error for empty file, got nil")
	}
}

func TestExtractMetadata_ChannelVideosSuffixRemoval(t *testing.T) {
	// Test that " - Videos" suffix is correctly removed from channel names
	testJSON := `{
		"id": "UCtest",
		"title": "Test Channel Name - Videos",
		"channel": "Test Channel Name",
		"channel_id": "UCtest",
		"channel_url": "https://www.youtube.com/@TestChannel",
		"description": "Test",
		"thumbnails": [],
		"channel_follower_count": 1000,
		"playlist_count": 5,
		"_type": "playlist"
	}`

	tmpDir := t.TempDir()
	metadataFile := filepath.Join(tmpDir, "channel_suffix.info.json")

	err := os.WriteFile(metadataFile, []byte(testJSON), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	metadata, err := ExtractMetadata(metadataFile)
	if err != nil {
		t.Fatalf("ExtractMetadata() error = %v", err)
	}

	channelMeta, ok := metadata.(*domain.ChannelMetadata)
	if !ok {
		t.Fatalf("Expected ChannelMetadata, got %T", metadata)
	}

	expectedName := "Test Channel Name"
	if channelMeta.Channel != expectedName {
		t.Errorf("Channel name = %v, want %v (suffix should be removed)", channelMeta.Channel, expectedName)
	}
}
