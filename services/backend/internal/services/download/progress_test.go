package download

import (
	"strings"
	"testing"
	"video-archiver/internal/domain"
	"video-archiver/internal/testutil"
)

func TestNewProgressTracker(t *testing.T) {
	mockRepo := testutil.NewMockJobRepository()
	config := &Config{
		JobRepository: mockRepo,
		DownloadPath:  "/tmp/test",
		Concurrency:   2,
		MaxQuality:    1080,
	}
	service := NewService(config)

	tracker := NewProgressTracker(service, "test-job-id", "video")

	if tracker == nil {
		t.Fatal("NewProgressTracker returned nil")
	}
	if tracker.state.JobID != "test-job-id" {
		t.Errorf("JobID = %v, want %v", tracker.state.JobID, "test-job-id")
	}
	if tracker.state.JobType != "video" {
		t.Errorf("JobType = %v, want %v", tracker.state.JobType, "video")
	}
	if tracker.state.Phase != domain.DownloadPhaseMetadata {
		t.Errorf("Initial Phase = %v, want %v", tracker.state.Phase, domain.DownloadPhaseMetadata)
	}
	if tracker.state.TotalItems != 1 {
		t.Errorf("TotalItems = %v, want %v", tracker.state.TotalItems, 1)
	}
}

func TestProgressTemplatePattern(t *testing.T) {
	tests := []struct {
		name      string
		line      string
		shouldMatch bool
	}{
		{
			name:      "valid progress line",
			line:      "[1][NA][dQw4w9WgXcQ][Never Gonna Give You Up][401][1080p][avc1][none]prog:[1048576/10485760][  10.0%][1.5MiB/s][00:06]",
			shouldMatch: true,
		},
		{
			name:      "audio progress line",
			line:      "[1][NA][dQw4w9WgXcQ][Test Video][251][opus][none][opus]prog:[524288/5242880][  50.0%][800KiB/s][00:03]",
			shouldMatch: true,
		},
		{
			name:      "invalid line",
			line:      "[download] Downloading video 1 of 1",
			shouldMatch: false,
		},
		{
			name:      "metadata line",
			line:      "[youtube] Extracting URL: https://www.youtube.com/watch?v=test",
			shouldMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matched := progressTemplatePattern.MatchString(tt.line)
			if matched != tt.shouldMatch {
				t.Errorf("progressTemplatePattern.MatchString() = %v, want %v for line: %s", matched, tt.shouldMatch, tt.line)
			}
		})
	}
}

func TestDetectContentType(t *testing.T) {
	mockRepo := testutil.NewMockJobRepository()
	config := &Config{
		JobRepository: mockRepo,
		DownloadPath:  "/tmp/test",
		Concurrency:   2,
		MaxQuality:    1080,
	}
	service := NewService(config)

	tests := []struct {
		name         string
		line         string
		expectedType string
	}{
		{
			name:         "single video",
			line:         "[youtube] Extracting URL: https://www.youtube.com/watch?v=test",
			expectedType: "video",
		},
		{
			name:         "playlist",
			line:         "[download] Downloading playlist: My Awesome Playlist",
			expectedType: "playlist",
		},
		{
			name:         "channel",
			line:         "[download] Downloading playlist: TestChannel - Videos",
			expectedType: "channel",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tracker := NewProgressTracker(service, "test-job", "video")
			tracker.detectContentType(tt.line)

			if tracker.state.ContentType != tt.expectedType {
				t.Errorf("ContentType = %v, want %v", tracker.state.ContentType, tt.expectedType)
			}
		})
	}
}

func TestUpdatePlaylistProgress(t *testing.T) {
	mockRepo := testutil.NewMockJobRepository()
	config := &Config{
		JobRepository: mockRepo,
		DownloadPath:  "/tmp/test",
		Concurrency:   2,
		MaxQuality:    1080,
	}
	service := NewService(config)

	tracker := NewProgressTracker(service, "test-job", "video")

	line := "[download] Downloading item 3 of 10"
	tracker.updatePlaylistProgress(line)

	if tracker.state.CurrentItem != 3 {
		t.Errorf("CurrentItem = %v, want %v", tracker.state.CurrentItem, 3)
	}
	if tracker.state.TotalItems != 10 {
		t.Errorf("TotalItems = %v, want %v", tracker.state.TotalItems, 10)
	}
	if tracker.state.CurrentProgress != 0 {
		t.Errorf("CurrentProgress = %v, want %v after item change", tracker.state.CurrentProgress, 0.0)
	}
}

func TestDetectPhase(t *testing.T) {
	mockRepo := testutil.NewMockJobRepository()
	config := &Config{
		JobRepository: mockRepo,
		DownloadPath:  "/tmp/test",
		Concurrency:   2,
		MaxQuality:    1080,
	}
	service := NewService(config)

	tests := []struct {
		name          string
		line          string
		expectedPhase string
	}{
		{
			name:          "metadata phase",
			line:          "[youtube] dQw4w9WgXcQ: Downloading webpage",
			expectedPhase: domain.DownloadPhaseMetadata,
		},
		{
			name:          "merge phase",
			line:          "[Merger] Merging formats into \"video.mp4\"",
			expectedPhase: domain.DownloadPhaseMerging,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tracker := NewProgressTracker(service, "test-job", "video")
			tracker.detectPhase(tt.line)

			if tracker.state.Phase != tt.expectedPhase {
				t.Errorf("Phase = %v, want %v", tracker.state.Phase, tt.expectedPhase)
			}
		})
	}
}

func TestDetectPhaseWithProgressTemplate(t *testing.T) {
	mockRepo := testutil.NewMockJobRepository()
	config := &Config{
		JobRepository: mockRepo,
		DownloadPath:  "/tmp/test",
		Concurrency:   2,
		MaxQuality:    1080,
	}
	service := NewService(config)

	tests := []struct {
		name            string
		line            string
		expectedPhase   string
		expectedProgress float64
		minProgress     float64
		maxProgress     float64
	}{
		{
			name:            "video stream 50%",
			line:            "[1][NA][testID][Test Video][401][1080p][avc1][none]prog:[5242880/10485760][  50.0%][1.5MiB/s][00:05]",
			expectedPhase:   domain.DownloadPhaseVideo,
			expectedProgress: 40.0, // 50% of video = 40% overall (video is 80% of total)
			minProgress:     39.0,
			maxProgress:     41.0,
		},
		{
			name:            "audio stream 50%",
			line:            "[1][NA][testID][Test Video][251][opus][none][opus]prog:[2621440/5242880][  50.0%][800KiB/s][00:03]",
			expectedPhase:   domain.DownloadPhaseAudio,
			expectedProgress: 90.0, // 80% (video base) + 50% of 20% = 90%
			minProgress:     85.0,
			maxProgress:     95.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tracker := NewProgressTracker(service, "test-job", "video")

			// For audio test, simulate video completion first
			if strings.Contains(tt.name, "audio") {
				tracker.state.VideoProgress["testID"] = 80.0
			}

			tracker.detectPhase(tt.line)

			if tracker.state.Phase != tt.expectedPhase {
				t.Errorf("Phase = %v, want %v", tracker.state.Phase, tt.expectedPhase)
			}

			if tracker.state.CurrentProgress < tt.minProgress || tracker.state.CurrentProgress > tt.maxProgress {
				t.Errorf("CurrentProgress = %v, want between %v and %v",
					tracker.state.CurrentProgress, tt.minProgress, tt.maxProgress)
			}
		})
	}
}

func TestHandleSpecialCases(t *testing.T) {
	mockRepo := testutil.NewMockJobRepository()
	config := &Config{
		JobRepository: mockRepo,
		DownloadPath:  "/tmp/test",
		Concurrency:   2,
		MaxQuality:    1080,
	}
	service := NewService(config)

	tests := []struct {
		name             string
		line             string
		expectedProgress float64
		expectedPhase    string
	}{
		{
			name:             "already downloaded",
			line:             "[download] video.mp4 has already been downloaded",
			expectedProgress: 100.0,
			expectedPhase:    domain.DownloadPhaseComplete,
		},
		{
			name:             "recorded in archive",
			line:             "[download] video has already been recorded in archive",
			expectedProgress: 100.0,
			expectedPhase:    domain.DownloadPhaseComplete,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tracker := NewProgressTracker(service, "test-job", "video")
			tracker.handleSpecialCases(tt.line)

			if tracker.state.CurrentProgress != tt.expectedProgress {
				t.Errorf("CurrentProgress = %v, want %v", tracker.state.CurrentProgress, tt.expectedProgress)
			}
			if tracker.state.Phase != tt.expectedPhase {
				t.Errorf("Phase = %v, want %v", tracker.state.Phase, tt.expectedPhase)
			}
		})
	}
}

func TestCalculateProgress(t *testing.T) {
	mockRepo := testutil.NewMockJobRepository()
	config := &Config{
		JobRepository: mockRepo,
		DownloadPath:  "/tmp/test",
		Concurrency:   2,
		MaxQuality:    1080,
	}
	service := NewService(config)

	tests := []struct {
		name             string
		totalItems       int
		currentItem      int
		itemsCompleted   int
		currentProgress  float64
		expectedOverall  float64
	}{
		{
			name:            "single video 50%",
			totalItems:      1,
			currentItem:     1,
			itemsCompleted:  0,
			currentProgress: 50.0,
			expectedOverall: 50.0,
		},
		{
			name:            "playlist 2 of 5 at 0%",
			totalItems:      5,
			currentItem:     2,
			itemsCompleted:  1,
			currentProgress: 0.0,
			expectedOverall: 20.0, // 1 complete / 5 = 20%
		},
		{
			name:            "playlist 2 of 5 at 50%",
			totalItems:      5,
			currentItem:     2,
			itemsCompleted:  1,
			currentProgress: 50.0,
			expectedOverall: 30.0, // (100 + 50) / 5 = 30%
		},
		{
			name:            "playlist all complete",
			totalItems:      5,
			currentItem:     5,
			itemsCompleted:  5,
			currentProgress: 100.0,
			expectedOverall: 100.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tracker := NewProgressTracker(service, "test-job", "video")
			tracker.state.TotalItems = tt.totalItems
			tracker.state.CurrentItem = tt.currentItem
			tracker.state.ItemsCompleted = tt.itemsCompleted
			tracker.state.CurrentProgress = tt.currentProgress

			tracker.calculateProgress()

			if tracker.state.OverallProgress != tt.expectedOverall {
				t.Errorf("OverallProgress = %v, want %v", tracker.state.OverallProgress, tt.expectedOverall)
			}
		})
	}
}

func TestCompleteCurrentItem(t *testing.T) {
	mockRepo := testutil.NewMockJobRepository()
	config := &Config{
		JobRepository: mockRepo,
		DownloadPath:  "/tmp/test",
		Concurrency:   2,
		MaxQuality:    1080,
	}
	service := NewService(config)

	tests := []struct {
		name                 string
		initialCompleted     int
		currentItem          int
		totalItems           int
		expectedCompleted    int
		expectedPhase        string
		expectedProgress     float64
		expectedOverallDone  bool
	}{
		{
			name:                "complete first item of many",
			initialCompleted:    0,
			currentItem:         1,
			totalItems:          5,
			expectedCompleted:   1,
			expectedPhase:       domain.DownloadPhaseComplete,
			expectedProgress:    100.0,
			expectedOverallDone: false,
		},
		{
			name:                "complete last item",
			initialCompleted:    4,
			currentItem:         5,
			totalItems:          5,
			expectedCompleted:   5,
			expectedPhase:       domain.DownloadPhaseComplete,
			expectedProgress:    100.0,
			expectedOverallDone: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tracker := NewProgressTracker(service, "test-job", "video")
			tracker.state.ItemsCompleted = tt.initialCompleted
			tracker.state.CurrentItem = tt.currentItem
			tracker.state.TotalItems = tt.totalItems

			tracker.completeCurrentItem()

			if tracker.state.ItemsCompleted != tt.expectedCompleted {
				t.Errorf("ItemsCompleted = %v, want %v", tracker.state.ItemsCompleted, tt.expectedCompleted)
			}
			if tracker.state.Phase != tt.expectedPhase {
				t.Errorf("Phase = %v, want %v", tracker.state.Phase, tt.expectedPhase)
			}
			if tracker.state.CurrentProgress != tt.expectedProgress {
				t.Errorf("CurrentProgress = %v, want %v", tracker.state.CurrentProgress, tt.expectedProgress)
			}
			if tt.expectedOverallDone && tracker.state.OverallProgress != 100.0 {
				t.Errorf("OverallProgress = %v, want 100.0 for last item", tracker.state.OverallProgress)
			}
		})
	}
}

func TestStreamDownloadPattern(t *testing.T) {
	tests := []struct {
		name        string
		line        string
		shouldMatch bool
	}{
		{
			name:        "video stream destination",
			line:        "[download] Destination: /path/to/video [testID].f401.mp4",
			shouldMatch: true,
		},
		{
			name:        "audio stream destination",
			line:        "[download] Destination: /path/to/audio [testID].f251.webm",
			shouldMatch: true,
		},
		{
			name:        "non-stream line",
			line:        "[download] Downloading video 1 of 1",
			shouldMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matched := streamDownloadPattern.MatchString(tt.line)
			if matched != tt.shouldMatch {
				t.Errorf("streamDownloadPattern.MatchString() = %v, want %v for line: %s", matched, tt.shouldMatch, tt.line)
			}
		})
	}
}

func TestAlreadyDownloadedPattern(t *testing.T) {
	tests := []struct {
		name        string
		line        string
		shouldMatch bool
	}{
		{
			name:        "already downloaded",
			line:        "[download] /path/to/video.mp4 has already been downloaded",
			shouldMatch: true,
		},
		{
			name:        "recorded in archive",
			line:        "[download] video has already been recorded in archive",
			shouldMatch: true,
		},
		{
			name:        "normal download",
			line:        "[download] Downloading video",
			shouldMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matched := alreadyDownloadedPattern.MatchString(tt.line)
			if matched != tt.shouldMatch {
				t.Errorf("alreadyDownloadedPattern.MatchString() = %v, want %v for line: %s", matched, tt.shouldMatch, tt.line)
			}
		})
	}
}
