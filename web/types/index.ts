// Code generated by tygo. DO NOT EDIT.

//////////
// source: job.go

export type JobStatus = string;
export const JobStatusPending: JobStatus = "pending";
export const JobStatusInProgress: JobStatus = "in_progress";
export const JobStatusComplete: JobStatus = "complete";
export const JobStatusError: JobStatus = "error";
export interface Job {
  id: string;
  url: string;
  status: JobStatus;
  progress: number /* float64 */;
  created_at: string /* RFC3339 */;
  updated_at: string /* RFC3339 */;
}

export type JobRepository = any;
export type JobType = string;
export const JobTypeVideo: JobType = "video";
export const JobTypeAudio: JobType = "audio";
export const JobTypeMetadata: JobType = "metadata";
export interface JobWithMetadata {
  job?: Job;
  metadata?: Metadata;
}
export interface ProgressUpdate {
  jobID: string;
  jobType: string;
  currentItem: number /* int */;
  totalItems: number /* int */;
  progress: number /* float64 */;
  currentVideoProgress: number /* float64 */;
}
export interface VideoMetadata {
  id: string;
  title: string;
  description: string;
  thumbnail: string;
  duration: number /* int */;
  view_count: number /* int */;
  channel: string;
  channel_id: string;
  channel_url: string;
  channel_follower_count: number /* int */;
  tags: string[];
  categories: string[];
  upload_date: string;
  filesize_approx: number /* int64 */;
  _type: string;
}
export interface Thumbnail {
  url: string;
  height: number /* int */;
  width: number /* int */;
  id: string;
}
export interface PlaylistMetadata {
  id: string;
  title: string;
  description: string;
  thumbnails: Thumbnail[];
  uploader_id: string;
  uploader_url: string;
  channel_id: string;
  channel: string;
  channel_url: string;
  channel_follower_count: number /* int */;
  playlist_count: number /* int */;
  _type: string;
}
export interface ChannelMetadata {
  id: string;
  channel: string;
  channel_url: string;
  description: string;
  thumbnails: Thumbnail[];
  channel_follower_count: number /* int */;
  playlist_count: number /* int */;
  _type: string;
}
export interface MetadataUpdate {
  jobID: string;
  metadata: Metadata;
}
export type Metadata = any;

//////////
// source: statistics.go

export interface Statistics {
  total_jobs: number /* int */;
  total_videos: number /* int */;
  total_playlists: number /* int */;
  total_channels: number /* int */;
  total_storage: number /* int */;
  videos_storage: number /* int */;
  playlist_storage: number /* int */;
  channel_storage: number /* int */;
  last_update: string /* RFC3339 */;
}
