#!/bin/bash
set -euo pipefail

if [[ $# -lt 1 || $# -gt 2 ]]; then
  echo "Usage: $0 <container-id|name> [output-file]" >&2
  exit 1
fi

container_id_or_name=$1

echo $container_id_or_name

if [[ $# -eq 2 ]]; then
  output_file=$2
else
  if container_name=$(podman inspect --format '{{.Name}}' "$container_id_or_name" 2>/dev/null); then
    container_name=${container_name#/}
  else
    container_name=$container_id_or_name
  fi
  output_file="./logs/${container_name}.log"
fi

output_dir=$(dirname "$output_file")
mkdir -p "$output_dir"

header="timestamp,container_id,name,cpu_percent,mem_usage,mem_limit,mem_percent,net_io_rx,net_io_tx,block_io_read,block_io_write,pids"
ensure_header() {
  if [[ ! -f "$output_file" || ! -s "$output_file" ]]; then
    echo "$header">>"$output_file"
    return
  fi

  if ! head -n 1 "$output_file" | grep -Fxq "$header"; then
    tmp_file=$(mktemp)
    {
      echo "$header"
      cat "$output_file"
    } >"$tmp_file"
    mv "$tmp_file" "$output_file"
  fi
}

ensure_header

format='{{.Name}},{{.CPUPerc}},{{.MemUsage}},{{.MemPerc}},{{.NetIO}},{{.BlockIO}},{{.PIDs}}'

echo running stats for $container_id_or_name
podman stats "$container_id_or_name" --format "$format" |
while IFS= read -r stats_line; do
  timestamp=$(date -Iseconds)

  IFS=',' read -r name cpu_percent mem_usage_raw mem_percent net_io_raw block_io_raw pids <<<"$stats_line"

  trim_split() {
    local raw_value="$1"
    local part_index="$2"
    echo "$raw_value" | awk -F'/' -v idx="$part_index" '{gsub(/^[ \t]+|[ \t]+$/, "", $idx); print $idx}' | sed 's/\x1b\[[0-9;]*[A-Za-z]//g'
  }

  name=$(trim_split "$name" 1)
  mem_usage=$(trim_split "$mem_usage_raw" 1)
  mem_limit=$(trim_split "$mem_usage_raw" 2)
  net_io_rx=$(trim_split "$net_io_raw" 1)
  net_io_tx=$(trim_split "$net_io_raw" 2)
  block_io_read=$(trim_split "$block_io_raw" 1)
  block_io_write=$(trim_split "$block_io_raw" 2)

  echo "$timestamp,$name,$cpu_percent,$mem_usage,$mem_limit,$mem_percent,$net_io_rx,$net_io_tx,$block_io_read,$block_io_write,$pids" >>"$output_file"
done
