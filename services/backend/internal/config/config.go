package config

type Config struct {
	Server struct {
		Port         int    `env:"PORT" envDefault:"8080"`
		WSPort       int    `env:"WS_PORT" envDefault:"8081"`
		DownloadPath string `env:"DOWNLOAD_PATH" envDefault:"./data/downloads"`
		DatabasePath string `env:"DATABASE_PATH" envDefault:"./data/db/video-archiver.db"`
	}
	YtDlp struct {
		Concurrency int    `env:"YTDLP_CONCURRENCY" envDefault:"2"`
		MaxQuality  string `env:"YTDLP_MAX_QUALITY" envDefault:"1080"`
	}
}

func Load() (*Config, error) {
	cfg := &Config{}
	// TODO: use a proper env parsing library like envconfig
	return cfg, nil
}
