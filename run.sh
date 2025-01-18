#!/bin/bash

export CURRENT_UID=$(id -u)
export CURRENT_GID=$(id -g)

DEBUG=""

while [[ $# -gt 0 ]]; do
  case $1 in
    --debug)
      export DEBUG=true
      shift
      ;;
    --clear)
      docker-compose down
      rm -rf data/db/* && touch data/db/.gitkeep
      rm -rf data/downloads/* && touch data/downloads/.gitkeep
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

# Initialize directories and permissions if they don't exist
mkdir -p data/downloads data/db data/cache
chmod 755 data/downloads data/db data/cache
chown -R $(id -u):$(id -g) data/

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