package version

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	ytDlpReleasesURL = "https://api.github.com/repos/yt-dlp/yt-dlp/releases/latest"
	checkInterval    = 24 * time.Hour // Check once per day
)

type GitHubRelease struct {
	TagName string `json:"tag_name"`
	Name    string `json:"name"`
	HTMLURL string `json:"html_url"`
}

// StartVersionChecker periodically checks if yt-dlp is up to date
func StartVersionChecker() {
	// Check immediately on startup
	checkYtDlpVersion()

	// Then check periodically
	ticker := time.NewTicker(checkInterval)
	go func() {
		for range ticker.C {
			checkYtDlpVersion()
		}
	}()
}

// checkYtDlpVersion compares the current yt-dlp version with the latest release on GitHub
func checkYtDlpVersion() {
	currentVersion, err := getCurrentYtDlpVersion()
	if err != nil {
		log.WithError(err).Warn("Failed to get current yt-dlp version")
		return
	}

	latestVersion, releaseURL, err := getLatestYtDlpVersion()
	if err != nil {
		log.WithError(err).Warn("Failed to get latest yt-dlp version from GitHub")
		return
	}

	currentVersion = strings.TrimPrefix(currentVersion, "v")
	latestVersion = strings.TrimPrefix(latestVersion, "v")

	if currentVersion != latestVersion {
		fmt.Println("\n\n" + strings.Repeat("=", 80))
		fmt.Println("⚠️  YT-DLP VERSION ALERT ⚠️")
		fmt.Println(strings.Repeat("=", 80))
		fmt.Printf("Current yt-dlp version: %s\n", currentVersion)
		fmt.Printf("Latest yt-dlp version: %s\n", latestVersion)
		fmt.Printf("Please update your Dockerfile to use the latest version: %s\n", releaseURL)
		fmt.Println(strings.Repeat("=", 80) + "\n\n")

		log.WithFields(log.Fields{
			"current_version": currentVersion,
			"latest_version":  latestVersion,
			"release_url":     releaseURL,
		}).Warn("yt-dlp is outdated")
	} else {
		fmt.Println("yt-dlp is up to date (" + currentVersion + ")")
	}
}

func getCurrentYtDlpVersion() (string, error) {
	cmd := exec.Command("yt-dlp", "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to execute yt-dlp --version: %w", err)
	}

	version := strings.TrimSpace(string(output))
	return version, nil
}

func getLatestYtDlpVersion() (version string, url string, err error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(ytDlpReleasesURL)
	if err != nil {
		return "", "", fmt.Errorf("failed to fetch releases: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("GitHub API returned non-OK status: %d", resp.StatusCode)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", "", fmt.Errorf("failed to decode GitHub response: %w", err)
	}

	return release.TagName, release.HTMLURL, nil
}
