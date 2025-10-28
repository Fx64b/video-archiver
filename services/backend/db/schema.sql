PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS jobs (
                                    id INTEGER PRIMARY KEY AUTOINCREMENT,
                                    job_id TEXT UNIQUE,
                                    url TEXT NOT NULL,
                                    is_playlist BOOLEAN DEFAULT FALSE,
                                    status TEXT,
                                    progress REAL,
                                    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                                    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS videos (
                                      id INTEGER PRIMARY KEY AUTOINCREMENT,
                                      job_id TEXT,
                                      title TEXT,
                                      metadata_json TEXT,
                                      FOREIGN KEY (job_id) REFERENCES jobs (job_id)
);

CREATE TABLE IF NOT EXISTS playlists (
                                         id INTEGER PRIMARY KEY AUTOINCREMENT,
                                         job_id TEXT,
                                         title TEXT,
                                         metadata_json TEXT,
                                         FOREIGN KEY (job_id) REFERENCES jobs (job_id)
);

CREATE TABLE IF NOT EXISTS channels (
                                        id INTEGER PRIMARY KEY AUTOINCREMENT,
                                        job_id TEXT,
                                        name TEXT NOT NULL,
                                        metadata_json TEXT,
                                        FOREIGN KEY (job_id) REFERENCES jobs (job_id)
);

CREATE TABLE IF NOT EXISTS video_memberships (
                                                 id INTEGER PRIMARY KEY AUTOINCREMENT,
                                                 video_job_id TEXT,
                                                 parent_job_id TEXT,
                                                 membership_type TEXT,
                                                 FOREIGN KEY (video_job_id) REFERENCES jobs (job_id),
                                                 FOREIGN KEY (parent_job_id) REFERENCES jobs (job_id)
);

CREATE TABLE IF NOT EXISTS settings (
                                        id INTEGER PRIMARY KEY CHECK (id = 1),
                                        theme TEXT DEFAULT 'system',
                                        download_quality INTEGER DEFAULT 1080,
                                        concurrent_downloads INTEGER DEFAULT 2,
                                        tools_default_format TEXT DEFAULT 'mp4',
                                        tools_default_quality TEXT DEFAULT '1080p',
                                        tools_preserve_original BOOLEAN DEFAULT 1,
                                        tools_output_path TEXT DEFAULT './data/processed',
                                        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                                        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS tools_jobs (
                                          id TEXT PRIMARY KEY,
                                          operation_type TEXT NOT NULL,
                                          status TEXT NOT NULL,
                                          progress REAL NOT NULL DEFAULT 0,
                                          input_files TEXT NOT NULL,
                                          input_type TEXT NOT NULL DEFAULT 'videos',
                                          output_file TEXT,
                                          parameters TEXT,
                                          error_message TEXT,
                                          created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                                          updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                                          completed_at TIMESTAMP,
                                          estimated_size INTEGER,
                                          actual_size INTEGER
);

CREATE INDEX IF NOT EXISTS idx_tools_jobs_status ON tools_jobs(status);
CREATE INDEX IF NOT EXISTS idx_tools_jobs_created_at ON tools_jobs(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_tools_jobs_operation_type ON tools_jobs(operation_type);
-- Composite index for filtered pagination queries (e.g., "get pending jobs ordered by date")
CREATE INDEX IF NOT EXISTS idx_tools_jobs_status_created ON tools_jobs(status, created_at DESC);

INSERT OR IGNORE INTO settings (id, theme, download_quality, concurrent_downloads, tools_default_format, tools_default_quality, tools_preserve_original, tools_output_path)
VALUES (1, 'system', 1080, 2, 'mp4', '1080p', 1, './data/processed');