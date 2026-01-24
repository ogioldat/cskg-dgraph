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

    kill -TERM -"$pgid" || true
    wait "$pid" || true
}


DB_PERF_PIDG=$(start_container_perf "dgraph")
APP_PERF_PIDG=$(start_container_perf "dgraph-client")


echo 'Running benchmark'

while IFS=',' read -r id label; do
  podman exec -it dgraph-client /usr/local/bin/client \
    --query=1 \
    --vars "{\"uri\":\"$id\"}"
    # --quiet \
    # </dev/null || true

#   podman compose -it dgraph-client /usr/local/bin/client \
#     --query=17 \
#     --vars "{\"uri\":\"$id\"}" \
#     --quiet \
#     </dev/null || true
done < <(tail -n +2 "$INPUT_FILE" | tr -d '\r')

echo 'Finished benchmark'

stop_container_perf $DB_PERF_PIDG
stop_container_perf $APP_PERF_PIDG 

exit 0
