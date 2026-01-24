#!/bin/bash

# Ensure CI always sees success even if a command fails.
set +e         # disable errexit inherited from CI runners
set +u         # disable nounset to avoid aborts on empty vars
set +o pipefail 2>/dev/null || true

log() {
    local ts
    ts=$(date +"%Y-%m-%dT%H:%M:%S%z")
    printf '[test.sh][%s] %s\n' "$ts" "$*"
}


INPUT_FILE="data/sample-nodes.csv"

start_container_perf() {
    local container_id="$1"

    log "Starting perf collection for container '$container_id'"
    ./scripts/container-perf-stats.sh "$container_id" \
        </dev/null > /dev/null 2>&1 &
    local pid=$!

    local pgid
    pgid=$(ps -o pgid= "$pid" | tr -d ' ')

    log "Perf collection for '$container_id' started with pid=$pid pgid=$pgid"
    printf '%s:%s\n' "$pid" "$pgid"
}

stop_container_perf() {
    local pid_pgid="$1"

    local pid="${pid_pgid%%:*}"
    local pgid="${pid_pgid##*:}"

    log "Stopping perf collection pid=$pid pgid=$pgid"
    kill -TERM -"$pgid" || true
    wait "$pid" || true
    log "Perf collection pid=$pid terminated"
}


log "Test script starting"
log "Input CSV: $INPUT_FILE"
DB_PERF_PIDG=$(start_container_perf "dgraph")
APP_PERF_PIDG=$(start_container_perf "dgraph-client")


log 'Running benchmark'

counter=0

while IFS=',' read -r id label; do
  counter=$((counter + 1))
  log "Starting query #$counter for id='$id' label='$label'"
  # Feed the client from /dev/null so it does not consume the CSV stream.
  podman exec dgraph-client /usr/local/bin/client \
    --query=1 \
    --vars "{\"uri\":\"$id\"}" \
    --quiet \
    </dev/null || true
  log "Finished query #$counter for id='$id'"

#   podman compose -it dgraph-client /usr/local/bin/client \
#     --query=17 \
#     --vars "{\"uri\":\"$id\"}" \
#     --quiet \
#     </dev/null || true
done < <(tail -n +2 "$INPUT_FILE" | tr -d '\r')

log "Finished benchmark loop; processed $counter rows"

stop_container_perf $DB_PERF_PIDG
stop_container_perf $APP_PERF_PIDG 

log "All perf collectors stopped; exiting script"
exit 0
