package sqlite

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	// The pure-Go driver keeps builds fast: the cgo driver recompiles the
	// whole SQLite C amalgamation on any cold cache, which took minutes.
	_ "modernc.org/sqlite"
	schema "video-archiver/db"
)

// migrations is the ordered list of schema changes applied to databases
// created before the current schema. PRAGMA user_version records how many
// have run. Each migration must be safe to run on a database that already
// happens to contain its change (IF NOT EXISTS / column checks), because
// databases from before versioning report user_version 0 regardless of
// their actual shape.
var migrations = []func(*sql.DB) error{
	// 1: warnings column on jobs
	func(db *sql.DB) error {
		return addColumnIfMissing(db, "jobs", "warnings", "TEXT")
	},
	// 2: tagging tables
	func(db *sql.DB) error {
		_, err := db.Exec(`
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
    `)
		return err
	},
	// 3: downloaded media file location on jobs
	func(db *sql.DB) error {
		return addColumnIfMissing(db, "jobs", "file_path", "TEXT")
	},
}

func NewDB(dbPath string) (*sql.DB, error) {
	dbDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return nil, fmt.Errorf("create database directory: %w", err)
	}

	// WAL lets readers proceed during writes; the busy timeout makes writers
	// wait for the lock instead of failing with SQLITE_BUSY ("database is
	// locked") when download workers, progress updates and tools jobs write
	// concurrently.
	dsn := dbPath + "?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=foreign_keys(1)"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	// A single connection serializes writes at the pool level — simpler and
	// no slower for this write pattern than juggling SQLITE_BUSY handling
	// across multiple connections.
	db.SetMaxOpenConns(1)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	if err := initOrMigrate(db); err != nil {
		return nil, err
	}

	return db, nil
}

// initOrMigrate creates the schema on a fresh database, or brings an existing
// one up to date by running the not-yet-applied migrations in order.
func initOrMigrate(db *sql.DB) error {
	exists, err := tableExists(db, "jobs")
	if err != nil {
		return fmt.Errorf("check schema: %w", err)
	}

	if !exists {
		if _, err := db.Exec(schema.Schema); err != nil {
			return fmt.Errorf("execute schema: %w", err)
		}
		// A fresh schema already contains everything the migrations add.
		return setUserVersion(db, len(migrations))
	}

	version, err := userVersion(db)
	if err != nil {
		return fmt.Errorf("read schema version: %w", err)
	}

	for i := version; i < len(migrations); i++ {
		if err := migrations[i](db); err != nil {
			return fmt.Errorf("migration %d: %w", i+1, err)
		}
		if err := setUserVersion(db, i+1); err != nil {
			return fmt.Errorf("record migration %d: %w", i+1, err)
		}
	}
	return nil
}

func userVersion(db *sql.DB) (int, error) {
	var v int
	err := db.QueryRow("PRAGMA user_version").Scan(&v)
	return v, err
}

func setUserVersion(db *sql.DB, v int) error {
	// PRAGMA does not support parameter binding; v is always a trusted int.
	_, err := db.Exec(fmt.Sprintf("PRAGMA user_version = %d", v))
	return err
}

func addColumnIfMissing(db *sql.DB, table, column, columnType string) error {
	exists, err := columnExists(db, table, column)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	_, err = db.Exec(fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", table, column, columnType))
	return err
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
