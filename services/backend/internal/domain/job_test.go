package domain

import (
	"testing"
)

func TestVideoMetadata_GetType(t *testing.T) {
	video := &VideoMetadata{
		ID:    "test-id",
		Title: "Test Video",
	}

	if got := video.GetType(); got != "video" {
		t.Errorf("VideoMetadata.GetType() = %v, want %v", got, "video")
	}
}

func TestPlaylistMetadata_GetType(t *testing.T) {
	playlist := &PlaylistMetadata{
		ID:    "test-playlist-id",
		Title: "Test Playlist",
	}

	if got := playlist.GetType(); got != "playlist" {
		t.Errorf("PlaylistMetadata.GetType() = %v, want %v", got, "playlist")
	}
}

func TestChannelMetadata_GetType(t *testing.T) {
	channel := &ChannelMetadata{
		ID:      "test-channel-id",
		Channel: "Test Channel",
	}

	if got := channel.GetType(); got != "channel" {
		t.Errorf("ChannelMetadata.GetType() = %v, want %v", got, "channel")
	}
}

func TestJobStatus_Constants(t *testing.T) {
	tests := []struct {
		name     string
		status   JobStatus
		expected string
	}{
		{"pending", JobStatusPending, "pending"},
		{"in_progress", JobStatusInProgress, "in_progress"},
		{"complete", JobStatusComplete, "complete"},
		{"error", JobStatusError, "error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.status) != tt.expected {
				t.Errorf("JobStatus constant = %v, want %v", tt.status, tt.expected)
			}
		})
	}
}

func TestJobType_Constants(t *testing.T) {
	tests := []struct {
		name     string
		jobType  JobType
		expected string
	}{
		{"video", JobTypeVideo, "video"},
		{"audio", JobTypeAudio, "audio"},
		{"metadata", JobTypeMetadata, "metadata"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.jobType) != tt.expected {
				t.Errorf("JobType constant = %v, want %v", tt.jobType, tt.expected)
			}
		})
	}
}

func TestDownloadPhase_Constants(t *testing.T) {
	tests := []struct {
		name     string
		phase    string
		expected string
	}{
		{"metadata", DownloadPhaseMetadata, "metadata"},
		{"video", DownloadPhaseVideo, "video"},
		{"audio", DownloadPhaseAudio, "audio"},
		{"merging", DownloadPhaseMerging, "merging"},
		{"complete", DownloadPhaseComplete, "complete"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.phase != tt.expected {
				t.Errorf("DownloadPhase constant = %v, want %v", tt.phase, tt.expected)
			}
		})
	}
}

func TestMetadataInterface(t *testing.T) {
	// Ensure all metadata types implement the Metadata interface
	var _ Metadata = (*VideoMetadata)(nil)
	var _ Metadata = (*PlaylistMetadata)(nil)
	var _ Metadata = (*ChannelMetadata)(nil)
}

func TestProgressUpdate_Structure(t *testing.T) {
	// Test that ProgressUpdate can be created and fields are accessible
	update := ProgressUpdate{
		JobID:                "test-job",
		JobType:              "video",
		CurrentItem:          1,
		TotalItems:           1,
		Progress:             50.5,
		CurrentVideoProgress: 75.0,
		DownloadPhase:        DownloadPhaseVideo,
	}

	if update.JobID != "test-job" {
		t.Errorf("ProgressUpdate.JobID = %v, want %v", update.JobID, "test-job")
	}
	if update.Progress != 50.5 {
		t.Errorf("ProgressUpdate.Progress = %v, want %v", update.Progress, 50.5)
	}
	if update.DownloadPhase != DownloadPhaseVideo {
		t.Errorf("ProgressUpdate.DownloadPhase = %v, want %v", update.DownloadPhase, DownloadPhaseVideo)
	}
}
