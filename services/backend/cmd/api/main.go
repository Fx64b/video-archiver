package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
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

	// Tag items downloaded before auto-tagging existed; idempotent, so it can
	// run on every startup without growing the tag set.
	go func() {
		if err := jobRepo.BackfillAutoTags(); err != nil {
			log.WithError(err).Warn("Auto-tag backfill failed")
		}
	}()

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
	toolsService := tools.NewService(&tools.Config{
		ToolsRepository: toolsRepo,
		JobRepository:   jobRepo,
		// Reuse the download service's WebSocket hub so tools progress reaches
		// the frontend over the same /ws connection.
		Broadcaster:   downloadService.GetHub(),
		DownloadPath:  cfg.Server.DownloadPath,
		ProcessedPath: cfg.Server.ProcessedPath,
		Concurrency:   2,
	})

	if err := toolsService.Start(); err != nil {
		log.Fatalf("Failed to start tools service: %v", err)
	}
	defer toolsService.Stop()

	handler := handlers.NewHandler(downloadService, cfg.Server.DownloadPath, settingsRepo,
		toolsService, toolsRepo, tools.NewFFmpeg())
	toolsHandler := handlers.NewToolsHandler(toolsService)

	apiRouter := chi.NewRouter()
	handler.RegisterRoutes(apiRouter)
	toolsHandler.RegisterRoutes(apiRouter)

	wsRouter := chi.NewRouter()
	handler.RegisterWSRoutes(wsRouter)

	// Explicit timeouts so slow or stalled clients can't pin server resources
	// indefinitely. Write timeouts are deliberately absent: /video streams
	// large files and /ws connections are long-lived.
	apiServer := &http.Server{
		Addr:              cfg.Server.Port,
		Handler:           apiRouter,
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       120 * time.Second,
	}
	wsServer := &http.Server{
		Addr:              cfg.Server.WSPort,
		Handler:           wsRouter,
		ReadHeaderTimeout: 10 * time.Second,
	}

	serverErrors := make(chan error, 2)
	go func() {
		fmt.Printf("Starting WebSocket server on %s...\n", cfg.Server.WSPort)
		serverErrors <- wsServer.ListenAndServe()
	}()
	go func() {
		fmt.Printf("Starting API server on %s...\n", cfg.Server.Port)
		serverErrors <- apiServer.ListenAndServe()
	}()

	// Shutdown order: stop accepting HTTP traffic, then stop the services
	// (which cancels running yt-dlp/ffmpeg processes and closes the WebSocket
	// hub via their deferred Stop calls), then close the database (deferred).
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	select {
	case err := <-serverErrors:
		if err != nil && err != http.ErrServerClosed {
			log.WithError(err).Error("HTTP server failed")
		}
	case <-ctx.Done():
		fmt.Println("Shutting down...")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := apiServer.Shutdown(shutdownCtx); err != nil {
		log.WithError(err).Warn("API server shutdown incomplete")
	}
	if err := wsServer.Shutdown(shutdownCtx); err != nil {
		log.WithError(err).Warn("WebSocket server shutdown incomplete")
	}
}
