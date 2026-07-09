package sqlite

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"path/filepath"
)

func NewDB(dbPath string) (*sql.DB, error) {
	dbDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return nil, fmt.Errorf("create database directory: %w", err)
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	// Check if schema already exists by looking for a core table
	exists, err := tableExists(db, "jobs")
	if err != nil {
		return nil, fmt.Errorf("check schema: %w", err)
	}

	if !exists {
		if err := initSchema(db); err != nil {
			return nil, fmt.Errorf("init schema: %w", err)
		}
	} else {
		// Run migrations for existing databases
		if err := runMigrations(db); err != nil {
			return nil, fmt.Errorf("run migrations: %w", err)
		}
	}

	return db, nil
}

func tableExists(db *sql.DB, tableName string) (bool, error) {
	var name string
	err := db.QueryRow(`
        SELECT name FROM sqlite_master 
        WHERE type='table' AND name=?`, tableName).Scan(&name)

	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("query table existence: %w", err)
	}

	return true, nil
}

func initSchema(db *sql.DB) error {
	schema, err := os.ReadFile("./db/schema.sql")
	if err != nil {
		return fmt.Errorf("read schema file: %w", err)
	}

	_, err = db.Exec(string(schema))
	if err != nil {
		return fmt.Errorf("execute schema: %w", err)
	}

	return nil
}

func columnExists(db *sql.DB, tableName, columnName string) (bool, error) {
	rows, err := db.Query(fmt.Sprintf("PRAGMA table_info(%s)", tableName))
	if err != nil {
		return false, fmt.Errorf("query table info: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name string
		var dataType string
		var notNull int
		var defaultValue sql.NullString
		var pk int

		if err := rows.Scan(&cid, &name, &dataType, &notNull, &defaultValue, &pk); err != nil {
			return false, fmt.Errorf("scan column info: %w", err)
		}

		if name == columnName {
			return true, nil
		}
	}

	return false, nil
}

func runMigrations(db *sql.DB) error {
	// Migration 1: Add warnings column to jobs table
	hasWarnings, err := columnExists(db, "jobs", "warnings")
	if err != nil {
		return fmt.Errorf("check warnings column: %w", err)
	}

	if !hasWarnings {
		_, err := db.Exec("ALTER TABLE jobs ADD COLUMN warnings TEXT")
		if err != nil {
			return fmt.Errorf("add warnings column: %w", err)
		}
	}

	// Migration 2: Tagging tables for databases created before tags existed
	if _, err := db.Exec(`
        CREATE TABLE IF NOT EXISTS tags (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            name TEXT NOT NULL UNIQUE COLLATE NOCASE,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        );
        CREATE TABLE IF NOT EXISTS job_tags (
            job_id TEXT NOT NULL,
            tag_id INTEGER NOT NULL,
            source TEXT NOT NULL DEFAULT 'user',
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            PRIMARY KEY (job_id, tag_id),
            FOREIGN KEY (job_id) REFERENCES jobs (job_id),
            FOREIGN KEY (tag_id) REFERENCES tags (id)
        );
        CREATE INDEX IF NOT EXISTS idx_job_tags_tag_id ON job_tags(tag_id);
    `); err != nil {
		return fmt.Errorf("create tag tables: %w", err)
	}

	// Migration 3: output media metadata on tools_jobs. Guarded on the table
	// existing — a database created before the tools feature has no tools_jobs
	// and ALTER TABLE on a missing table would fail startup.
	hasToolsJobs, err := tableExists(db, "tools_jobs")
	if err != nil {
		return fmt.Errorf("check tools_jobs table: %w", err)
	}
	if hasToolsJobs {
		toolsColumns := []struct{ name, ddl string }{
			{"media_kind", "ALTER TABLE tools_jobs ADD COLUMN media_kind TEXT NOT NULL DEFAULT ''"},
			{"duration", "ALTER TABLE tools_jobs ADD COLUMN duration REAL NOT NULL DEFAULT 0"},
			{"width", "ALTER TABLE tools_jobs ADD COLUMN width INTEGER NOT NULL DEFAULT 0"},
			{"height", "ALTER TABLE tools_jobs ADD COLUMN height INTEGER NOT NULL DEFAULT 0"},
			{"video_codec", "ALTER TABLE tools_jobs ADD COLUMN video_codec TEXT NOT NULL DEFAULT ''"},
			{"audio_codec", "ALTER TABLE tools_jobs ADD COLUMN audio_codec TEXT NOT NULL DEFAULT ''"},
		}
		for _, column := range toolsColumns {
			has, err := columnExists(db, "tools_jobs", column.name)
			if err != nil {
				return fmt.Errorf("check tools_jobs %s column: %w", column.name, err)
			}
			if !has {
				if _, err := db.Exec(column.ddl); err != nil {
					return fmt.Errorf("add tools_jobs %s column: %w", column.name, err)
				}
			}
		}
	}

	return nil
}
