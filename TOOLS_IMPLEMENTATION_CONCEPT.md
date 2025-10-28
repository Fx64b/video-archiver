> **⚠️ IMPORTANT: DELETE THIS FILE AFTER PR IS MERGED**
>
> This is a temporary design document for PR implementation. Once the implementation is complete and the PR is merged, this file should be deleted as the design details will be reflected in the actual code and documentation.

---

# Tools Endpoint Implementation Concept

## Executive Summary

This document outlines a comprehensive implementation plan for the `/tools` endpoint in the Video Archiver application. The tools endpoint will provide video processing capabilities including trimming, concatenation, format conversion, audio extraction, and quality adjustment.

### Key Use Case: Automated Workflows

A critical requirement is the ability to **chain operations together** into automated workflows. The primary use case is:

**Concat → Save → Audio Extract Workflow:**
1. Concatenate multiple video files into a single video
2. Save the concatenated video file
3. Automatically extract audio from the concatenated video
4. Save the audio file (e.g., MP3 for podcasts)

This workflow is essential for content creators who want to merge multiple video segments and then extract the audio for podcast distribution, while also keeping the full video file.

### Input Selection Flexibility

A key feature is the ability to select inputs at different granularities:

- **Individual Videos**: Precise control - select exactly which videos to process
- **Entire Playlists**: Convenience - select a playlist and automatically include all videos in order
- **Entire Channels**: Bulk operations - select a channel and process all videos with custom sorting

**Example Use Cases:**
- Merge all videos from a podcast playlist into one video + audio file
- Convert all videos from a channel to audio for archival
- Concatenate all tutorial videos from a channel in chronological order

This dramatically simplifies workflows when dealing with large collections of videos.

---

## 1. Recommended Tools & Features

Based on the current architecture and the presence of ffmpeg in the backend Docker container, I recommend implementing the following tools:

### 1.1 Core Video Processing Tools

#### A. **Video Trimmer**
- Cut videos by specifying start and end timestamps
- Preview mode to verify the trim before processing
- Preserve original or re-encode based on user choice
- Use cases: Remove intros/outros, extract specific segments, create clips

#### B. **Video Concatenator**
- Merge multiple videos into a single file
- **Input selection**: Individual videos, entire playlists, or entire channels
- Support for videos with matching codecs (fast copy) or different codecs (re-encode)
- Drag-and-drop ordering of videos
- Automatic resolution/codec matching detection
- **Smart expansion**: When a playlist/channel is selected, automatically expand to all videos
- Use cases: Combine playlist videos, create compilations, merge split videos, merge all videos from a channel

#### C. **Format Converter**
- Convert between MP4, WEBM, MKV, AVI, MOV
- Preserve quality or apply custom settings
- Batch conversion support
- Use cases: Compatibility for different devices, reduce file sizes

#### D. **Audio Extractor**
- Extract audio track from videos
- Output formats: MP3, AAC, FLAC, WAV, OGG
- Configurable bitrate (128k, 192k, 256k, 320k)
- Use cases: Create podcasts, music extraction, audio-only archival

#### E. **Quality/Resolution Adjuster**
- Change video resolution (2160p → 1080p, 1080p → 720p, etc.)
- Adjust bitrate for size optimization
- Two-pass encoding for better quality
- Use cases: Reduce file sizes, optimize for mobile devices

#### F. **Video Rotation/Flip**
- Rotate video (90°, 180°, 270°)
- Flip horizontally or vertically
- Fix orientation metadata issues
- Use cases: Correct improperly oriented videos

### 1.2 Advanced Tools (Phase 2)

#### G. **Subtitle Tools**
- Extract embedded subtitles
- Add external subtitle files
- Burn-in subtitles (hardcode into video)
- Format conversion (SRT, VTT, ASS)

#### H. **Thumbnail Generator**
- Extract frames at specific timestamps
- Generate thumbnail grids (contact sheets)
- Custom dimensions and quality

#### I. **Video Speed Adjuster**
- Speed up or slow down video (0.5x - 2.0x)
- Maintain or adjust audio pitch
- Use cases: Time-lapse, slow motion review

#### J. **Batch Operations**
- Apply operations to multiple videos simultaneously
- Queue management for long-running tasks
- Progress tracking per video and overall

#### K. **Workflow/Pipeline Operations** (High Priority)
- Chain multiple operations into a single workflow
- Common workflows:
  - **Concat → Audio Extract**: Merge videos, then extract audio (e.g., podcast episodes)
  - **Trim → Convert**: Cut video, then convert to different format
  - **Adjust Quality → Extract Audio**: Reduce video quality, then extract audio
- Save intermediate files (video) and final output (audio)
- Single progress tracker for entire workflow
- Rollback support if any step fails
- Template workflows that users can save and reuse
- Use cases: Automated content processing, consistent output generation

---

## 2. Technical Architecture

### 2.1 Backend Architecture

#### Component Structure
```
services/backend/internal/
├── api/handlers/
│   └── tools.go                    # Tools endpoint handlers
├── services/
│   └── tools/
│       ├── service.go              # Main tools service
│       ├── job_manager.go          # Job queue & execution
│       ├── ffmpeg_wrapper.go       # FFmpeg command builder
│       ├── operations/
│       │   ├── trim.go             # Video trimming logic
│       │   ├── concat.go           # Video concatenation
│       │   ├── convert.go          # Format conversion
│       │   ├── extract_audio.go    # Audio extraction
│       │   ├── adjust_quality.go   # Quality adjustment
│       │   └── rotate.go           # Rotation/flip
│       └── validator.go            # Input validation
├── domain/
│   └── tools.go                    # Tools domain entities
└── repositories/
    └── sqlite/
        └── tools_repository.go     # Tools job persistence
```

#### Domain Entities (`domain/tools.go`)

```go
type ToolsJobStatus string

const (
    ToolsJobStatusPending    ToolsJobStatus = "pending"
    ToolsJobStatusProcessing ToolsJobStatus = "processing"
    ToolsJobStatusComplete   ToolsJobStatus = "complete"
    ToolsJobStatusFailed     ToolsJobStatus = "failed"
    ToolsJobStatusCancelled  ToolsJobStatus = "cancelled"
)

type ToolsOperationType string

const (
    OpTypeTrim         ToolsOperationType = "trim"
    OpTypeConcat       ToolsOperationType = "concat"
    OpTypeConvert      ToolsOperationType = "convert"
    OpTypeExtractAudio ToolsOperationType = "extract_audio"
    OpTypeAdjustQuality ToolsOperationType = "adjust_quality"
    OpTypeRotate       ToolsOperationType = "rotate"
    OpTypeWorkflow     ToolsOperationType = "workflow"
)

type ToolsJob struct {
    ID            string             `json:"id"`
    OperationType ToolsOperationType `json:"operation_type"`
    Status        ToolsJobStatus     `json:"status"`
    Progress      float64            `json:"progress"`        // 0-100
    InputFiles    []string           `json:"input_files"`     // Job IDs (videos, playlists, or channels)
    InputType     string             `json:"input_type"`      // "videos", "playlist", "channel"
    OutputFile    string             `json:"output_file"`     // Generated file path
    Parameters    map[string]any     `json:"parameters"`      // Operation-specific params
    ErrorMessage  string             `json:"error_message,omitempty"`
    CreatedAt     time.Time          `json:"created_at"`
    UpdatedAt     time.Time          `json:"updated_at"`
    CompletedAt   *time.Time         `json:"completed_at,omitempty"`
    EstimatedSize int64              `json:"estimated_size,omitempty"` // Bytes
    ActualSize    int64              `json:"actual_size,omitempty"`     // Bytes
}

// Operation-specific parameter structures
type TrimParameters struct {
    StartTime string `json:"start_time"` // HH:MM:SS or seconds
    EndTime   string `json:"end_time"`   // HH:MM:SS or seconds
    ReEncode  bool   `json:"re_encode"`  // Force re-encode or stream copy
}

type ConcatParameters struct {
    OutputFormat string   `json:"output_format"` // mp4, mkv, etc.
    ReEncode     bool     `json:"re_encode"`     // Force re-encode if codecs differ
    FileOrder    []string `json:"file_order"`    // Explicit ordering of input files
}

type ConvertParameters struct {
    OutputFormat  string `json:"output_format"`  // mp4, webm, mkv, avi, mov
    VideoCodec    string `json:"video_codec"`    // h264, h265, vp9, etc.
    AudioCodec    string `json:"audio_codec"`    // aac, mp3, opus, etc.
    Bitrate       string `json:"bitrate"`        // e.g., "2M", "5M"
    PreserveQuality bool `json:"preserve_quality"` // Use original quality settings
}

type ExtractAudioParameters struct {
    OutputFormat string `json:"output_format"` // mp3, aac, flac, wav, ogg
    Bitrate      string `json:"bitrate"`       // e.g., "128k", "320k"
    SampleRate   int    `json:"sample_rate"`   // e.g., 44100, 48000
}

type AdjustQualityParameters struct {
    Resolution string `json:"resolution"` // 480p, 720p, 1080p, etc.
    Bitrate    string `json:"bitrate"`    // Target bitrate
    CRF        int    `json:"crf"`        // Constant Rate Factor (0-51)
    TwoPass    bool   `json:"two_pass"`   // Use two-pass encoding
}

type RotateParameters struct {
    Rotation int  `json:"rotation"` // 90, 180, 270 degrees
    FlipH    bool `json:"flip_h"`   // Flip horizontal
    FlipV    bool `json:"flip_v"`   // Flip vertical
}

type WorkflowParameters struct {
    Steps                []WorkflowStep `json:"steps"`                   // Ordered list of operations
    KeepIntermediateFiles bool          `json:"keep_intermediate_files"` // Save intermediate outputs
    StopOnError          bool          `json:"stop_on_error"`           // Stop workflow if step fails
}

type WorkflowStep struct {
    Operation  ToolsOperationType `json:"operation"`  // Operation type for this step
    Parameters map[string]any     `json:"parameters"` // Parameters for this operation
    OutputName string             `json:"output_name,omitempty"` // Custom name for output
}

// Progress update for WebSocket
type ToolsProgressUpdate struct {
    JobID       string         `json:"job_id"`
    Status      ToolsJobStatus `json:"status"`
    Progress    float64        `json:"progress"`
    CurrentStep string         `json:"current_step"` // e.g., "Analyzing", "Encoding", "Finalizing"
    TimeElapsed int            `json:"time_elapsed"` // Seconds
    TimeRemaining int          `json:"time_remaining"` // Estimated seconds
    Error       string         `json:"error,omitempty"`
}
```

### 2.2 API Endpoints

#### Base Path: `/api/tools`

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/tools/operations` | List available operations with descriptions |
| POST | `/api/tools/trim` | Create video trim job |
| POST | `/api/tools/concat` | Create video concatenation job |
| POST | `/api/tools/convert` | Create format conversion job |
| POST | `/api/tools/extract-audio` | Create audio extraction job |
| POST | `/api/tools/adjust-quality` | Create quality adjustment job |
| POST | `/api/tools/rotate` | Create rotation/flip job |
| POST | `/api/tools/workflow` | Create multi-step workflow job |
| GET | `/api/tools/jobs` | List all tools jobs (paginated) |
| GET | `/api/tools/jobs/{id}` | Get specific job details |
| DELETE | `/api/tools/jobs/{id}` | Cancel/delete a job |
| GET | `/api/tools/preview/{jobID}/thumbnail` | Get thumbnail preview |
| GET | `/api/tools/output/{jobID}` | Download processed video |
| GET | `/api/tools/validate` | Validate operation parameters before submission |

#### Example Request: Trim Video
```json
POST /api/tools/trim
{
  "input_file": "job-uuid-123",  // Job ID from downloads
  "parameters": {
    "start_time": "00:01:30",
    "end_time": "00:05:45",
    "re_encode": false
  }
}

Response:
{
  "message": {
    "id": "tools-job-uuid-456",
    "operation_type": "trim",
    "status": "pending",
    "created_at": "2025-10-28T10:30:00Z"
  }
}
```

#### Example Request: Concatenate Videos (Individual Videos)
```json
POST /api/tools/concat
{
  "input_files": ["job-uuid-1", "job-uuid-2", "job-uuid-3"],
  "input_type": "videos",
  "parameters": {
    "output_format": "mp4",
    "re_encode": false,
    "file_order": ["job-uuid-1", "job-uuid-2", "job-uuid-3"]
  }
}
```

#### Example Request: Concatenate Entire Playlist
```json
POST /api/tools/concat
{
  "input_files": ["playlist-job-uuid"],
  "input_type": "playlist",
  "parameters": {
    "output_format": "mp4",
    "re_encode": false
  }
}

// Backend automatically expands the playlist to all its videos
// and concatenates them in the playlist order
```

#### Example Request: Concatenate All Videos from a Channel
```json
POST /api/tools/concat
{
  "input_files": ["channel-job-uuid"],
  "input_type": "channel",
  "parameters": {
    "output_format": "mp4",
    "re_encode": false,
    "sort_by": "upload_date",  // or "title", "duration"
    "order": "asc"  // or "desc"
  }
}

// Backend queries all videos belonging to this channel
// and concatenates them in the specified order
```

#### Example Request: Extract Audio
```json
POST /api/tools/extract-audio
{
  "input_file": "job-uuid-123",
  "parameters": {
    "output_format": "mp3",
    "bitrate": "320k",
    "sample_rate": 48000
  }
}
```

#### Example Request: Workflow (Concat → Audio Extract)
```json
POST /api/tools/workflow
{
  "input_files": ["job-uuid-1", "job-uuid-2", "job-uuid-3"],
  "parameters": {
    "steps": [
      {
        "operation": "concat",
        "parameters": {
          "output_format": "mp4",
          "re_encode": false,
          "file_order": ["job-uuid-1", "job-uuid-2", "job-uuid-3"]
        },
        "output_name": "podcast_episode_full_video.mp4"
      },
      {
        "operation": "extract_audio",
        "parameters": {
          "output_format": "mp3",
          "bitrate": "320k",
          "sample_rate": 48000
        },
        "output_name": "podcast_episode.mp3"
      }
    ],
    "keep_intermediate_files": true,  // Keep the concatenated video
    "stop_on_error": true
  }
}

Response:
{
  "message": {
    "id": "tools-job-workflow-xyz",
    "operation_type": "workflow",
    "status": "pending",
    "created_at": "2025-10-28T10:30:00Z",
    "estimated_size": 250000000  // Combined estimate
  }
}
```

### 2.3 Database Schema

#### New Table: `tools_jobs`
```sql
CREATE TABLE IF NOT EXISTS tools_jobs (
    id TEXT PRIMARY KEY,
    operation_type TEXT NOT NULL,
    status TEXT NOT NULL,
    progress REAL NOT NULL DEFAULT 0,
    input_files TEXT NOT NULL,  -- JSON array of job IDs (videos/playlists/channels)
    input_type TEXT NOT NULL DEFAULT 'videos',  -- 'videos', 'playlist', 'channel'
    output_file TEXT,
    parameters TEXT,  -- JSON object
    error_message TEXT,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    completed_at DATETIME,
    estimated_size INTEGER,
    actual_size INTEGER
);

CREATE INDEX idx_tools_jobs_status ON tools_jobs(status);
CREATE INDEX idx_tools_jobs_created_at ON tools_jobs(created_at DESC);
CREATE INDEX idx_tools_jobs_operation_type ON tools_jobs(operation_type);
```

#### Modified Table: `settings` (add tools preferences)
```sql
ALTER TABLE settings ADD COLUMN tools_default_format TEXT DEFAULT 'mp4';
ALTER TABLE settings ADD COLUMN tools_default_quality TEXT DEFAULT '1080p';
ALTER TABLE settings ADD COLUMN tools_preserve_original BOOLEAN DEFAULT 1;
ALTER TABLE settings ADD COLUMN tools_output_path TEXT DEFAULT './data/processed';
```

### 2.4 Input Selection & Expansion

The tools system supports three types of input selection:

#### 2.4.1 Individual Videos
- User selects specific video job IDs
- Direct mapping to video files
- Full control over order and selection

#### 2.4.2 Entire Playlists
- User selects a playlist job ID
- Backend queries `video_memberships` table to get all videos in the playlist
- Videos are automatically ordered according to playlist order
- Example query:
```sql
SELECT v.job_id, v.metadata
FROM videos v
JOIN video_memberships vm ON v.job_id = vm.video_job_id
WHERE vm.parent_job_id = 'playlist-uuid'
ORDER BY vm.position ASC;
```

#### 2.4.3 Entire Channels
- User selects a channel job ID
- Backend queries all videos belonging to the channel
- Supports custom sorting: upload_date, title, duration, view_count
- Supports ascending or descending order
- Example query:
```sql
SELECT v.job_id, v.metadata
FROM videos v
WHERE JSON_EXTRACT(v.metadata, '$.channel_id') =
      (SELECT JSON_EXTRACT(metadata, '$.id') FROM channels WHERE job_id = 'channel-uuid')
ORDER BY JSON_EXTRACT(v.metadata, '$.upload_date') DESC;
```

#### 2.4.4 Mixed Selection (Advanced)
- User can combine individual videos with expanded playlists/channels
- Backend deduplicates videos if the same video appears multiple times
- Preserves explicit video order, then appends expanded selections

**Implementation in Tools Service:**
```go
func (s *Service) expandInputFiles(inputFiles []string, inputType string, params map[string]any) ([]string, error) {
    switch inputType {
    case "videos":
        // Already individual video IDs
        return inputFiles, nil

    case "playlist":
        // Expand playlist to video IDs
        playlistJobID := inputFiles[0]
        videos, err := s.jobRepo.GetVideosForParent(playlistJobID)
        if err != nil {
            return nil, err
        }
        videoIDs := make([]string, len(videos))
        for i, v := range videos {
            videoIDs[i] = v.Job.ID
        }
        return videoIDs, nil

    case "channel":
        // Expand channel to video IDs with sorting
        channelJobID := inputFiles[0]
        sortBy := params["sort_by"].(string)
        order := params["order"].(string)
        videos, err := s.jobRepo.GetVideosForChannel(channelJobID, sortBy, order)
        if err != nil {
            return nil, err
        }
        videoIDs := make([]string, len(videos))
        for i, v := range videos {
            videoIDs[i] = v.Job.ID
        }
        return videoIDs, nil

    default:
        return nil, fmt.Errorf("invalid input type: %s", inputType)
    }
}
```

### 2.5 FFmpeg Integration

#### FFmpeg Wrapper Service (`services/tools/ffmpeg_wrapper.go`)

```go
package tools

import (
    "context"
    "fmt"
    "os/exec"
    "strings"
)

type FFmpegWrapper struct {
    ffmpegPath string
    ffprobePath string
}

func NewFFmpegWrapper() *FFmpegWrapper {
    return &FFmpegWrapper{
        ffmpegPath: "/usr/bin/ffmpeg",
        ffprobePath: "/usr/bin/ffprobe",
    }
}

// Get video metadata using ffprobe
func (f *FFmpegWrapper) GetMetadata(inputPath string) (*VideoInfo, error) {
    // Execute: ffprobe -v quiet -print_format json -show_format -show_streams
}

// Trim video
func (f *FFmpegWrapper) Trim(ctx context.Context, input, output, start, end string, reEncode bool) error {
    args := []string{
        "-i", input,
        "-ss", start,
        "-to", end,
    }

    if !reEncode {
        // Stream copy (fast, no quality loss)
        args = append(args, "-c", "copy")
    } else {
        // Re-encode with high quality
        args = append(args, "-c:v", "libx264", "-crf", "18", "-c:a", "aac", "-b:a", "192k")
    }

    args = append(args, "-progress", "pipe:1", output)
    return f.executeWithProgress(ctx, args)
}

// Concatenate videos
func (f *FFmpegWrapper) Concat(ctx context.Context, inputs []string, output string, reEncode bool) error {
    // Create concat demuxer file list
    // Execute: ffmpeg -f concat -safe 0 -i list.txt -c copy output.mp4
}

// Convert format
func (f *FFmpegWrapper) Convert(ctx context.Context, input, output string, params ConvertParameters) error {
    // Execute: ffmpeg -i input -c:v codec -b:v bitrate -c:a codec output
}

// Extract audio
func (f *FFmpegWrapper) ExtractAudio(ctx context.Context, input, output string, params ExtractAudioParameters) error {
    args := []string{
        "-i", input,
        "-vn",  // No video
        "-acodec", f.getAudioCodec(params.OutputFormat),
        "-b:a", params.Bitrate,
        "-ar", fmt.Sprintf("%d", params.SampleRate),
        "-progress", "pipe:1",
        output,
    }
    return f.executeWithProgress(ctx, args)
}

// Parse ffmpeg progress output for real-time updates
func (f *FFmpegWrapper) parseProgress(line string, duration float64) float64 {
    // Parse "out_time_us=12345678" and calculate percentage
}
```

### 2.5 Job Queue & Execution

The tools service should reuse the existing job queue pattern from the download service:

- **Worker pool**: 2-4 concurrent processing workers (configurable)
- **Queue**: Channel-based job queue
- **Progress tracking**: Parse ffmpeg output and send WebSocket updates
- **Cancellation**: Context-based cancellation for long-running jobs
- **Retry logic**: Retry failed jobs (configurable, default 1 retry)

---

## 3. Frontend Architecture

### 3.1 Page Structure (`/web/app/(dashboard)/tools/`)

```
web/app/(dashboard)/tools/
├── page.tsx                    # Main tools dashboard
├── layout.tsx                  # Tools layout wrapper
├── trim/
│   └── page.tsx               # Video trimmer interface
├── concat/
│   └── page.tsx               # Video concatenator
├── convert/
│   └── page.tsx               # Format converter
├── extract-audio/
│   └── page.tsx               # Audio extractor
├── adjust-quality/
│   └── page.tsx               # Quality adjuster
├── rotate/
│   └── page.tsx               # Rotation tool
└── workflow/
    └── page.tsx               # Workflow builder interface
```

### 3.2 Component Structure

```
web/components/
├── tools/
│   ├── ToolCard.tsx                  # Tool selection card
│   ├── VideoSelector.tsx             # Select videos from downloads
│   ├── ParameterForm.tsx             # Generic parameter input form
│   ├── ProgressTracker.tsx           # Real-time progress display
│   ├── VideoPreview.tsx              # Preview video with timeline
│   ├── TimelineSelector.tsx          # Trim start/end selector
│   ├── VideoListSortable.tsx         # Drag-drop video ordering (concat)
│   ├── QualityPresetSelector.tsx     # Quick quality presets
│   ├── JobsTable.tsx                 # Tools jobs history table
│   ├── WorkflowBuilder.tsx           # Visual workflow builder
│   ├── WorkflowStepCard.tsx          # Individual workflow step
│   └── WorkflowTemplates.tsx         # Pre-built workflow templates
└── ui/
    └── (existing shadcn components)
```

### 3.3 State Management

Add tools store to Zustand (`web/store/toolsState.ts`):

```typescript
interface ToolsState {
  activeJobs: ToolsJob[]
  jobHistory: ToolsJob[]
  selectedVideos: string[]

  // Actions
  addJob: (job: ToolsJob) => void
  updateJob: (jobId: string, update: Partial<ToolsJob>) => void
  removeJob: (jobId: string) => void
  setSelectedVideos: (videoIds: string[]) => void
  clearSelectedVideos: () => void
}
```

### 3.4 WebSocket Integration

Extend existing WebSocket hook to handle tools progress updates:

```typescript
// web/hooks/useWebSocket.ts (extend existing)
useEffect(() => {
  // ... existing download progress handling

  // Add tools progress handling
  if (data.jobID && data.jobID.startsWith('tools-')) {
    updateToolsJob(data.jobID, {
      progress: data.progress,
      status: data.status,
      currentStep: data.currentStep,
      timeRemaining: data.timeRemaining,
    })
  }
}, [message])
```

### 3.5 UI/UX Design Recommendations

#### Main Tools Dashboard (`/tools`)
- **Grid layout** of tool cards (2x3 or 3x2 depending on screen size)
- Each card shows:
  - Tool icon
  - Tool name
  - Brief description
  - "Use Tool" button
- **Recent Jobs** section below showing last 5 processing jobs
- **Quick Access**: "Resume Processing" for failed jobs

#### Individual Tool Pages (e.g., `/tools/trim`)
- **Three-column layout**:
  1. **Left**: Video selector (list of downloaded videos)
  2. **Center**: Preview & controls (video player with timeline)
  3. **Right**: Parameters form & action buttons
- **Bottom**: Progress tracker (when job is running)

#### Video Selector Component
- **Three-tab interface**: Videos | Playlists | Channels
- Search/filter by title, channel, date
- Thumbnail grid or list view toggle
- Show video duration, resolution, file size
- **Individual selection**: Multi-select videos for concat tool
- **Bulk selection**: Select entire playlist/channel with one click
- **Smart badges**: Show "12 videos" badge on playlists/channels
- Preview tooltip on hover
- Expand button to preview playlist/channel contents before selection

#### Timeline Selector (Trim Tool)
- Visual timeline with playhead
- Draggable start/end markers
- Click to set position, drag to adjust
- Show timestamps
- Preview trimmed segment
- Duration display (original vs. trimmed)

#### Parameter Forms
- Use shadcn/ui components (Select, Input, Switch, Slider)
- Preset buttons for common configurations
- Advanced options in collapsible section
- Real-time validation with error messages
- Estimated output file size display

#### Progress Tracker
- Progress bar with percentage
- Current step description (e.g., "Analyzing video", "Encoding", "Finalizing")
- Time elapsed / Time remaining
- Cancel button
- Speed (e.g., "Processing at 2.5x speed")
- Minimizable to continue browsing

---

## 4. Implementation Phases

### Phase 1: Foundation (Week 1-2)
**Goal**: Establish core infrastructure

- [ ] Create domain entities (`ToolsJob`, parameters structs)
- [ ] Set up database schema and migrations
- [ ] Create FFmpeg wrapper service with metadata extraction
- [ ] Implement job queue and worker pool
- [ ] Create tools repository (SQLite)
- [ ] Add WebSocket progress updates for tools
- [ ] Create basic API handlers structure
- [ ] Update TypeScript types generation

**Deliverables**:
- Backend can accept and queue tools jobs
- WebSocket sends progress updates
- Database stores jobs

### Phase 2: Core Tools - Trim & Extract Audio (Week 3-4)
**Goal**: Implement first two tools end-to-end

**Backend**:
- [ ] Implement trim operation with ffmpeg
- [ ] Implement extract audio operation
- [ ] Add `/api/tools/trim` endpoint
- [ ] Add `/api/tools/extract-audio` endpoint
- [ ] Add `/api/tools/jobs` endpoints (list, get, delete)
- [ ] Progress tracking with ffmpeg output parsing

**Frontend**:
- [ ] Create main tools dashboard (`/tools`)
- [ ] Create trim tool page (`/tools/trim`)
- [ ] Create extract audio page (`/tools/extract-audio`)
- [ ] Implement VideoSelector component
- [ ] Implement TimelineSelector for trim
- [ ] Implement ParameterForm components
- [ ] Implement ProgressTracker component
- [ ] Add tools state management (Zustand)
- [ ] WebSocket integration for progress

**Deliverables**:
- Users can trim videos via web interface
- Users can extract audio from videos
- Real-time progress updates

### Phase 3: Concat & Convert (Week 5-6)
**Goal**: Add concatenation and format conversion

**Backend**:
- [ ] Implement concat operation
- [ ] Implement convert operation
- [ ] Add `/api/tools/concat` endpoint
- [ ] Add `/api/tools/convert` endpoint
- [ ] Handle codec compatibility detection
- [ ] Automatic file list generation for concat

**Frontend**:
- [ ] Create concat tool page with drag-drop ordering
- [ ] Create convert tool page
- [ ] Implement VideoListSortable component
- [ ] Multi-select in VideoSelector
- [ ] Format/codec selector components
- [ ] Preview concat order

**Deliverables**:
- Users can concatenate multiple videos
- Users can convert video formats
- Drag-drop video ordering

### Phase 4: Quality Adjust & Rotate (Week 7-8)
**Goal**: Complete core tools suite

**Backend**:
- [ ] Implement adjust quality operation
- [ ] Implement rotate/flip operation
- [ ] Add `/api/tools/adjust-quality` endpoint
- [ ] Add `/api/tools/rotate` endpoint
- [ ] Two-pass encoding for quality
- [ ] Rotation metadata handling

**Frontend**:
- [ ] Create quality adjustment page
- [ ] Create rotation tool page
- [ ] Implement QualityPresetSelector
- [ ] Resolution/bitrate calculators
- [ ] Visual rotation preview
- [ ] Before/after comparison

**Deliverables**:
- Users can adjust video quality/resolution
- Users can rotate/flip videos
- All 6 core tools functional

### Phase 4.5: Workflow/Pipeline System (Week 8-9) - HIGH PRIORITY
**Goal**: Enable multi-step automated workflows

**Backend**:
- [ ] Implement workflow orchestration engine
- [ ] Add `/api/tools/workflow` endpoint
- [ ] Implement step-by-step execution with intermediate file handling
- [ ] Add workflow state management (track current step)
- [ ] Implement rollback on failure (optional)
- [ ] Support for keeping/deleting intermediate files
- [ ] Progress tracking across multiple steps

**Frontend**:
- [ ] Create workflow builder page (`/tools/workflow`)
- [ ] Implement WorkflowBuilder component (visual step builder)
- [ ] Implement WorkflowStepCard component
- [ ] Create WorkflowTemplates component with pre-built workflows
- [ ] Add "Concat → Audio Extract" template
- [ ] Drag-drop step ordering
- [ ] Step configuration forms
- [ ] Combined progress tracking for workflows

**Pre-built Templates**:
- [ ] Concat → Audio Extract (for podcasts)
- [ ] Trim → Convert (cut and convert format)
- [ ] Adjust Quality → Extract Audio (optimize then extract)

**Deliverables**:
- Users can create multi-step workflows
- **Concat → Audio Extract workflow fully functional**
- Template library for common workflows
- Single job tracks entire workflow

### Phase 5: Polish & Optimization (Week 10-11)
**Goal**: Improve UX, performance, and reliability

- [ ] Add thumbnail preview generation
- [ ] Implement job cancellation
- [ ] Add validation endpoint
- [ ] Batch operations support
- [ ] Error handling and retry logic
- [ ] Output file management (auto-cleanup old files)
- [ ] Settings page integration (default preferences)
- [ ] Performance monitoring
- [ ] Comprehensive error messages
- [ ] User documentation

**Deliverables**:
- Production-ready tools endpoint
- Comprehensive documentation
- Optimized performance

### Phase 6: Advanced Features (Future)
**Goal**: Add advanced capabilities

- [ ] Subtitle tools (extract, add, burn-in)
- [ ] Thumbnail generator
- [ ] Speed adjustment
- [ ] Batch operations UI
- [ ] Presets system (save common configurations)
- [ ] History/analytics for tools usage
- [ ] Advanced ffmpeg options for power users

---

## 5. Technical Considerations

### 5.1 Performance

**File I/O**:
- Operations should work on existing downloaded files (no copying unless necessary)
- Output files stored in separate directory (`./data/processed/`)
- Use stream copy when possible (no re-encoding) for faster processing
- Implement smart codec detection to determine when re-encoding is needed

**Concurrency**:
- Limit concurrent ffmpeg processes (2-4 workers)
- Queue jobs when worker pool is full
- Allow users to set priority for jobs
- Prevent CPU/memory exhaustion

**Progress Tracking**:
- Parse ffmpeg progress output in real-time
- Calculate accurate ETAs based on current speed
- Handle progress for multi-file operations (concat)

### 5.2 Error Handling

**Validation**:
- Validate file existence before processing
- Check file format compatibility
- Validate timestamp formats (HH:MM:SS)
- Check available disk space before operations

**Error Recovery**:
- Retry failed jobs once with exponential backoff
- Clean up partial output files on failure
- Detailed error messages for users
- Log ffmpeg errors for debugging

**Edge Cases**:
- Handle corrupt video files gracefully
- Deal with unsupported codecs/formats
- Handle very large files (>10GB)
- Timeout for extremely long operations

### 5.3 Storage Management

**Output Files**:
- Store in separate directory from downloads
- Maintain reference to source files (job IDs)
- Auto-cleanup old processed files (configurable retention)
- Display storage usage in UI

**Temporary Files**:
- Use temp directory for intermediate processing
- Clean up temp files on completion/failure
- Handle disk space issues

### 5.4 Security

**Input Validation**:
- Sanitize all file paths (prevent directory traversal)
- Validate ffmpeg parameters (prevent command injection)
- Rate limit API requests
- Validate file ownership (users can only process their downloads)

**Resource Limits**:
- Max job queue size (prevent DoS)
- Max output file size limits
- Processing timeout per job
- Memory limits for ffmpeg processes

### 5.5 User Experience

**Feedback**:
- Show estimated processing time before starting
- Display file size comparison (before/after)
- Preview results before finalizing
- Cancel/pause long-running jobs

**Defaults**:
- Smart defaults based on source video
- Remember user preferences
- Quick presets for common operations
- Bulk apply settings

**History**:
- Track all processed videos
- Show success/failure rates
- Allow re-processing with same settings
- Export processing history

---

## 6. API Examples

### 6.1 List Available Operations
```bash
GET /api/tools/operations

Response:
{
  "message": [
    {
      "type": "trim",
      "name": "Video Trimmer",
      "description": "Cut videos by specifying start and end times",
      "supported_formats": ["mp4", "mkv", "webm", "avi", "mov"]
    },
    {
      "type": "concat",
      "name": "Video Concatenator",
      "description": "Merge multiple videos into one file",
      "supported_formats": ["mp4", "mkv", "webm"]
    },
    // ... other operations
  ]
}
```

### 6.2 Validate Parameters
```bash
POST /api/tools/validate
{
  "operation_type": "trim",
  "input_file": "job-uuid-123",
  "parameters": {
    "start_time": "00:01:30",
    "end_time": "00:05:45"
  }
}

Response:
{
  "valid": true,
  "estimated_duration": 255,  // seconds
  "estimated_size": 125829120,  // bytes (120 MB)
  "warnings": []
}

OR (if invalid):
{
  "valid": false,
  "errors": [
    "End time (00:05:45) exceeds video duration (00:04:30)"
  ]
}
```

### 6.3 Submit Job
```bash
POST /api/tools/trim
{
  "input_file": "job-uuid-123",
  "parameters": {
    "start_time": "00:01:30",
    "end_time": "00:04:30",
    "re_encode": false
  }
}

Response:
{
  "message": {
    "id": "tools-job-abc123",
    "operation_type": "trim",
    "status": "pending",
    "progress": 0,
    "created_at": "2025-10-28T10:30:00Z",
    "estimated_size": 85000000
  }
}
```

### 6.4 Get Job Status
```bash
GET /api/tools/jobs/tools-job-abc123

Response:
{
  "message": {
    "id": "tools-job-abc123",
    "operation_type": "trim",
    "status": "processing",
    "progress": 47.5,
    "input_files": ["job-uuid-123"],
    "output_file": "/data/processed/trimmed_video_abc123.mp4",
    "parameters": {
      "start_time": "00:01:30",
      "end_time": "00:04:30",
      "re_encode": false
    },
    "created_at": "2025-10-28T10:30:00Z",
    "updated_at": "2025-10-28T10:31:15Z",
    "estimated_size": 85000000
  }
}
```

### 6.5 List Jobs (Paginated)
```bash
GET /api/tools/jobs?page=1&limit=20&status=complete&operation_type=trim

Response:
{
  "message": {
    "items": [
      { /* ToolsJob object */ },
      { /* ToolsJob object */ }
    ],
    "total_count": 45,
    "page": 1,
    "limit": 20,
    "total_pages": 3
  }
}
```

### 6.6 Cancel Job
```bash
DELETE /api/tools/jobs/tools-job-abc123

Response:
{
  "message": "Job cancelled successfully"
}
```

### 6.7 Download Processed File
```bash
GET /api/tools/output/tools-job-abc123

Response:
- File download with appropriate Content-Type header
- Content-Disposition: attachment; filename="trimmed_video.mp4"
```

---

## 7. Testing Strategy

### 7.1 Unit Tests
- FFmpeg wrapper functions
- Parameter validation
- Progress parsing
- Job queue logic

### 7.2 Integration Tests
- End-to-end operation execution
- WebSocket progress updates
- Database persistence
- API endpoint responses

### 7.3 End-to-End Tests
- User workflow: select video → configure → process → download
- Multiple concurrent jobs
- Job cancellation
- Error scenarios

### 7.4 Performance Tests
- Large file processing (>5GB)
- Concurrent job execution
- Memory usage monitoring
- Processing speed benchmarks

### 7.5 Edge Cases
- Corrupt video files
- Unsupported formats
- Disk space issues
- Very long videos (>2 hours)
- Network interruptions during processing

---

## 8. Documentation Requirements

### 8.1 User Documentation
- Tool descriptions and use cases
- Step-by-step guides with screenshots
- Parameter explanations
- Troubleshooting common issues
- Best practices for quality/file size

### 8.2 API Documentation
- OpenAPI/Swagger specification
- Request/response examples
- Error codes and messages
- Rate limits and quotas

### 8.3 Developer Documentation
- Architecture overview
- Code structure and patterns
- Adding new tools (extension guide)
- FFmpeg command reference
- WebSocket protocol

---

## 9. Monitoring & Observability

### 9.1 Metrics
- Jobs processed (by type)
- Success/failure rates
- Average processing time (by operation)
- Storage usage (processed files)
- Active jobs count
- Queue depth

### 9.2 Logging
- Job lifecycle events (created, started, completed, failed)
- FFmpeg command execution
- Error messages with context
- Performance metrics (processing speed)

### 9.3 Alerts
- High failure rate
- Disk space low
- Processing time exceeded threshold
- Queue backlog growing

---

## 10. Future Enhancements

### 10.1 Advanced Features
- GPU acceleration for encoding (NVIDIA NVENC, Intel Quick Sync)
- AI-powered scene detection for smart trimming
- Automatic chapter generation
- Intelligent quality optimization
- Content-aware trimming (remove silence, static frames)

### 10.2 Integrations
- Direct upload to cloud storage (S3, Google Drive)
- Webhook notifications on job completion
- API for external integrations
- Plugin system for custom operations

### 10.3 UI Improvements
- Drag-drop video upload for tools
- Side-by-side before/after preview
- Waveform visualization for audio extraction
- Keyboard shortcuts for timeline navigation
- Dark mode optimizations

---

## 11. Migration Path

For existing video-archiver installations:

1. **Database Migration**: Run migration script to add `tools_jobs` table
2. **Settings Migration**: Add new tools-related settings with defaults
3. **Directory Creation**: Create `./data/processed/` directory
4. **FFmpeg Verification**: Verify ffmpeg is accessible and up-to-date
5. **Gradual Rollout**: Deploy tools endpoint without frontend first, test backend
6. **Frontend Deployment**: Deploy frontend tools pages
7. **User Communication**: Announce new feature with documentation

---

## 12. Success Metrics

### 12.1 Technical Metrics
- Processing success rate: >95%
- Average processing time: <1 minute per minute of video (1x speed)
- Concurrent job handling: 4+ simultaneous jobs
- API response time: <200ms (excluding processing)

### 12.2 User Metrics
- Tools adoption rate: % of users who use tools within 30 days
- Most popular tool (track usage)
- Average jobs per user per month
- User satisfaction (feedback/ratings)

---

## 13. Workflow Example: Concat → Audio Extract

### Visual Flow Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│  Input: Multiple Downloaded Videos                               │
│  - video-1.mp4 (job-uuid-1)                                     │
│  - video-2.mp4 (job-uuid-2)                                     │
│  - video-3.mp4 (job-uuid-3)                                     │
└────────────────────┬────────────────────────────────────────────┘
                     │
                     ▼
         ┌───────────────────────┐
         │   STEP 1: CONCAT      │
         │   Operation Type:     │
         │   "concat"            │
         │                       │
         │   ffmpeg -f concat    │
         │   -i filelist.txt     │
         │   -c copy output.mp4  │
         └───────────┬───────────┘
                     │
                     │ Progress: 0-50%
                     ▼
         ┌───────────────────────┐
         │   Intermediate File:  │
         │   podcast_ep_full.mp4 │
         │   (SAVED)             │
         │   Size: ~500MB        │
         └───────────┬───────────┘
                     │
                     ▼
         ┌───────────────────────┐
         │   STEP 2: EXTRACT     │
         │   Operation Type:     │
         │   "extract_audio"     │
         │                       │
         │   ffmpeg -i input.mp4 │
         │   -vn -acodec mp3     │
         │   -b:a 320k out.mp3   │
         └───────────┬───────────┘
                     │
                     │ Progress: 50-100%
                     ▼
         ┌───────────────────────┐
         │   Final Output:       │
         │   podcast_episode.mp3 │
         │   Size: ~80MB         │
         └───────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│  Results:                                                        │
│  ✓ Concatenated video saved: podcast_ep_full.mp4               │
│  ✓ Audio extracted: podcast_episode.mp3                        │
│  Both files available for download                              │
└─────────────────────────────────────────────────────────────────┘
```

### Implementation Details

#### Backend Flow:
1. **Receive Workflow Job**: POST /api/tools/workflow with steps array
2. **Create Workflow Job**: Store in database with status "pending"
3. **Execute Step 1 (Concat)**:
   - Create temp file list for ffmpeg concat demuxer
   - Execute: `ffmpeg -f concat -safe 0 -i list.txt -c copy output.mp4`
   - Parse progress and send WebSocket updates (0-50%)
   - Save output to: `./data/processed/podcast_ep_full.mp4`
   - Store file path for next step
4. **Execute Step 2 (Extract Audio)**:
   - Use output from Step 1 as input
   - Execute: `ffmpeg -i input.mp4 -vn -acodec libmp3lame -b:a 320k -ar 48000 output.mp3`
   - Parse progress and send WebSocket updates (50-100%)
   - Save output to: `./data/processed/podcast_episode.mp3`
5. **Complete Workflow**: Update job status to "complete"
6. **Cleanup**: Delete temp files (if keep_intermediate_files=false)

#### Frontend Flow:
1. **User Navigates**: `/tools/workflow`
2. **Selects Template**: "Concat → Audio Extract" from WorkflowTemplates
3. **Configures Step 1**: Select videos to concatenate, set output name
4. **Configures Step 2**: Set audio format (MP3), bitrate (320k), sample rate (48000)
5. **Submits Workflow**: POST request sent to backend
6. **Monitors Progress**: WebSocket updates show current step and progress
7. **Downloads Results**: Both files available via download buttons

### API Request for This Workflow

```bash
curl -X POST http://localhost:8080/api/tools/workflow \
  -H "Content-Type: application/json" \
  -d '{
    "input_files": ["job-uuid-1", "job-uuid-2", "job-uuid-3"],
    "parameters": {
      "steps": [
        {
          "operation": "concat",
          "parameters": {
            "output_format": "mp4",
            "re_encode": false,
            "file_order": ["job-uuid-1", "job-uuid-2", "job-uuid-3"]
          },
          "output_name": "podcast_ep_full.mp4"
        },
        {
          "operation": "extract_audio",
          "parameters": {
            "output_format": "mp3",
            "bitrate": "320k",
            "sample_rate": 48000
          },
          "output_name": "podcast_episode.mp3"
        }
      ],
      "keep_intermediate_files": true,
      "stop_on_error": true
    }
  }'
```

### WebSocket Progress Updates

```json
// Step 1 starting
{
  "job_id": "tools-job-workflow-xyz",
  "status": "processing",
  "progress": 5,
  "current_step": "Step 1/2: Concatenating videos",
  "time_elapsed": 10,
  "time_remaining": 180
}

// Step 1 in progress
{
  "job_id": "tools-job-workflow-xyz",
  "status": "processing",
  "progress": 25,
  "current_step": "Step 1/2: Concatenating videos",
  "time_elapsed": 50,
  "time_remaining": 120
}

// Step 1 complete, Step 2 starting
{
  "job_id": "tools-job-workflow-xyz",
  "status": "processing",
  "progress": 50,
  "current_step": "Step 2/2: Extracting audio",
  "time_elapsed": 90,
  "time_remaining": 90
}

// Step 2 in progress
{
  "job_id": "tools-job-workflow-xyz",
  "status": "processing",
  "progress": 75,
  "current_step": "Step 2/2: Extracting audio",
  "time_elapsed": 130,
  "time_remaining": 50
}

// Workflow complete
{
  "job_id": "tools-job-workflow-xyz",
  "status": "complete",
  "progress": 100,
  "current_step": "Workflow complete",
  "time_elapsed": 180,
  "time_remaining": 0
}
```

### User Experience

#### Scenario 1: Individual Videos
1. **Before**: User has 3 downloaded videos (e.g., parts of a podcast recording)
2. **Select Workflow**: Click "Workflows" → Select "Concat → Audio Extract" template
3. **Configure**:
   - Switch to "Videos" tab in selector
   - Multi-select 3 videos
   - Drag videos into desired order
   - Name the concatenated video: "podcast_ep_full.mp4"
   - Set audio format: MP3, 320kbps
   - Name the audio file: "podcast_episode.mp3"
   - Check "Keep intermediate files" to save the video
4. **Execute**: Click "Start Workflow"
5. **Monitor**: Progress bar shows current step (concat or audio extract)
6. **Complete**:
   - Download concatenated video (500MB)
   - Download audio file (80MB)
   - Both files appear in processed files list

#### Scenario 2: Entire Playlist
1. **Before**: User has downloaded a playlist with 15 videos
2. **Select Workflow**: Click "Workflows" → Select "Concat → Audio Extract" template
3. **Configure**:
   - Switch to "Playlists" tab in selector
   - Click on playlist (shows "15 videos" badge)
   - Optionally expand to preview all videos
   - Select the playlist (all videos auto-selected in playlist order)
   - Name outputs: "playlist_full.mp4" and "playlist_audio.mp3"
   - Set audio format: MP3, 320kbps
   - Check "Keep intermediate files" to save the video
4. **Execute**: Click "Start Workflow"
5. **Monitor**: Progress shows "Processing 15 videos..." and current step
6. **Complete**: Both files ready for download

#### Scenario 3: Entire Channel
1. **Before**: User has downloaded all videos from a channel (e.g., 50 videos)
2. **Select Workflow**: Click "Workflows" → Select "Concat → Audio Extract" template
3. **Configure**:
   - Switch to "Channels" tab in selector
   - Select channel (shows "50 videos" badge)
   - Choose sorting: "Upload Date" ascending (chronological order)
   - Name outputs: "channel_compilation.mp4" and "channel_audio.mp3"
   - Set audio format: MP3, 320kbps
   - Check "Keep intermediate files"
4. **Execute**: Click "Start Workflow"
5. **Monitor**: Progress shows "Processing 50 videos..." and current step
6. **Complete**: Massive compilation ready (e.g., 10GB video, 2GB audio)

### Benefits of Workflow Approach

✅ **Automated**: No need to manually chain operations
✅ **Efficient**: Single job tracks entire process
✅ **Consistent**: Template ensures same settings every time
✅ **Flexible**: Can modify steps or create custom workflows
✅ **Reliable**: Automatic rollback on failure (optional)
✅ **Transparent**: Progress tracking for each step

---

## 14. Conclusion

The `/tools` endpoint will provide essential video processing capabilities to the Video Archiver application, leveraging the existing ffmpeg installation and following the established architectural patterns. By implementing tools in phases, the project can deliver value incrementally while maintaining code quality and user experience.

The proposed architecture is:
- **Scalable**: Worker pool can be adjusted based on server resources
- **Maintainable**: Clear separation of concerns, reuses existing patterns
- **Extensible**: Easy to add new tools following the established pattern
- **User-friendly**: Modern UI with real-time feedback
- **Performant**: Smart use of stream copy when possible, concurrent processing

This implementation will position Video Archiver as a comprehensive video management platform, not just a download tool, significantly increasing its value proposition for users.
