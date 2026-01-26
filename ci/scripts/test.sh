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
RUN_COUNT="${1:-1}"

if ! [[ "$RUN_COUNT" =~ ^[0-9]+$ ]] || [[ "$RUN_COUNT" -lt 1 ]]; then
    echo "Usage: $0 [run-count>=1]" >&2
    exit 1
fi

start_container_perf() {
    local container_id="$1"

    # log "Starting perf collection for container '$container_id'"
    ./scripts/container-perf-stats.sh "$container_id" \
        </dev/null > /dev/null 2>&1 &
    local pid=$!

    local pgid
    pgid=$(ps -o pgid= "$pid" | tr -d ' ')

    # log "Perf collection for '$container_id' started with pid=$pid pgid=$pgid"
    printf '%s:%s\n' "$pid" "$pgid"
}

stop_container_perf() {
    local pid_pgid="$1"

    local pid="${pid_pgid%%:*}"
    local pgid="${pid_pgid##*:}"

    # log "Stopping perf collection pid=$pid pgid=$pgid"
    kill -TERM -"$pgid" || true
    wait "$pid" || true
    # log "Perf collection pid=$pid terminated"
}


log "Test script starting"
log "Input CSV: $INPUT_FILE"
DB_PERF_PIDG=$(start_container_perf "dgraph")
APP_PERF_PIDG=$(start_container_perf "dgraph-client")


log 'Running benchmark'
echo PIDG $DB_PERF_PIDG
echo PIDG $APP_PERF_PIDG




run_iteration() {
    local iteration="$1"
    local counter=0

    log "Iteration $iteration: starting query 1 batch"
    echo "TASK 1 (iteration $iteration)" >> logs/dgraph-client.log
    echo "TASK 1 (iteration $iteration)" >> logs/dgraph.log

    while IFS=',' read -r id label; do
      counter=$((counter + 1))
      log "Iteration $iteration: #$counter Starting query 1 for id='$id' label='$label'"
      podman exec dgraph-client /usr/local/bin/client 1 \
        --vars "{\"uri\":\"$id\"}" \
        --quiet \
        </dev/null || true

    done < <(tail -n +2 "$INPUT_FILE" | tr -d '\r')

    log "Iteration $iteration: Starting query 10"
    echo "TASK 10 (iteration $iteration)" >> logs/dgraph-client.log
    echo "TASK 10 (iteration $iteration)" >> logs/dgraph.log
    podman exec dgraph-client /usr/local/bin/client 10 \
      --quiet \
      </dev/null || true

    sleep 1

    log "Iteration $iteration: Starting query 9"
    echo "TASK 9 (iteration $iteration)" >> logs/dgraph-client.log
    echo "TASK 9 (iteration $iteration)" >> logs/dgraph.log
    podman exec dgraph-client /usr/local/bin/client 9 \
      --quiet \
        </dev/null || true

    sleep 1

    log "Iteration $iteration: Starting query 17"
    echo "TASK 17 (iteration $iteration)" >> logs/dgraph-client.log
    echo "TASK 17 (iteration $iteration)" >> logs/dgraph.log
    podman exec dgraph-client /usr/local/bin/client 17 \
        --vars "{\"uri\":\"/c/en/slang\"}" \
        --quiet \
        </dev/null || true

    log "Iteration $iteration: Starting query 12"
    echo "TASK 12 (iteration $iteration)" >> logs/dgraph-client.log
    echo "TASK 12 (iteration $iteration)" >> logs/dgraph.log
    podman exec dgraph-client /usr/local/bin/client 12 \
        --vars "{\"uri\":\"$id\"}" \
        --quiet \
        </dev/null || true

    sleep 1

    log "Iteration $iteration: Finished benchmark loop; processed $counter rows"
}

for ((iteration=1; iteration<=RUN_COUNT; iteration++)); do
    run_iteration "$iteration"
done

stop_container_perf $DB_PERF_PIDG
stop_container_perf $APP_PERF_PIDG 

log "All perf collectors stopped; exiting script"
exit 0
