# video-archiver
A YouTube Video Archiver with Webinterface.

> [!CAUTION]
> This project is under active development and likely to not work when you clone it.

## Run it

````bash
./run.sh
````

or with `--clear` to clear the database

````bash
./run.sh --clear
````

**Debug backend container:**

````bash
docker-compose exec backend sh
````

<br>

## Planned Features
- [x] Download any YouTube video, playlist or channel
  - [ ] Select Download Quality
  - [ ] Store Metadata
- [ ] Tools
  - [ ] Video to audio
  - [ ] Basic video cutting
  - [ ] Basic merging of Videos
  - [ ] Video/Audio to text
- [ ] Dashboard
  - [ ] Display downloaded elements
  - [ ] Usage statistics
  - [ ] Full text search
- [ ] Overview
  - [x] Show current downloads and progress
  - [ ] Show recent downloads
- [ ] Streaming
  - [ ] Stream to other devices 

