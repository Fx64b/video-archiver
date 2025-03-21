#!/bin/bash

set -e

export CURRENT_UID=$(id -u)
export CURRENT_GID=$(id -g)
DEBUG=""
export COMPOSE_BAKE=true

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
    --help|-h)
      echo "This script is used to run the application in a docker container."
      echo "Available flags:"
      echo "  --clear    Clear the database and downloads"
      echo "  --build    Rebuild the containers"
      echo "  --debug    Enable debug logging in backend"
      echo "  --help|-h  Show this help message"
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
   tygo generate
   mkdir -p ../../web/types
   cp generated/types/index.ts ../../web/types/index.ts
   cd ../..
   docker-compose up --build --remove-orphans
else
   docker-compose up --remove-orphans
fi
