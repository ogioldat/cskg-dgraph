#!/bin/bash

set -euo pipefail


INPUT_FILE="data/sample-nodes.csv"

start_container_perf() {
    local container_id="$1"

    ./scripts/container-perf-stats.sh "$container_id" \
        </dev/null > /dev/null 2>&1 &
    local pid=$!

    local pgid
    pgid=$(ps -o pgid= "$pid" | tr -d ' ')

    printf '%s:%s\n' "$pid" "$pgid"
}

stop_container_perf() {
    local pid_pgid="$1"

    local pid="${pid_pgid%%:*}"
    local pgid="${pid_pgid##*:}"

    kill -TERM -"$pgid"
    wait "$pid"
}


DB_CONTAINER_ID=$(docker ps -q --filter name=dgraph-alpha)
DB_PERF_PIDG=$(start_container_perf "$DB_CONTAINER_ID")


echo 'Running benchmark'

tail -n +2 "$INPUT_FILE" | while IFS=',' read -r id label; do
    ./bin/client --query=1 --vars="{\"uri\":\"$id\"}" --quiet
done

echo 'Finished benchmark'

stop_container_perf $DB_PERF_PIDG
