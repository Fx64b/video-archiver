package testutil

// VideoMetadataJSON is a sample yt-dlp video metadata JSON
const VideoMetadataJSON = `{
  "id": "dQw4w9WgXcQ",
  "title": "Rick Astley - Never Gonna Give You Up (Official Video)",
  "description": "The official video for Never Gonna Give You Up",
  "thumbnail": "https://i.ytimg.com/vi/dQw4w9WgXcQ/maxresdefault.jpg",
  "duration": 212,
  "duration_string": "3:32",
  "view_count": 1400000000,
  "like_count": 15000000,
  "comment_count": 2500000,
  "channel": "Rick Astley",
  "channel_id": "UCuAXFkgsw1L7xaCfnd5JJOw",
  "channel_url": "https://www.youtube.com/channel/UCuAXFkgsw1L7xaCfnd5JJOw",
  "channel_follower_count": 3500000,
  "channel_is_verified": true,
  "uploader": "Rick Astley",
  "uploader_id": "@RickAstley",
  "uploader_url": "https://www.youtube.com/@RickAstley",
  "tags": ["Rick Astley", "Never Gonna Give You Up", "Music"],
  "categories": ["Music"],
  "upload_date": "20091024",
  "filesize_approx": 25000000,
  "format": "1080p",
  "ext": "mp4",
  "language": "en",
  "width": 1920,
  "height": 1080,
  "resolution": "1920x1080",
  "fps": 30.0,
  "dynamic_range": "SDR",
  "vcodec": "avc1.640028",
  "aspect_ratio": 1.78,
  "acodec": "mp4a.40.2",
  "audio_channels": 2,
  "was_live": false,
  "webpage_url_domain": "youtube.com",
  "extractor": "youtube",
  "fulltitle": "Rick Astley - Never Gonna Give You Up (Official Video)",
  "_type": "video"
}`

// PlaylistMetadataJSON is a sample yt-dlp playlist metadata JSON
const PlaylistMetadataJSON = `{
  "id": "PLrAXtmErZgOeiKm4sgNOknGvNjby9efdf",
  "title": "Best Music Videos",
  "description": "A collection of the best music videos",
  "thumbnails": [
    {
      "url": "https://i.ytimg.com/vi/dQw4w9WgXcQ/maxresdefault.jpg",
      "height": 720,
      "width": 1280,
      "id": "0"
    }
  ],
  "uploader_id": "@TestUser",
  "uploader_url": "https://www.youtube.com/@TestUser",
  "channel_id": "UCtest123",
  "channel": "Test User",
  "channel_url": "https://www.youtube.com/channel/UCtest123",
  "channel_follower_count": 100000,
  "playlist_count": 25,
  "view_count": 5000000,
  "_type": "playlist"
}`

// ChannelMetadataJSON is a sample yt-dlp channel metadata JSON
const ChannelMetadataJSON = `{
  "id": "UCuAXFkgsw1L7xaCfnd5JJOw",
  "title": "Rick Astley - Videos",
  "channel": "Rick Astley",
  "channel_url": "https://www.youtube.com/@RickAstley",
  "description": "Official Rick Astley YouTube Channel",
  "thumbnails": [
    {
      "url": "https://yt3.googleusercontent.com/test",
      "height": 800,
      "width": 800,
      "id": "avatar"
    }
  ],
  "channel_follower_count": 3500000,
  "playlist_count": 10,
  "channel_id": "UCuAXFkgsw1L7xaCfnd5JJOw",
  "_type": "playlist"
}`

// YtDlpProgressOutput simulates yt-dlp progress output
const YtDlpProgressOutput = `[youtube] Extracting URL: https://www.youtube.com/watch?v=dQw4w9WgXcQ
[youtube] dQw4w9WgXcQ: Downloading webpage
[youtube] dQw4w9WgXcQ: Downloading tv client config
[info] Writing video metadata as JSON to: /downloads/Rick Astley/Never Gonna Give You Up.info.json
[download] Destination: /downloads/Rick Astley/Never Gonna Give You Up [dQw4w9WgXcQ].f401.mp4
[1][NA][dQw4w9WgXcQ][Never Gonna Give You Up][401][1080p][avc1][none]prog:[1048576/20971520][   5.0%][2.5MiB/s][00:08]
[1][NA][dQw4w9WgXcQ][Never Gonna Give You Up][401][1080p][avc1][none]prog:[5242880/20971520][  25.0%][2.8MiB/s][00:06]
[1][NA][dQw4w9WgXcQ][Never Gonna Give You Up][401][1080p][avc1][none]prog:[10485760/20971520][  50.0%][3.0MiB/s][00:04]
[1][NA][dQw4w9WgXcQ][Never Gonna Give You Up][401][1080p][avc1][none]prog:[15728640/20971520][  75.0%][3.2MiB/s][00:02]
[1][NA][dQw4w9WgXcQ][Never Gonna Give You Up][401][1080p][avc1][none]prog:[20971520/20971520][ 100.0%][3.3MiB/s][00:00]
[download] Destination: /downloads/Rick Astley/Never Gonna Give You Up [dQw4w9WgXcQ].f251.webm
[1][NA][dQw4w9WgXcQ][Never Gonna Give You Up][251][opus][none][opus]prog:[524288/5242880][  10.0%][1.2MiB/s][00:04]
[1][NA][dQw4w9WgXcQ][Never Gonna Give You Up][251][opus][none][opus]prog:[2621440/5242880][  50.0%][1.5MiB/s][00:02]
[1][NA][dQw4w9WgXcQ][Never Gonna Give You Up][251][opus][none][opus]prog:[5242880/5242880][ 100.0%][1.6MiB/s][00:00]
[Merger] Merging formats into "/downloads/Rick Astley/Never Gonna Give You Up [dQw4w9WgXcQ].mp4"
Deleting original file /downloads/Rick Astley/Never Gonna Give You Up [dQw4w9WgXcQ].f401.mp4
Deleting original file /downloads/Rick Astley/Never Gonna Give You Up [dQw4w9WgXcQ].f251.webm`

// YtDlpPlaylistProgressOutput simulates yt-dlp playlist download progress
const YtDlpPlaylistProgressOutput = `[youtube:tab] Extracting URL: https://www.youtube.com/playlist?list=PLtest
[youtube:tab] PLtest: Downloading playlist metadata
[download] Downloading playlist: Best Music Videos
[download] Downloading item 1 of 3
[youtube] Extracting URL: https://www.youtube.com/watch?v=video1
[youtube] video1: Downloading webpage
[download] Destination: /downloads/Channel/Video 1 [video1].f401.mp4
[3][1][video1][Video 1][401][1080p][avc1][none]prog:[10485760/20971520][  50.0%][3.0MiB/s][00:04]
[3][1][video1][Video 1][401][1080p][avc1][none]prog:[20971520/20971520][ 100.0%][3.3MiB/s][00:00]
[download] Downloading item 2 of 3
[youtube] Extracting URL: https://www.youtube.com/watch?v=video2
[youtube] video2: Downloading webpage
[download] Destination: /downloads/Channel/Video 2 [video2].f401.mp4
[3][2][video2][Video 2][401][1080p][avc1][none]prog:[10485760/20971520][  50.0%][3.0MiB/s][00:04]
[download] Downloading item 3 of 3
[youtube] video3 has already been downloaded
[download] 100% of playlist downloaded`

// YtDlpAlreadyDownloadedOutput simulates already downloaded scenario
const YtDlpAlreadyDownloadedOutput = `[youtube] Extracting URL: https://www.youtube.com/watch?v=dQw4w9WgXcQ
[youtube] dQw4w9WgXcQ: Downloading webpage
[download] /downloads/Rick Astley/Never Gonna Give You Up.mp4 has already been downloaded
[download] 100% of 1`

// ArchiveFileContent is sample content for yt-dlp archive file
const ArchiveFileContent = `youtube dQw4w9WgXcQ
youtube abc123def456
youtube xyz789uvw012`
