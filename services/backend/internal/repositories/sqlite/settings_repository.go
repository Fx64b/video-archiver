package sqlite

import (
	"database/sql"
	"fmt"
	"time"
	"video-archiver/internal/domain"
)

type SettingsRepository struct {
	db *sql.DB
}

func NewSettingsRepository(db *sql.DB) *SettingsRepository {
	return &SettingsRepository{db: db}
}

func (r *SettingsRepository) Get() (*domain.Settings, error) {
	settings := &domain.Settings{}
	err := r.db.QueryRow(`
        SELECT id, theme, download_quality, concurrent_downloads, tools_default_format,
               tools_default_quality, tools_preserve_original, tools_output_path, created_at, updated_at
        FROM settings
        WHERE id = 1`).
		Scan(&settings.ID, &settings.Theme, &settings.DownloadQuality, &settings.ConcurrentDownloads,
			&settings.ToolsDefaultFormat, &settings.ToolsDefaultQuality, &settings.ToolsPreserveOriginal,
			&settings.ToolsOutputPath, &settings.CreatedAt, &settings.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get settings: %w", err)
	}
	return settings, nil
}

func (r *SettingsRepository) Update(settings *domain.Settings) error {
	settings.UpdatedAt = time.Now()
	_, err := r.db.Exec(`
        UPDATE settings
        SET theme = ?, download_quality = ?, concurrent_downloads = ?,
            tools_default_format = ?, tools_default_quality = ?,
            tools_preserve_original = ?, tools_output_path = ?, updated_at = ?
        WHERE id = 1`,
		settings.Theme, settings.DownloadQuality, settings.ConcurrentDownloads,
		settings.ToolsDefaultFormat, settings.ToolsDefaultQuality,
		settings.ToolsPreserveOriginal, settings.ToolsOutputPath, settings.UpdatedAt)
	if err != nil {
		return fmt.Errorf("update settings: %w", err)
	}
	return nil
}
