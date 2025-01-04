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
