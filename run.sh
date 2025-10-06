#!/bin/bash

set -e

export CURRENT_UID=$(id -u)
export CURRENT_GID=$(id -g)
export COMPOSE_BAKE=true

DEBUG=""
SERVICES=""

while [[ $# -gt 0 ]]; do
  case $1 in
    --debug)
      export DEBUG=true
      shift
      ;;
    --clear)
      docker-compose down
      rm -rf data/db data/downloads data/cache
      shift
      ;;
    --build)
      BUILD="--build"
      shift
      ;;
    --backend-only)
      SERVICES="backend"
      shift
      ;;
    --test)
      echo "Running backend tests..."
      cd "$(dirname "$0")/services/backend"
      make test
      exit 0
      ;;
    --test-verbose)
      echo "Running backend tests with verbose output..."
      cd "$(dirname "$0")/services/backend"
      make test-verbose
      exit 0
      ;;
    --test-coverage)
      echo "Running backend tests with coverage..."
      cd "$(dirname "$0")/services/backend"
      make test-coverage
      exit 0
      ;;
    --help|-h)
      echo "This script is used to run the application in a docker container."
      echo "Available flags:"
      echo "  --clear           Clear the database and downloads"
      echo "  --build           Rebuild the containers"
      echo "  --debug           Enable debug logging in backend"
      echo "  --backend-only    Start only the backend service"
      echo "  --test            Run backend unit tests"
      echo "  --test-verbose    Run backend tests with verbose output"
      echo "  --test-coverage   Run backend tests with coverage report"
      echo "  --help|-h         Show this help message"
      exit 0
      ;;
    *)
      shift
      ;;
  esac
done

echo "Setting up directories and permissions..."
mkdir -p data/db data/downloads data/cache
chmod 777 data/db
chmod 755 data/downloads data/cache
chown -R $CURRENT_UID:$CURRENT_GID data/


clear
echo "./run.sh --help for more information."

if [ -n "$BUILD" ]; then
   echo "Generating TypeScript types..."
   cd services/backend

   # Check if tygo is installed, install if not
   if ! command -v tygo &> /dev/null; then
       echo "tygo not found. Installing..."
       go install github.com/gzuidhof/tygo@latest
   fi

   tygo generate
   mkdir -p ../../web/types
   cp generated/types/index.ts ../../web/types/index.ts
   cd ../..

  # The prune is optional, but during development it is likely that you will end up with several GB of cache
   docker builder prune -f --filter 'until=48h'

   docker-compose up --build --remove-orphans $SERVICES
else
   docker-compose up --remove-orphans $SERVICES
fi
