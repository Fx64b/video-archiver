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
	"video-archiver/internal/services/tools"
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
	db, err := sqlite.NewDB(cfg.Server.DatabasePath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	jobRepo := sqlite.NewJobRepository(db)
	settingsRepo := sqlite.NewSettingsRepository(db)
	toolsRepo := sqlite.NewToolsRepository(db)

	fmt.Println("Starting Download Service...")
	downloadService := download.NewService(&download.Config{
		JobRepository:      jobRepo,
		SettingsRepository: settingsRepo,
		DownloadPath:       cfg.Server.DownloadPath,
		Concurrency:        cfg.YtDlp.Concurrency,
		MaxQuality:         cfg.YtDlp.MaxQuality,
	})

	if err := downloadService.Start(); err != nil {
		log.Fatalf("Failed to start download service: %v", err)
	}
	defer downloadService.Stop()

	fmt.Println("Starting Tools Service...")
	processedPath := os.Getenv("PROCESSED_PATH")
	if processedPath == "" {
		processedPath = "./data/processed"
	}

	toolsService := tools.NewService(&tools.Config{
		ToolsRepository: toolsRepo,
		JobRepository:   jobRepo,
		// Reuse the download service's WebSocket hub so tools progress reaches
		// the frontend over the same /ws connection.
		Broadcaster:   downloadService.GetHub(),
		DownloadPath:  cfg.Server.DownloadPath,
		ProcessedPath: processedPath,
		Concurrency:   2,
	})

	if err := toolsService.Start(); err != nil {
		log.Fatalf("Failed to start tools service: %v", err)
	}
	defer toolsService.Stop()

	handler := handlers.NewHandler(downloadService, cfg.Server.DownloadPath, settingsRepo)
	toolsHandler := handlers.NewToolsHandler(toolsService)

	apiRouter := chi.NewRouter()
	handler.RegisterRoutes(apiRouter)
	toolsHandler.RegisterRoutes(apiRouter)

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
