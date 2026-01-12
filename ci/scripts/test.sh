#!/bin/bash

set -euo pipefail


INPUT_FILE="data/sample-nodes.csv"

start_container_perf() {
    local container_id="$1"

    ./ci/scripts/container-perf-stats.sh "$container_id" \
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


# DB_PERF_PIDG=$(start_container_perf "dgraph")
# APP_PERF_PIDG=$(start_container_perf "app-ci")


echo 'Running benchmark'

 tail -n +2 "$INPUT_FILE" | while IFS=',' read -r id label; do
      docker run --rm \
        --network host \
        -w /app \
        --entrypoint /usr/local/bin/client \
        ci-app \
        --query=1 \
        --quiet \
        --vars='{"uri":"'"$id"'"}' \
        </dev/null
  done

echo 'Finished benchmark'

# stop_container_perf $DB_PERF_PIDG
# stop_container_perf $APP_PERF_PIDG
