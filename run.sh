#!/usr/bin/env bash

set -euo pipefail

SCRIPT_NAME=${0##*/}
ROOT_DIR="$(cd "$(dirname "$0")" && pwd)"
COMPOSE_FILE="$ROOT_DIR/docker/docker-compose.yml"
PROJECT_NAME="evm-event-indexer"

usage() {
  echo "$SCRIPT_NAME - 控制本專案 docker-compose"
  echo
  echo "用法:"
  echo "  $SCRIPT_NAME up     啟動 docker-compose 服務"
  echo "  $SCRIPT_NAME down   關閉 docker-compose 服務"
  echo
  exit 1
}

if [ $# -lt 1 ]; then
  usage
fi

case "$1" in
  up)
    echo "啟動 docker-compose (檔案: $COMPOSE_FILE)..."
    docker compose -p "$PROJECT_NAME" -f "$COMPOSE_FILE" down -v
    docker compose -p "$PROJECT_NAME" -f "$COMPOSE_FILE" up -d --build
    echo "服務已啟動"
    ;;
  down)
    echo "關閉 docker-compose (檔案: $COMPOSE_FILE)..."
    docker compose -p "$PROJECT_NAME" -f "$COMPOSE_FILE" down -v
    echo "服務已關閉"
    ;;
  -h|--help)
    usage
    ;;
  *)
    usage
    ;;
esac

