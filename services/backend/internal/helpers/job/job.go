package job

import "strings"

const CHANNEL_IDENTIFIER = "youtube.com/@"
const PLAYLIST_IDENTIFIER = "youtube.com/playlist?list="

func IsPlaylist(url string) bool {
	return strings.Contains(url, PLAYLIST_IDENTIFIER)
}

func isChannel(url string) bool {
	return strings.Contains(url, CHANNEL_IDENTIFIER)
}

func IsVideo(url string) bool {
	return !IsPlaylist(url) && !isChannel(url)
}
