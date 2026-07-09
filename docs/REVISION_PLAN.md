# Revision Plan

A phased roadmap for revising the codebase after two years of accumulated growth. The stack stays as-is (Go backend, Vite + React SPA, SQLite, Docker Compose, yt-dlp + ffmpeg); each phase is an incremental refactor that leaves the app deployable. The tagging/library feature set (PR #387) and the tools service are treated as baseline to preserve — the tools package is the reference pattern for how the rest of the backend should look.

## Root causes this plan addresses

- **Playback (paths):** on-disk files were located by *reconstructing* `DOWNLOAD_PATH/<uploader>/<title>.<ext>` with a homemade sanitizer that does not match yt-dlp's filename sanitization (`/`→`⧸`, `:`→`：`, …). When the uploader directory name contains any such character, lookup fails even though the file exists. Deletion uses the same resolution, silently orphaning files.
- **Playback (codecs):** `--format "bestvideo+bestaudio" --merge-output-format mp4` *remuxes* VP9/AV1 + Opus into an `.mp4` container without transcoding. Browsers (Safari especially) cannot decode those files, so local videos fail to play with a generic error.
- **Backend stability:** WebSocket hub data races (map mutation under `RLock`, concurrent writers per connection), SQLite without WAL/busy-timeout ("database is locked"), duplicate job INSERTs, doubled progress trackers, a blocking queue send in the HTTP path, shutdown hangs, and several hundred lines of dead code.
- **Frontend structure:** ~8 hand-rolled fetch/useState blocks with no caching, three competing theme systems (one entirely dead), field-sniffing WebSocket message routing, broken channel→video navigation, video-card markup duplicated five times, and `Metadata` generated as `any`.
- **Tooling:** no dev mode at all — every code change requires a full Docker image rebuild; no Vite proxy or build tuning; backend URLs baked into the frontend image at build time; CI never lints and builds images cold; Go/Node/pnpm versions drift across files.

## Phase 1 — Video playback + path handling

**Goal:** videos reliably play in the browser; the file path is recorded at download time instead of reconstructed; legacy incompatible files get a one-click transcode.

- **Store the actual file path.** New `file_path` column on `jobs` (schema + migration). Captured from yt-dlp via `--print-to-file "after_move:%(filepath)s"` (single videos) and `after_move:%(id)s\t%(filepath)s` (playlists/channels) — a print *file* so it can't interleave with the `--progress-template` stdout stream, and `after_move` so it's the final merged path.
- **Resolution precedence.** `ResolveVideoFileWithHint`: stored path → exact reconstruction → in-directory normalized title match → *new* one-level normalized directory scan (fixes yt-dlp's `⧸`/`：` directory names for legacy rows). Successful fallback resolution writes the path back (lazy self-heal); playback, tools input resolution, and delete all use it.
- **Browser-safe future downloads.** Replace the format filter with `-f bv*+ba/b -S res:<N>,vcodec:h264,acodec:m4a` (sorting, not filtering: prefers avc1+aac at ≤N, degrades gracefully), keeping `--merge-output-format mp4`.
- **Transcode-on-demand.** `GET /video/{id}/playback-info` probes the file with ffprobe and reports codecs + browser compatibility + any existing transcode; `POST /video/{id}/transcode` submits a regular tools `convert` job (libx264/aac, `+faststart`); output is served by the existing tools output endpoint. The player shows a codec notice with a "Create compatible version" button and swaps sources when the transcode finishes.
- **Fix tools output persistence.** Mount `./data/processed` and set `PROCESSED_PATH` in docker-compose — previously all tools output was written to the ephemeral container layer.

## Phase 2 — Backend correctness & stability

**Goal:** no data races, no "database is locked", clean lifecycle.

- Rewrite the WebSocket hub: per-client buffered send channel + a single writer goroutine per connection (gorilla/websocket forbids concurrent writers); hub owns its maps in one goroutine; slow clients get dropped instead of stalling all producers; hub gains `Stop()`.
- SQLite: `?_journal_mode=WAL&_busy_timeout=5000&_foreign_keys=on`, `SetMaxOpenConns(1)`; replace ad-hoc migrations with a `PRAGMA user_version`-ordered list; embed `schema.sql` with `go:embed`.
- Download service surgery: remove the duplicate job INSERT, make the queue send non-blocking (return "queue full"), emit a single final progress update, tie workers/hub to a service context, remove the detached 5-minute goroutine that can hang shutdown.
- Delete dead code: `migration.go`, `metadataPaths`, `processPlaylistMetadata`, `isPlaylistOrChannel`, `GetRecent`, the `is_playlist` column.
- `http.Server` with timeouts + graceful shutdown ordering; move `PROCESSED_PATH` into the config struct; stop silently dropping `tools_*` fields in settings updates.
- Verify with `go test -race ./...` and a concurrent-download + slow-WS-client soak.

## Phase 3 — Dev tooling & deployment shape

**Goal:** sub-second iteration loop; runtime-portable images.

- `docker-compose.dev.yml`: backend via Air with a source mount, web via the Vite dev server with a source mount; `./run.sh dev`.
- Vite `server.proxy` for `/api` + `/ws` so dev needs no `VITE_SERVER_URL` and CORS disappears; add `build.rollupOptions.manualChunks`.
- Prod: nginx serves the SPA and proxies `/api` + `/ws` to the backend; the frontend switches to relative URLs; the WebSocket listener consolidates onto the API port (drop :8081). Images stop baking in URLs at build time.

## Phase 4 — Frontend architecture

**Goal:** one data layer, one theme system, a typed WS protocol.

- TanStack Query + a single typed API client; migrate all hand-rolled fetch blocks; one `PaginatedResponse<T>`; playback-info polling becomes `refetchInterval`.
- Typed WS envelope `{type, payload}` for *all* messages (download progress joins the existing typed tools messages); route by type instead of field-sniffing; exponential-backoff reconnect; explicit connect from the app root.
- One theme system (keep the settings-store implementation; delete the dead ThemeProvider/ModeToggle; the `index.html` pre-paint script reads the same key).
- Replace `Metadata = any` with a discriminated union wrapper over the tygo output; remove scattered casts.
- Fix channel→video navigation using `/job/{id}/videos` (real job IDs instead of YouTube video IDs).

## Phase 5 — UI/design consolidation

**Goal:** one look, accessible, resilient.

- Shared `VideoCard` (replaces five duplicated card markups) and `PageShell` (normalizes three padding conventions).
- Fix the unguarded `formatResolution` throw, dedupe `formatBytes`, remove decorative focusable Play buttons, add aria-labels to icon buttons, keyboard controls for the player, local poster instead of remote YouTube thumbnails.
- 404 route + top-level error boundary; style the sidebar trigger; refactor `Workflow.tsx` (630 lines) onto the shared `ToolPageShell`/`useToolSubmit`.

## Phase 6 — CI, dependencies, testing hardening

**Goal:** drift can't come back.

- chi → chi/v5, logrus → slog, align Go / Node / pnpm versions across go.mod, Dockerfiles, CI, and package.json.
- CI: real lint gates (golangci-lint, eslint, prettier check), `go test -race`, Buildx `cache-from`/`cache-to`, and a tygo drift check (regenerate types, `git diff --exit-code`).
- Tests for progress parsing (including non-YouTube extractors), the playback-info/transcode handlers, and the frontend API client / WS / player error states; replace YouTube-only progress heuristics; store the real `webpage_url` instead of fabricating virtual-job URLs.

## Ordering rationale

Playback first — it is the top user-facing pain and its schema change is self-contained. Stability (2) before tooling (3) so later verification runs on a trustworthy core. Hot reload (3) before the frontend-heavy phases (4, 5) because it makes them dramatically cheaper. Shared UI (5) builds on the final data layer (4). Lint/CI gates (6) land last, once churn stops.
