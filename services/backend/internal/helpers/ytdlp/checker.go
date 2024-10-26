package ytdlp

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"os/exec"
)

func CheckAndInstall() error {
	if !isInstalled() {
		log.Warn("yt-dlp not found, and installation in a container is unsupported. Exiting.")
		return fmt.Errorf("yt-dlp must be pre-installed in containerized environments")
	}
	return nil
}

func isInstalled() bool {
	_, err := exec.LookPath("yt-dlp")
	return err == nil
}
