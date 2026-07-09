#!/usr/bin/env bash
#
# video-archiver — development and deployment management script.
#
# Run './run.sh help' for usage.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

GO_IMAGE="golang:1.25-alpine"

# ------------------------------------------------------------------------------
# Output helpers
# ------------------------------------------------------------------------------

if [[ -t 1 && -z "${NO_COLOR:-}" ]]; then
    BOLD=$'\033[1m' RED=$'\033[31m' BLUE=$'\033[34m' RESET=$'\033[0m'
else
    BOLD='' RED='' BLUE='' RESET=''
fi

info()  { echo "${BLUE}==>${RESET} ${BOLD}$*${RESET}"; }
die()   { echo "${RED}error:${RESET} $*" >&2; exit 1; }

usage() {
    cat <<EOF
${BOLD}video-archiver${RESET} — self-hosted YouTube archiver

${BOLD}Usage:${RESET}
  ./run.sh [command] [options]

${BOLD}Commands:${RESET}
  start             Start the application (default; builds images on first run)
  dev               Start the development stack with hot reload (Air + Vite HMR)
  build             Rebuild images, then start
  stop              Stop the application
  logs [service]    Follow container logs (backend, web)
  test [target]     Run backend tests (target: coverage, verbose; default: all)
  types             Regenerate TypeScript types from the Go domain structs
  clean             Remove containers and locally built images
  reset             Stop and delete all data (database, downloads)
  help              Show this help message

${BOLD}Options:${RESET}
  -d, --detach      Run containers in the background
      --debug       Enable debug logging in the backend
      --backend-only
                    Operate on the backend service only
      --no-cache    (build) Rebuild images without using the build cache
  -y, --yes         Skip confirmation prompts

${BOLD}Examples:${RESET}
  ./run.sh                   # start everything
  ./run.sh dev               # develop with hot reload on :3000
  ./run.sh build --debug     # rebuild and start with debug logging
  ./run.sh reset --yes       # wipe all data without confirmation
  ./run.sh test coverage     # backend tests with coverage report
EOF
}

# ------------------------------------------------------------------------------
# Environment checks
# ------------------------------------------------------------------------------

require_docker() {
    command -v docker >/dev/null 2>&1 \
        || die "docker is not installed. See https://docs.docker.com/get-docker/"
    docker info >/dev/null 2>&1 \
        || die "the Docker daemon is not running (or you lack permission to use it)"

    if docker compose version >/dev/null 2>&1; then
        COMPOSE=(docker compose)
    elif command -v docker-compose >/dev/null 2>&1; then
        COMPOSE=(docker-compose)
    else
        die "Docker Compose is not available. See https://docs.docker.com/compose/install/"
    fi
}

prepare_env() {
    CURRENT_UID="$(id -u)" CURRENT_GID="$(id -g)"
    export CURRENT_UID CURRENT_GID
    mkdir -p data/db data/downloads data/processed data/cache
}

confirm() {
    [[ "$ASSUME_YES" == true ]] && return 0
    local reply
    read -r -p "$1 [y/N] " reply
    [[ "$reply" == [yY] || "$reply" == [yY][eE][sS] ]]
}

# ------------------------------------------------------------------------------
# Commands
# ------------------------------------------------------------------------------

cmd_start() {
    require_docker
    prepare_env
    local args=(up --remove-orphans)
    [[ "$DETACH" == true ]] && args+=(--detach)
    info "Starting video-archiver (web: http://localhost:3000, api: http://localhost:8080)"
    "${COMPOSE[@]}" "${args[@]}" ${SERVICES:+"$SERVICES"}
}

cmd_dev() {
    require_docker
    prepare_env
    local args=(-f docker-compose.dev.yml up --build --remove-orphans)
    [[ "$DETACH" == true ]] && args+=(--detach)
    info "Starting dev stack (web: http://localhost:3000, api: http://localhost:8080)"
    info "Go and frontend changes hot-reload; Ctrl-C to stop"
    "${COMPOSE[@]}" "${args[@]}" ${SERVICES:+"$SERVICES"}
}

cmd_build() {
    require_docker
    prepare_env
    local args=(build)
    [[ "$NO_CACHE" == true ]] && args+=(--no-cache --pull)
    info "Building images"
    "${COMPOSE[@]}" "${args[@]}" ${SERVICES:+"$SERVICES"}
    cmd_start
}

cmd_stop() {
    require_docker
    info "Stopping containers"
    "${COMPOSE[@]}" down --remove-orphans
}

cmd_logs() {
    require_docker
    "${COMPOSE[@]}" logs --follow ${1:+"$1"}
}

cmd_test() {
    local target="${1:-all}" make_target
    case "$target" in
        all)      make_target="test" ;;
        verbose)  make_target="test-verbose" ;;
        coverage) make_target="test-coverage" ;;
        *)        die "unknown test target '$target' (expected: all, verbose, coverage)" ;;
    esac

    if command -v go >/dev/null 2>&1; then
        info "Running backend tests ($target)"
        make -C services/backend "$make_target"
    else
        require_docker
        info "Go not found on host — running backend tests in Docker"
        docker run --rm \
            -v "$SCRIPT_DIR/services/backend":/src -w /src \
            "$GO_IMAGE" \
            sh -c "apk add --no-cache --quiet make && make $make_target"
    fi
}

cmd_types() {
    # The generated types are committed, so this is only needed after
    # changing the Go structs in services/backend/internal/domain.
    info "Generating TypeScript types from Go structs"
    if command -v tygo >/dev/null 2>&1; then
        (cd services/backend && tygo generate)
    elif command -v go >/dev/null 2>&1; then
        (cd services/backend && go run github.com/gzuidhof/tygo@latest generate)
    else
        require_docker
        info "Go not found on host — running tygo in Docker"
        docker run --rm \
            --user "$(id -u):$(id -g)" \
            -e HOME=/tmp -e GOCACHE=/tmp/gocache -e GOMODCACHE=/tmp/gomod \
            -v "$SCRIPT_DIR/services/backend":/src -w /src \
            "$GO_IMAGE" \
            go run github.com/gzuidhof/tygo@latest generate
    fi
    mkdir -p web/types
    cp services/backend/generated/types/index.ts web/types/index.ts
    info "Updated web/types/index.ts"
}

cmd_clean() {
    require_docker
    confirm "Remove containers and locally built images?" || die "aborted"
    "${COMPOSE[@]}" down --remove-orphans --rmi local
    info "Done. Build cache was kept; run 'docker builder prune' to clear it too."
}

wipe_data() {
    confirm "This deletes the database and ALL downloaded videos. Continue?" || die "aborted"
    "${COMPOSE[@]}" down --remove-orphans
    rm -rf data/db data/downloads data/processed data/cache
    info "All application data removed"
}

cmd_reset() {
    require_docker
    wipe_data
}

# ------------------------------------------------------------------------------
# Argument parsing
# ------------------------------------------------------------------------------

COMMAND=""
SERVICES=""
DETACH=false
NO_CACHE=false
ASSUME_YES=false
CLEAR_DATA=false
POSITIONAL=()

while [[ $# -gt 0 ]]; do
    case $1 in
        start|dev|build|stop|logs|test|types|clean|reset|help)
            [[ -n "$COMMAND" ]] && die "multiple commands given: '$COMMAND' and '$1'"
            COMMAND=$1
            ;;
        -d|--detach)    DETACH=true ;;
        --debug)        export DEBUG=true ;;
        --backend-only) SERVICES="backend" ;;
        --no-cache)     NO_CACHE=true ;;
        -y|--yes)       ASSUME_YES=true ;;
        -h|--help)      COMMAND="help" ;;
        # Legacy flags, kept for backwards compatibility:
        --build)         COMMAND="build" ;;
        --clear)         CLEAR_DATA=true ;;
        --test)          COMMAND="test"; POSITIONAL+=(all) ;;
        --test-verbose)  COMMAND="test"; POSITIONAL+=(verbose) ;;
        --test-coverage) COMMAND="test"; POSITIONAL+=(coverage) ;;
        -*)
            die "unknown option '$1' — run './run.sh help' for usage"
            ;;
        *)
            POSITIONAL+=("$1")
            ;;
    esac
    shift
done

case "${COMMAND:-start}" in
    logs|test) ;;  # these accept a positional argument
    *)
        if [[ ${#POSITIONAL[@]} -gt 0 ]]; then
            die "unknown command or argument '${POSITIONAL[0]}' — run './run.sh help' for usage"
        fi
        ;;
esac

if [[ "$CLEAR_DATA" == true ]]; then
    require_docker
    wipe_data
fi

case "${COMMAND:-start}" in
    start) cmd_start ;;
    build) cmd_build ;;
    stop)  cmd_stop ;;
    logs)  cmd_logs "${POSITIONAL[@]:-}" ;;
    test)  cmd_test "${POSITIONAL[@]:-}" ;;
    types) cmd_types ;;
    clean) cmd_clean ;;
    reset) cmd_reset ;;
    help)  usage ;;
esac
