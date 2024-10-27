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
                                      created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                                      FOREIGN KEY (job_id) REFERENCES jobs (job_id)
    );

CREATE TABLE IF NOT EXISTS tags (
                                    id INTEGER PRIMARY KEY AUTOINCREMENT,
                                    name TEXT UNIQUE
);

CREATE TABLE IF NOT EXISTS video_tags (
                                          url TEXT,
                                          tag_id INTEGER,
                                          FOREIGN KEY (url) REFERENCES videos (id),
    FOREIGN KEY (tag_id) REFERENCES tags (id),
    PRIMARY KEY (url, tag_id)
    );
