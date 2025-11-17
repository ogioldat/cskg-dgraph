#!/usr/bin/env bash

src="data/out"
batch_size=100
i=0
batch=1

mkdir -p "$src"

for f in "$src"/*.json; do
    folder="$src/batch_$batch"
    mkdir -p "$folder"
    mv "$f" "$folder"/
    i=$((i+1))
    if [ $i -ge $batch_size ]; then
        i=0
        batch=$((batch+1))
    fi
done