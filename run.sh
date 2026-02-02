#!/usr/bin/env bash

set -euo pipefail

SCRIPT_NAME=${0##*/}
ROOT_DIR="$(cd "$(dirname "$0")" && pwd)"
COMPOSE_FILE="$ROOT_DIR/docker/docker-compose.yml"
PROJECT_NAME="evm-event-indexer"

usage() {
  echo "usage: $SCRIPT_NAME [up|down]"
  echo "  $SCRIPT_NAME  up [-i|--infra] [-p|--purge]  start docker-compose services"
  echo "  $SCRIPT_NAME  down [-p|--purge]             stop docker-compose services"
  echo "options:"
  echo "  -i|--infra   start only infra (mysql, redis, anvil)"
  echo "  -p|--purge   remove volumes before running up services"
  echo
  exit 1
}

if [ $# -lt 1 ]; then
  usage
fi

command="$1"
shift

infra_only=false
purge=false
while [ $# -gt 0 ]; do
  case "$1" in
    -i|--infra)
      infra_only=true
      shift
      ;;
    -p|--purge)
      purge=true
      shift
      ;;
    -h|--help)
      usage
      ;;
    *)
      usage
      ;;
  esac
done

case "$command" in
  up)
    echo "starting docker-compose (file: $COMPOSE_FILE)..."
    if [ "$purge" = true ]; then
      docker compose -p "$PROJECT_NAME" -f "$COMPOSE_FILE" down -v
    else
      docker compose -p "$PROJECT_NAME" -f "$COMPOSE_FILE" down
    fi
    if [ "$infra_only" = true ]; then
      docker compose -p "$PROJECT_NAME" -f "$COMPOSE_FILE" up -d --build mysql redis anvil
    else
      docker compose -p "$PROJECT_NAME" -f "$COMPOSE_FILE" up -d --build
    fi
    echo "services started"
    ;;
  down)
    echo "stopping docker-compose (file: $COMPOSE_FILE)..."
    if [ "$purge" = true ]; then
    docker compose -p "$PROJECT_NAME" -f "$COMPOSE_FILE" down -v
    fi
    docker compose -p "$PROJECT_NAME" -f "$COMPOSE_FILE" down
    echo "services stopped"
    ;;
  -h|--help)
    usage
    ;;
  *)
    usage
    ;;
esac
