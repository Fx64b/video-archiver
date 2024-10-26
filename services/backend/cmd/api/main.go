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
)

func main() {
	log.SetReportCaller(true)

	if err := ytdlp.CheckAndInstall(); err != nil {
		log.Fatalf("yt-dlp setup failed: %v", err)
		os.Exit(1) // Terminate if yt-dlp setup fails
	}

	var r *chi.Mux = chi.NewRouter()
	handlers.Handler(r)

	fmt.Println("Starting Queue worker...")
	queue.StartQueueWorker()

	fmt.Println("Starting GO API service...")

	fmt.Printf(`

 ___      ___ ___  ________  _______   ________          ________  ________  ________  ___  ___  ___  ___      ___ _______   ________     
|\  \    /  /|\  \|\   ___ \|\  ___ \ |\   __  \        |\   __  \|\   __  \|\   ____\|\  \|\  \|\  \|\  \    /  /|\  ___ \ |\   __  \    
\ \  \  /  / | \  \ \  \_|\ \ \   __/|\ \  \|\  \       \ \  \|\  \ \  \|\  \ \  \___|\ \  \\\  \ \  \ \  \  /  / | \   __/|\ \  \|\  \   
 \ \  \/  / / \ \  \ \  \ \\ \ \  \_|/_\ \  \\\  \       \ \   __  \ \   _  _\ \  \    \ \   __  \ \  \ \  \/  / / \ \  \_|/_\ \   _  _\  
  \ \    / /   \ \  \ \  \_\\ \ \  \_|\ \ \  \\\  \       \ \  \ \  \ \  \\  \\ \  \____\ \  \ \  \ \  \ \    / /   \ \  \_|\ \ \  \\  \| 
   \ \__/ /     \ \__\ \_______\ \_______\ \_______\       \ \__\ \__\ \__\\ _\\ \_______\ \__\ \__\ \__\ \__/ /     \ \_______\ \__\\ _\ 
    \|__|/       \|__|\|_______|\|_______|\|_______|        \|__|\|__|\|__|\|__|\|_______|\|__|\|__|\|__|\|__|/       \|_______|\|__|\|__|
                                                                                                                                          
                                                                                                                                          
                                                                                                                                          

`)

	err := http.ListenAndServe("0.0.0.0:8080", r)
	if err != nil {
		log.Error(err)
	}
}
