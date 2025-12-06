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

	return nil
}
