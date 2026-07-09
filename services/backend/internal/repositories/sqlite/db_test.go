package sqlite

import (
	"database/sql"
	"path/filepath"
	"testing"
)

func TestNewDBFreshSchema(t *testing.T) {
	db, err := NewDB(filepath.Join(t.TempDir(), "fresh.db"))
	if err != nil {
		t.Fatalf("NewDB: %v", err)
	}
	defer db.Close()

	v, err := userVersion(db)
	if err != nil {
		t.Fatalf("user_version: %v", err)
	}
	if v != len(migrations) {
		t.Errorf("fresh database user_version = %d, want %d", v, len(migrations))
	}

	for _, col := range []string{"warnings", "file_path"} {
		ok, err := columnExists(db, "jobs", col)
		if err != nil || !ok {
			t.Errorf("fresh schema missing jobs.%s (err=%v)", col, err)
		}
	}
	for _, table := range []string{"tags", "job_tags", "tools_jobs"} {
		ok, err := tableExists(db, table)
		if err != nil || !ok {
			t.Errorf("fresh schema missing table %s (err=%v)", table, err)
		}
	}
}

func TestNewDBMigratesLegacyDatabase(t *testing.T) {
	path := filepath.Join(t.TempDir(), "legacy.db")

	// Simulate a database from before versioning: jobs without the migrated
	// columns, no tag tables, user_version 0.
	raw, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := raw.Exec(`
        CREATE TABLE jobs (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            job_id TEXT UNIQUE,
            url TEXT NOT NULL,
            status TEXT,
            progress REAL,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        );
        INSERT INTO jobs (job_id, url, status, progress) VALUES ('old-1', 'https://example.com', 'complete', 100);
    `); err != nil {
		t.Fatal(err)
	}
	raw.Close()

	db, err := NewDB(path)
	if err != nil {
		t.Fatalf("NewDB on legacy database: %v", err)
	}
	defer db.Close()

	v, err := userVersion(db)
	if err != nil || v != len(migrations) {
		t.Errorf("migrated user_version = %d (err=%v), want %d", v, err, len(migrations))
	}
	for _, col := range []string{"warnings", "file_path"} {
		ok, err := columnExists(db, "jobs", col)
		if err != nil || !ok {
			t.Errorf("migration did not add jobs.%s (err=%v)", col, err)
		}
	}
	ok, err := tableExists(db, "tags")
	if err != nil || !ok {
		t.Errorf("migration did not create tags table (err=%v)", err)
	}

	// Existing data must survive.
	var url string
	if err := db.QueryRow("SELECT url FROM jobs WHERE job_id = 'old-1'").Scan(&url); err != nil {
		t.Errorf("legacy row lost after migration: %v", err)
	}
}

func TestNewDBReopenIsIdempotent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "reopen.db")

	db1, err := NewDB(path)
	if err != nil {
		t.Fatal(err)
	}
	db1.Close()

	db2, err := NewDB(path)
	if err != nil {
		t.Fatalf("reopening migrated database: %v", err)
	}
	defer db2.Close()

	v, err := userVersion(db2)
	if err != nil || v != len(migrations) {
		t.Errorf("reopened user_version = %d (err=%v), want %d", v, err, len(migrations))
	}
}
