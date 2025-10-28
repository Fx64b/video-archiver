package config

import (
	"github.com/caarlos0/env/v11"
)

type Config struct {
	Server struct {
		Port           string `env:"PORT" envDefault:":8080"`
		WSPort         string `env:"WS_PORT" envDefault:":8081"`
		DownloadPath   string `env:"DOWNLOAD_PATH" envDefault:"./data/downloads"`
		DatabasePath   string `env:"DATABASE_PATH" envDefault:"./data/db/video-archiver.db"`
		AllowedOrigins string `env:"ALLOWED_ORIGINS" envDefault:"http://localhost:3000"`
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
