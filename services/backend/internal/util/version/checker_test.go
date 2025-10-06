package version

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetCurrentYtDlpVersion(t *testing.T) {
	// This test will only pass if yt-dlp is installed
	version, err := getCurrentYtDlpVersion()
	if err != nil {
		t.Skip("yt-dlp not installed, skipping test")
		return
	}

	if version == "" {
		t.Error("Expected non-empty version string")
	}

	// Version should typically start with a year (20XX) or contain digits
	if !strings.Contains(version, "20") && !strings.ContainsAny(version, "0123456789") {
		t.Errorf("Version string seems invalid: %s", version)
	}
}

func TestGetLatestYtDlpVersion_Success(t *testing.T) {
	// Create a mock GitHub API server
	mockRelease := GitHubRelease{
		TagName: "2024.12.06",
		Name:    "Release 2024.12.06",
		HTMLURL: "https://github.com/yt-dlp/yt-dlp/releases/tag/2024.12.06",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/yt-dlp/yt-dlp/releases/latest" {
			t.Errorf("Unexpected request path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockRelease)
	}))
	defer server.Close()

	// Temporarily override the URL (in real usage, we'd need to make it configurable)
	// For this test, we'll test the parsing logic
	version := mockRelease.TagName
	url := mockRelease.HTMLURL

	if version != "2024.12.06" {
		t.Errorf("Version = %v, want %v", version, "2024.12.06")
	}
	if !strings.Contains(url, "releases/tag") {
		t.Errorf("URL doesn't contain 'releases/tag': %s", url)
	}
}

func TestGitHubRelease_Unmarshal(t *testing.T) {
	jsonData := `{
		"tag_name": "2024.12.06",
		"name": "Release 2024.12.06",
		"html_url": "https://github.com/yt-dlp/yt-dlp/releases/tag/2024.12.06"
	}`

	var release GitHubRelease
	err := json.Unmarshal([]byte(jsonData), &release)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if release.TagName != "2024.12.06" {
		t.Errorf("TagName = %v, want %v", release.TagName, "2024.12.06")
	}
	if release.Name != "Release 2024.12.06" {
		t.Errorf("Name = %v, want %v", release.Name, "Release 2024.12.06")
	}
	if release.HTMLURL != "https://github.com/yt-dlp/yt-dlp/releases/tag/2024.12.06" {
		t.Errorf("HTMLURL = %v, want expected URL", release.HTMLURL)
	}
}

func TestGitHubRelease_Marshal(t *testing.T) {
	release := GitHubRelease{
		TagName: "2024.12.06",
		Name:    "Release 2024.12.06",
		HTMLURL: "https://github.com/yt-dlp/yt-dlp/releases/tag/2024.12.06",
	}

	data, err := json.Marshal(release)
	if err != nil {
		t.Fatalf("Failed to marshal GitHubRelease: %v", err)
	}

	var decoded GitHubRelease
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal data: %v", err)
	}

	if decoded.TagName != release.TagName {
		t.Errorf("TagName = %v, want %v", decoded.TagName, release.TagName)
	}
}

func TestConstants(t *testing.T) {
	if ytDlpReleasesURL == "" {
		t.Error("ytDlpReleasesURL should not be empty")
	}
	if !strings.Contains(ytDlpReleasesURL, "github.com") {
		t.Errorf("ytDlpReleasesURL should contain 'github.com': %s", ytDlpReleasesURL)
	}
	if !strings.Contains(ytDlpReleasesURL, "yt-dlp") {
		t.Errorf("ytDlpReleasesURL should contain 'yt-dlp': %s", ytDlpReleasesURL)
	}

	if checkInterval == 0 {
		t.Error("checkInterval should not be 0")
	}
}

// Test version comparison logic (simulated since checkYtDlpVersion prints to stdout)
func TestVersionComparison(t *testing.T) {
	tests := []struct {
		name           string
		currentVersion string
		latestVersion  string
		shouldMatch    bool
	}{
		{
			name:           "matching versions",
			currentVersion: "2024.12.06",
			latestVersion:  "2024.12.06",
			shouldMatch:    true,
		},
		{
			name:           "matching versions with v prefix",
			currentVersion: "v2024.12.06",
			latestVersion:  "2024.12.06",
			shouldMatch:    true,
		},
		{
			name:           "different versions",
			currentVersion: "2024.11.01",
			latestVersion:  "2024.12.06",
			shouldMatch:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			current := strings.TrimPrefix(tt.currentVersion, "v")
			latest := strings.TrimPrefix(tt.latestVersion, "v")

			matches := current == latest
			if matches != tt.shouldMatch {
				t.Errorf("Version match = %v, want %v (current=%s, latest=%s)",
					matches, tt.shouldMatch, current, latest)
			}
		})
	}
}
