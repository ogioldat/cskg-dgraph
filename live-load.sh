#!/usr/bin/env bash

# Get the container ID of running dgraph/standalone:latest
CID=$(docker ps -q --filter "ancestor=dgraph/dgraph:latest" | head -n 1)
echo "$CID"

# for i in data/out/*/; do
#   i=$(basename "$i")
#   echo "Running import for folder: $i"

#   docker exec "$CID" dgraph live \
#     -f "/data/out/$i" \
#     -s "/@schema.gql" \
#     --batch 200 \
#     --bufferSize 50 \
#     --conc 10
# done

docker exec "$CID" dgraph live \
    -f "/data/out/7" \
    -s "/@schema.dql" \
    --batch 200 \
    --bufferSize 50 \
    --conc 10