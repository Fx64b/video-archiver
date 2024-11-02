#!/bin/bash


rm -rf data/downloads/* && touch data/downloads/.gitkeep

if [ "$1" == "--clear" ]; then
  docker-compose down
  rm -rf data/db/* && touch data/db/.gitkeep
fi

if [ "$1" == "--help" ] || [ "$1" == "-h" ]; then
  echo "This script is used to run the application in a docker container."
  echo "If you want to run the application and clear the database, use the --clear flag."
  echo "If you want to run the application without clearing the database, simply run the script without any flags."
  exit 0
fi

clear

echo "./run.sh --help for more information."

docker-compose up --build && docker-compose down