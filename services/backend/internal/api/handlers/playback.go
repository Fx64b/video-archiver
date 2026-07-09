package handlers

import (
	"net/http"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
	"video-archiver/internal/domain"
)

// browserSafeCodecs reports whether a probed media file can be decoded by the
// HTML5 <video> element across browsers. Downloads are merged into mp4
// containers, but merging only remuxes: files can carry VP9/AV1 video or Opus
// audio that Safari (and others) cannot decode. h264 + aac/mp3 is the
// universally supported combination.
func browserSafeCodecs(videoCodec, audioCodec string, hasAudio bool) bool {
	if videoCodec != "h264" {
		return false
	}
	if !hasAudio {
		return true
	}
	return audioCodec == "aac" || audioCodec == "mp3"
}

// HandlePlaybackInfo reports a video's container/codecs, whether the browser
// can play it directly, and the state of any transcode job for it.
func (h *Handler) HandlePlaybackInfo(w http.ResponseWriter, r *http.Request) {
	job, metadata, ok := h.videoJobFromRequest(w, r)
	if !ok {
		return
	}

	videoPath, err := h.locateVideoFile(job, metadata)
	if err != nil {
		http.Error(w, "Video file not found", http.StatusNotFound)
		return
	}

	probe, err := h.ffmpeg.Probe(videoPath)
	if err != nil {
		log.WithError(err).WithField("path", videoPath).Error("Failed to probe video file")
		http.Error(w, "Failed to inspect video file", http.StatusInternalServerError)
		return
	}

	info := domain.PlaybackInfo{
		Container:   strings.TrimPrefix(strings.ToLower(filepath.Ext(videoPath)), "."),
		VideoCodec:  probe.VideoCodec,
		AudioCodec:  probe.AudioCodec,
		BrowserSafe: browserSafeCodecs(probe.VideoCodec, probe.AudioCodec, probe.HasAudio),
	}

	if transcode, err := h.toolsRepository.FindLatestConvertForInput(job.ID); err != nil {
		log.WithError(err).Warn("Failed to look up transcode job")
	} else if transcode != nil {
		info.Transcode = &domain.PlaybackTranscode{
			JobID:    transcode.ID,
			Status:   transcode.Status,
			Progress: transcode.Progress,
		}
	}

	writeJSON(w, http.StatusOK, Response{Message: info})
}

// HandleRequestTranscode submits a convert job that produces a browser-safe
// (h264/aac mp4) version of a video. If a transcode is already pending or
// running, that job is returned instead of starting a duplicate.
func (h *Handler) HandleRequestTranscode(w http.ResponseWriter, r *http.Request) {
	job, metadata, ok := h.videoJobFromRequest(w, r)
	if !ok {
		return
	}

	existing, err := h.toolsRepository.FindLatestConvertForInput(job.ID)
	if err != nil {
		log.WithError(err).Error("Failed to look up existing transcode job")
		http.Error(w, "Failed to check existing transcode", http.StatusInternalServerError)
		return
	}
	if existing != nil &&
		(existing.Status == domain.ToolsJobStatusPending || existing.Status == domain.ToolsJobStatusProcessing) {
		writeJSON(w, http.StatusOK, Response{Message: domain.PlaybackTranscode{
			JobID:    existing.ID,
			Status:   existing.Status,
			Progress: existing.Progress,
		}})
		return
	}

	toolsJob := &domain.ToolsJob{
		OperationType: domain.OpTypeConvert,
		InputFiles:    []string{job.ID},
		InputType:     domain.InputTypeVideos,
		Parameters: map[string]any{
			"output_format": "mp4",
			"video_codec":   "libx264",
			"audio_codec":   "aac",
			"output_name":   metadata.Title + " (web)",
		},
	}
	if err := h.toolsService.Submit(toolsJob); err != nil {
		log.WithError(err).Error("Failed to submit transcode job")
		http.Error(w, "Failed to start transcode", http.StatusServiceUnavailable)
		return
	}

	log.WithField("jobID", job.ID).WithField("toolsJobID", toolsJob.ID).Info("Transcode requested")
	writeJSON(w, http.StatusAccepted, Response{Message: domain.PlaybackTranscode{
		JobID:    toolsJob.ID,
		Status:   toolsJob.Status,
		Progress: toolsJob.Progress,
	}})
}

// videoJobFromRequest loads the job named by the jobID URL parameter and
// requires it to be a video. Responds with the appropriate error and returns
// ok=false when it isn't.
func (h *Handler) videoJobFromRequest(w http.ResponseWriter, r *http.Request) (*domain.Job, *domain.VideoMetadata, bool) {
	jobID := chi.URLParam(r, "jobID")
	if jobID == "" {
		http.Error(w, "Missing job ID", http.StatusBadRequest)
		return nil, nil, false
	}

	jobWithMetadata, err := h.downloadService.GetJobWithMetadata(jobID)
	if err != nil || jobWithMetadata == nil || jobWithMetadata.Job == nil {
		http.Error(w, "Video not found", http.StatusNotFound)
		return nil, nil, false
	}

	metadata, ok := jobWithMetadata.Metadata.(*domain.VideoMetadata)
	if !ok {
		http.Error(w, "Unsupported content type for video playback", http.StatusBadRequest)
		return nil, nil, false
	}

	return jobWithMetadata.Job, metadata, true
}
