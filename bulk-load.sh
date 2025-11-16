#!/usr/bin/env bash

# Get the container ID of running dgraph/standalone:latest
CID=$(docker ps -q --filter "ancestor=dgraph/standalone:latest")

echo "$CID"

docker exec "$CID" rm -rf /dgraph/p/
docker exec "$CID" dgraph bulk -f /data/cskg_test.json -s /schema.dql --map_shards=4 --reduce_shards=2 --http localhost:8000 --zero=localhost:5080
docker exec "$CID" cp -r /tmp/out/0/p/ /dgraph/p/

# docker restart "$CID"