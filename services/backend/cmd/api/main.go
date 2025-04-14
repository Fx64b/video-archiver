package main

import (
	"fmt"
	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"video-archiver/internal/api/handlers"
	"video-archiver/internal/repositories/sqlite"
	"video-archiver/internal/services/download"
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
		DownloadPath:  os.Getenv("DOWNLOAD_PATH"),
		// TODO: load these values from environment variables
		Concurrency: 2,
		MaxQuality:  1080,
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
		fmt.Println("Starting WebSocket server on :8081...")
		if err := http.ListenAndServe(":8081", wsRouter); err != nil {
			log.Fatal(err)
		}
	}()

	fmt.Println("Starting API server on :8080...")
	if err := http.ListenAndServe(":8080", apiRouter); err != nil {
		log.Fatal(err)
	}
}
