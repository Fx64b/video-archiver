// Code generated by tygo. DO NOT EDIT.

//////////
// source: job.go

export interface DownloadJob {
  ID: string;
  URL: string;
  TIMESTAMP: string /* RFC3339 */;
}
export interface JobData {
  ID: string;
  JobID: string;
  URL: string;
  IsPlaylist: boolean;
  STATUS: string;
  PROGRESS: number /* int */;
  CreatedAt: string /* RFC3339 */;
  UpdatedAt: string /* RFC3339 */;
}

//////////
// source: metadata.go

export interface VideoMetadata {
  id: string;
  title: string;
  uploader: string;
  filesize: number /* int64 */;
  duration: number /* float64 */;
  format: string;
  thumbnail: string;
}
export interface PlaylistMetadata {
  id: string;
  title: string;
  description: string;
}

//////////
// source: video.go

/**
 * Video represents metadata for a single video.
 */
export interface Video {
  id: number /* int */; // Primary key in database
  job_id: string; // ID associated with the download job
  title: string; // Title of the video
  uploader: string; // Uploader or channel name
  file_path: string; // Path where the video file is stored
  last_downloaded_at: string /* RFC3339 */; // Timestamp of when the video was last downloaded
  length: number /* float64 */; // Duration of the video in seconds
  size: number /* int64 */; // File size in bytes
  quality: string; // Video quality
}
/**
 * Playlist represents metadata for a playlist or channel.
 */
export interface Playlist {
  id: string; // Playlist or channel ID
  title: string; // Title of the playlist or channel
  description: string; // Description of the playlist or channel
}
