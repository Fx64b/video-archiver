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