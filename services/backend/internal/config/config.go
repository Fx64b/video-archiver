package config

import (
	"github.com/caarlos0/env/v11"
)

type Config struct {
	Server struct {
		Port          string `env:"PORT" envDefault:":8080"`
		DownloadPath  string `env:"DOWNLOAD_PATH" envDefault:"./data/downloads"`
		ProcessedPath string `env:"PROCESSED_PATH" envDefault:"./data/processed"`
		DatabasePath  string `env:"DATABASE_PATH" envDefault:"./data/db/video-archiver.db"`
	}
	YtDlp struct {
		Concurrency int `env:"YTDLP_CONCURRENCY" envDefault:"2"`
		MaxQuality  int `env:"YTDLP_MAX_QUALITY" envDefault:"1080"`
	}
}

func Load() (*Config, error) {
	cfg := &Config{}

	err := env.Parse(cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
