#!/usr/bin/env bash

set -euo pipefail

SCRIPT_NAME=${0##*/}
ROOT_DIR="$(cd "$(dirname "$0")" && pwd)"
COMPOSE_FILE="$ROOT_DIR/docker/docker-compose.yml"
PROJECT_NAME="evm-event-indexer"

usage() {
  echo "usage: $SCRIPT_NAME [up|down]"
  echo "  $SCRIPT_NAME up      start docker-compose services"
  echo "  $SCRIPT_NAME down    stop docker-compose services"
  echo
  exit 1
}

if [ $# -lt 1 ]; then
  usage
fi

case "$1" in
  up)
    echo "starting docker-compose (file: $COMPOSE_FILE)..."
    docker compose -p "$PROJECT_NAME" -f "$COMPOSE_FILE" down -v
    docker compose -p "$PROJECT_NAME" -f "$COMPOSE_FILE" up -d --build
    echo "services started"
    ;;
  down)
    echo "stopping docker-compose (file: $COMPOSE_FILE)..."
    docker compose -p "$PROJECT_NAME" -f "$COMPOSE_FILE" down -v
    echo "services stopped"
    ;;
  -h|--help)
    usage
    ;;
  *)
    usage
    ;;
esac

