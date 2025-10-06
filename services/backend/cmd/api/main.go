package main

import (
	"fmt"
	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"video-archiver/internal/api/handlers"
	"video-archiver/internal/config"
	"video-archiver/internal/repositories/sqlite"
	"video-archiver/internal/services/download"
	"video-archiver/internal/util/version"
)

func main() {
	log.SetReportCaller(true)

	if os.Getenv("DEBUG") == "true" {
		log.SetLevel(log.DebugLevel)
		log.Debug("Debug logging enabled")
	} else {
		log.SetLevel(log.InfoLevel)
	}

	fmt.Printf(`
__     _____ ____  _____ ___                       
\ \   / /_ _|  _ \| ____/ _ \                      
 \ \ / / | || | | |  _|| | | |                     
  \ V /  | || |_| | |__| |_| |                     
   \_/  |___|____/|_____\___/___     _______ ____  
   / \  |  _ \ / ___| | | |_ _\ \   / / ____|  _ \ 
  / _ \ | |_) | |   | |_| || | \ \ / /|  _| | |_) |
 / ___ \|  _ <| |___|  _  || |  \ V / | |___|  _ < 
/_/   \_\_| \_\\____|_| |_|___|  \_/  |_____|_| \_\

`)

	version.StartVersionChecker()

	fmt.Println("Loading configs...")
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	fmt.Println("Initializing database...")
	dbPath := os.Getenv("DATABASE_PATH")
	if dbPath == "" {
		dbPath = "./data/db/video-archiver.db"
	}

	db, err := sqlite.NewDB(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	jobRepo := sqlite.NewJobRepository(db)

	fmt.Println("Starting Download Service...")
	downloadService := download.NewService(&download.Config{
		JobRepository: jobRepo,
		DownloadPath:  cfg.Server.DownloadPath,
		Concurrency:   cfg.YtDlp.Concurrency,
		MaxQuality:    cfg.YtDlp.MaxQuality,
	})

	if err := downloadService.Start(); err != nil {
		log.Fatalf("Failed to start download service: %v", err)
	}
	defer downloadService.Stop()

	handler := handlers.NewHandler(downloadService, os.Getenv("DOWNLOAD_PATH"))

	apiRouter := chi.NewRouter()
	handler.RegisterRoutes(apiRouter)

	wsRouter := chi.NewRouter()
	handler.RegisterWSRoutes(wsRouter)

	go func() {
		fmt.Printf("Starting WebSocket server on %s...\n", cfg.Server.WSPort)
		if err := http.ListenAndServe(cfg.Server.WSPort, wsRouter); err != nil {
			log.Fatal(err)
		}
	}()

	fmt.Printf("Starting API server on %s...\n", cfg.Server.Port)
	if err := http.ListenAndServe(cfg.Server.Port, apiRouter); err != nil {
		log.Fatal(err)
	}
}
