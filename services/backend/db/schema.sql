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
                                        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                                        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

INSERT OR IGNORE INTO settings (id, theme, download_quality, concurrent_downloads)
VALUES (1, 'system', 1080, 2);