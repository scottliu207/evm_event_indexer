#!/usr/bin/env bash
# the file name "00-run.sh" is only used to make sure the file is always executed first.
# This script adds a simple recursive runner, execute files in the following directories:
#   /docker-entrypoint-initdb.d/schema/*
#   /docker-entrypoint-initdb.d/init/*

if ! declare -F docker_process_init_files >/dev/null 2>&1; then
	echo "docker_process_init_files not found; run.sh should be sourced by MySQL docker-entrypoint.sh" >&2
	return 0 2>/dev/null || exit 0
fi

BASE_DIR="/docker-entrypoint-initdb.d"

run_dir() {
	local dir="$1"
	[ -d "$dir" ] || return 0

	local -a files=()
	while IFS= read -r f; do
		[ -n "$f" ] && files+=("$f")
	done < <(
		find "$dir" -type f \( \
			-name '*.sh' -o \
			-name '*.sql' -o \
			-name '*.sql.gz' -o \
			-name '*.sql.xz' -o \
			-name '*.sql.zst' \
		\) | LC_ALL=C sort
	)

	((${#files[@]})) || return 0
	docker_process_init_files "${files[@]}"
}

run_dir "$BASE_DIR/schema"
run_dir "$BASE_DIR/init"
