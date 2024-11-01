package main

import (
	"fmt"
	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"video-archiver/internal/handlers"
	"video-archiver/internal/helpers/ytdlp"
	"video-archiver/internal/queue"
	"video-archiver/internal/storage"
)

func main() {
	log.SetReportCaller(true)

	if err := ytdlp.CheckAndInstall(); err != nil {
		log.Fatalf("yt-dlp setup failed: %v", err)
		os.Exit(1) // Terminate if yt-dlp setup fails
	}

	var r *chi.Mux = chi.NewRouter()
	handlers.Handler(r)

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

	if err := storage.InitDB(dbPath); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
		os.Exit(1) // Terminate if db intialization fails
	}

	defer func() {
		err := storage.CloseDB()
		if err != nil {
			log.Errorf("Failed to close database: %v", err)
		}
	}()

	fmt.Println("Starting Queue worker...")
	go queue.StartQueueWorker()

	fmt.Println("Starting WebSocket service...")
	go queue.StartWebSocketServer()

	fmt.Println("Starting GO API service...")

	err := http.ListenAndServe("0.0.0.0:8080", r)
	if err != nil {
		log.Error(err)
	}
}
