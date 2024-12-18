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

**Working on:** Show recent downloads


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
  - [ ] Categorization
- [ ] Overview
  - [x] Show current downloads and progress
  - [ ] Show recent downloads
- [ ] Streaming
  - [ ] Stream to other devices 

<br>

## Known Issues
First time startup after clonening can take up to 150 seconds due to docker downloading images and go building the app.
