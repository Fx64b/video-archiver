package storage

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	log "github.com/sirupsen/logrus"
	"os"
	"time"
)

var db *sql.DB

func InitDB(dataSourceName string) error {
	var err error
	db, err = sql.Open("sqlite3", dataSourceName)
	if err != nil {
		return err
	}

	if _, err := os.Stat(dataSourceName); os.IsNotExist(err) {
		log.Println("Database file created successfully.")
	}

	err = loadSchema("./db/schema.sql")
	if err != nil {
		return err
	}

	log.Println("Database initialized successfully.")
	return nil
}

func loadSchema(filePath string) error {
	sqlBytes, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read schema file: %w", err)
	}

	_, err = db.Exec(string(sqlBytes))
	if err != nil {
		return fmt.Errorf("failed to execute schema: %w", err)
	}

	log.Println("Database schema loaded successfully.")
	return nil
}

func AddJob(jobID string, url string, isPlaylist bool) error {
	_, err := db.Exec(`INSERT INTO jobs (job_id, url, is_playlist, status, progress) VALUES (?, ?, ?, 'pending', 0)`, jobID, url, isPlaylist)
	return err
}

func UpdateJobProgress(jobID string, progress float64, status string) error {
	_, err := db.Exec(`UPDATE jobs SET progress = ?, status = ?, updated_at = ? WHERE job_id = ?`, progress, status, time.Now(), jobID)
	return err
}

func StoreVideoMetadata(jobID, title, uploader, filePath string) error {
	_, err := db.Exec(`INSERT INTO videos (job_id, title, uploader, file_path, last_downloaded_at) VALUES (?, ?, ?, ?, ?)`, jobID, title, uploader, filePath, time.Now())
	return err
}

func CloseDB() error {
	if db != nil {
		return db.Close()
	}
	return nil
}
