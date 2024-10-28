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
                                      uploader TEXT,
                                      file_path TEXT,
                                      last_downloaded_at TIMESTAMP,
                                      length REAL,
                                      size INTEGER,
                                      quality TEXT,
                                      created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                                      FOREIGN KEY (job_id) REFERENCES jobs (job_id)
);

CREATE TABLE IF NOT EXISTS playlists (
                                         id TEXT PRIMARY KEY,
                                         title TEXT,
                                         description TEXT
);

CREATE TABLE IF NOT EXISTS tags (
                                    id INTEGER PRIMARY KEY AUTOINCREMENT,
                                    name TEXT UNIQUE
);

CREATE TABLE IF NOT EXISTS video_tags (
                                          video_id INTEGER,
                                          tag_id INTEGER,
                                          FOREIGN KEY (video_id) REFERENCES videos (id),
                                          FOREIGN KEY (tag_id) REFERENCES tags (id),
                                          PRIMARY KEY (video_id, tag_id)
);
